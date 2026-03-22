package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/require"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestNewDeerflowEinoHandler_ThreadDataAndSandbox(t *testing.T) {
	cfg := DefaultMiddlewareConfig()
	cfg.BaseDir = t.TempDir()
	chain := BuildLeadAgentMiddlewares(cfg)
	ts := state.NewThreadState()

	h := NewDeerflowEinoHandler(chain, ts, zap.NewNop())
	require.NotNil(t, h)

	ctx := WithThreadID(context.Background(), "thread-eino-1")
	mi := &model.CallbackInput{
		Messages: []*schema.Message{
			{Role: schema.User, Content: "hi"},
		},
	}
	ctx = h.OnStart(ctx, &callbacks.RunInfo{Name: "cm", Type: "chat", Component: components.ComponentOfChatModel}, mi)

	require.NotNil(t, ts.ThreadData)
	require.NotEmpty(t, ts.ThreadData.WorkspacePath)
	require.NotNil(t, ts.Sandbox)
	require.Equal(t, "local", ts.Sandbox.SandboxID)
}

func TestNewDeerflowEinoHandler_DanglingPatchBeforeModel(t *testing.T) {
	chain := NewMiddlewareChain()
	chain.Add(NewDanglingToolCallMiddleware(zap.NewNop()))
	ts := state.NewThreadState()
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, Content: "", ToolCalls: []schema.ToolCall{
			{ID: "call-1", Type: "function", Function: schema.FunctionCall{Name: "bash", Arguments: `{}`}},
		}},
	}

	h := NewDeerflowEinoHandler(chain, ts, zap.NewNop())
	ctx := WithThreadID(context.Background(), "t2")
	mi := &model.CallbackInput{Messages: ts.Messages}
	_ = h.OnStart(ctx, &callbacks.RunInfo{Name: "cm", Component: components.ComponentOfChatModel}, mi)

	var toolMsgs int
	for _, m := range ts.Messages {
		if m.Role == schema.Tool {
			toolMsgs++
		}
	}
	require.GreaterOrEqual(t, toolMsgs, 1)
}

func TestNewDeerflowEinoHandler_LoopDetectionInject(t *testing.T) {
	cfg := DefaultMiddlewareConfig()
	cfg.LoopWarnThreshold = 2
	cfg.LoopHardLimit = 5
	chain := NewMiddlewareChain()
	chain.Add(NewLoopDetectionMiddlewareWithConfig(2, 5, DefaultWindowSize, DefaultMaxTrackedThreads, zap.NewNop()))
	ts := state.NewThreadState()

	h := NewDeerflowEinoHandler(chain, ts, zap.NewNop())
	ctx := WithThreadID(context.Background(), "t-loop")

	tcs := []schema.ToolCall{
		{ID: "a", Type: "function", Function: schema.FunctionCall{Name: "x", Arguments: `{}`}},
	}
	out := &model.CallbackOutput{
		Message: &schema.Message{
			Role:       schema.Assistant,
			Content:    "ok",
			ToolCalls:  tcs,
		},
	}
	for i := 0; i < 2; i++ {
		_ = h.OnEnd(ctx, &callbacks.RunInfo{Name: "cm", Component: components.ComponentOfChatModel}, out)
	}
	require.Contains(t, out.Message.Content, "LOOP DETECTED")
}

func TestNewDeerflowEinoHandler_ToolOnEnd_AskClarification(t *testing.T) {
	chain := NewMiddlewareChain()
	chain.Add(NewToolErrorHandlingMiddleware(zap.NewNop()))
	chain.Add(NewClarificationMiddleware())
	ts := state.NewThreadState()
	h := NewDeerflowEinoHandler(chain, ts, zap.NewNop())

	ctx := WithThreadID(context.Background(), "t-clar")
	ctx = h.OnStart(ctx, &callbacks.RunInfo{Name: "ask_clarification", Component: components.ComponentOfTool}, &tool.CallbackInput{
		ArgumentsInJSON: `{"question":"Which API?","clarification_type":"approach_choice","context":"Need your pick"}`,
	})
	out := &tool.CallbackOutput{Response: "SHOULD_BE_REPLACED"}
	ctx = h.OnEnd(ctx, &callbacks.RunInfo{Name: "ask_clarification", Component: components.ComponentOfTool}, out)

	require.NotEqual(t, "SHOULD_BE_REPLACED", out.Response)
	require.Contains(t, out.Response, "Which API?")
	require.True(t, ts.ClarificationPending)
	require.NotEmpty(t, ts.LastClarificationMessage)
	_ = ctx
}

func TestNewDeerflowEinoHandler_ToolOnError_RecordsLastToolError(t *testing.T) {
	chain := NewMiddlewareChain()
	chain.Add(NewToolErrorHandlingMiddleware(zap.NewNop()))
	chain.Add(NewClarificationMiddleware())
	ts := state.NewThreadState()
	h := NewDeerflowEinoHandler(chain, ts, zap.NewNop())

	ctx := WithThreadID(context.Background(), "t-err")
	ctx = h.OnStart(ctx, &callbacks.RunInfo{Name: "bash", Component: components.ComponentOfTool}, &tool.CallbackInput{
		ArgumentsInJSON: `{"command":"ls"}`,
		Extra:             map[string]any{"tool_call_id": "tc-9"},
	})
	_ = h.OnError(ctx, &callbacks.RunInfo{Name: "bash", Component: components.ComponentOfTool}, errors.New("sandbox down"))

	require.NotNil(t, ts.LastToolError)
	require.Equal(t, "bash", ts.LastToolError.ToolName)
	require.Equal(t, "tc-9", ts.LastToolError.ToolCallID)
	require.Contains(t, ts.LastToolError.Error, "sandbox down")
}

func TestNewDeerflowEinoHandler_ToolOnError_ThenModelInjectsToolMessage(t *testing.T) {
	chain := NewMiddlewareChain()
	chain.Add(NewToolErrorHandlingMiddleware(zap.NewNop()))
	chain.Add(NewClarificationMiddleware())
	ts := state.NewThreadState()
	h := NewDeerflowEinoHandler(chain, ts, zap.NewNop())

	ctx := WithThreadID(context.Background(), "t-inj")
	ctx = h.OnStart(ctx, &callbacks.RunInfo{Name: "bash", Component: components.ComponentOfTool}, &tool.CallbackInput{
		ArgumentsInJSON: `{}`,
		Extra:             map[string]any{"tool_call_id": "call-err-1"},
	})
	_ = h.OnError(ctx, &callbacks.RunInfo{Name: "bash", Component: components.ComponentOfTool}, errors.New("disk full"))

	require.NotEmpty(t, ts.PendingToolErrorForModel)
	require.NotNil(t, ts.LastToolError)

	mi := &model.CallbackInput{Messages: []*schema.Message{schema.UserMessage("hi")}}
	ctx2 := WithThreadID(context.Background(), "t-inj")
	_ = h.OnStart(ctx2, &callbacks.RunInfo{Name: "cm", Component: components.ComponentOfChatModel}, mi)

	require.Empty(t, ts.PendingToolErrorForModel)
	var toolMsgs int
	for _, m := range ts.Messages {
		if m.Role == schema.Tool && m.ToolCallID == "call-err-1" {
			toolMsgs++
			require.Contains(t, m.Content, "bash")
		}
	}
	require.Equal(t, 1, toolMsgs)
}
