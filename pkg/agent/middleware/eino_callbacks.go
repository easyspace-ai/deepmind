package middleware

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	ubc "github.com/cloudwego/eino/utils/callbacks"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// uploadsWithFilesMiddleware 与 UploadsMiddleware 四参数 BeforeAgent 对齐（可选文件列表）。
type uploadsWithFilesMiddleware interface {
	Middleware
	BeforeAgent(ctx context.Context, ts *state.ThreadState, threadID string, additionalFiles []FileInfo) (map[string]interface{}, error)
}

// runBeforeAgentPhase 按链顺序执行带 threadID 的 BeforeAgent（含 Uploads 四参变体）。
func runBeforeAgentPhase(ctx context.Context, chain *MiddlewareChain, ts *state.ThreadState, threadID string) error {
	if chain == nil || ts == nil {
		return nil
	}
	for _, mw := range chain.Middlewares() {
		if mw == nil {
			continue
		}
		if u, ok := mw.(uploadsWithFilesMiddleware); ok {
			upd, err := u.BeforeAgent(ctx, ts, threadID, nil)
			if err != nil {
				return err
			}
			state.ApplyMiddlewareUpdates(ts, upd)
			continue
		}
		if u, ok := mw.(BeforeAgentWithIDMiddleware); ok {
			upd, err := u.BeforeAgent(ctx, ts, threadID)
			if err != nil {
				return err
			}
			state.ApplyMiddlewareUpdates(ts, upd)
			continue
		}
		if u, ok := mw.(BeforeAgentMiddleware); ok {
			upd, err := u.BeforeAgent(ctx, ts)
			if err != nil {
				return err
			}
			state.ApplyMiddlewareUpdates(ts, upd)
		}
	}
	return nil
}

func runBeforeModelPhase(ctx context.Context, chain *MiddlewareChain, ts *state.ThreadState) error {
	if chain == nil || ts == nil {
		return nil
	}
	for _, mw := range chain.Middlewares() {
		if u, ok := mw.(BeforeModelMiddleware); ok {
			upd, err := u.BeforeModel(ctx, ts)
			if err != nil {
				return err
			}
			state.ApplyMiddlewareUpdates(ts, upd)
		}
	}
	return nil
}

func runAfterModelPhase(ctx context.Context, chain *MiddlewareChain, ts *state.ThreadState) error {
	if chain == nil || ts == nil {
		return nil
	}
	for _, mw := range chain.Middlewares() {
		if u, ok := mw.(AfterModelMiddleware); ok {
			upd, err := u.AfterModel(ctx, ts)
			if err != nil {
				return err
			}
			state.ApplyMiddlewareUpdates(ts, upd)
		}
	}
	return nil
}

func findLoopDetection(chain *MiddlewareChain) *LoopDetectionMiddleware {
	if chain == nil {
		return nil
	}
	for _, mw := range chain.Middlewares() {
		if ld, ok := mw.(*LoopDetectionMiddleware); ok {
			return ld
		}
	}
	return nil
}

func toolCallsToLoopToolCalls(tcs []schema.ToolCall) []ToolCall {
	out := make([]ToolCall, 0, len(tcs))
	for _, tc := range tcs {
		var args map[string]interface{}
		if tc.Function.Arguments != "" {
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
		}
		if args == nil {
			args = map[string]interface{}{}
		}
		out = append(out, ToolCall{
			Name: tc.Function.Name,
			Args: args,
			ID:   tc.ID,
		})
	}
	return out
}

// toolCallIDFromToolCallback 解析当前工具调用的 tool_call_id（compose 注入或 CallbackInput.Extra）。
func toolCallIDFromToolCallback(ctx context.Context, tin *tool.CallbackInput) string {
	if id := compose.GetToolCallID(ctx); id != "" {
		return id
	}
	if tin == nil || tin.Extra == nil {
		return ""
	}
	for _, k := range []string{"tool_call_id", "toolCallID", "ToolCallID"} {
		if v, ok := tin.Extra[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// composeWrapToolCallChain 将链上所有 WrapToolCallMiddleware 按 DeerFlow 顺序组合为洋葱模型：
// 链中越靠后的中间件越靠近外层（越先拦截），与 chain.go 中「Clarification 必须最后」一致。
// injectPendingToolErrorAsToolMessage 将 OnError 阶段记录的友好错误注入为 Tool 角色消息，便于模型继续对话。
func injectPendingToolErrorAsToolMessage(ts *state.ThreadState, input *model.CallbackInput) {
	if ts == nil || input == nil {
		return
	}
	pending := strings.TrimSpace(ts.PendingToolErrorForModel)
	if pending == "" || ts.LastToolError == nil {
		return
	}
	tcid := strings.TrimSpace(ts.LastToolError.ToolCallID)
	if tcid == "" {
		tcid = "unknown-tool-call"
	}
	tname := strings.TrimSpace(ts.LastToolError.ToolName)
	var opts []schema.ToolMessageOption
	if tname != "" {
		opts = append(opts, schema.WithToolName(tname))
	}
	tm := schema.ToolMessage(pending, tcid, opts...)
	ts.Messages = append(ts.Messages, tm)
	input.Messages = ts.Messages
	ts.PendingToolErrorForModel = ""
}

func composeWrapToolCallChain(chain *MiddlewareChain, inner ToolCallHandler) ToolCallHandler {
	if chain == nil || inner == nil {
		return inner
	}
	var wrappers []WrapToolCallMiddleware
	for _, mw := range chain.Middlewares() {
		if mw == nil {
			continue
		}
		if w, ok := mw.(WrapToolCallMiddleware); ok {
			wrappers = append(wrappers, w)
		}
	}
	next := inner
	for i := len(wrappers) - 1; i >= 0; i-- {
		w := wrappers[i]
		prev := next
		next = func(ctx context.Context, tc *ToolCallInfo) (*ToolCallResult, error) {
			return w.WrapToolCall(ctx, tc, prev)
		}
	}
	return next
}

// NewDeerflowEinoHandler 构建 DeerFlow 中间件链对应的 Eino callbacks.Handler。
// 调用方须在 ctx 中使用 WithThreadID 注入 thread_id；ts 为当前线程的 ThreadState（与图外状态共享指针）。
// log 可为 nil（使用 Nop）。
func NewDeerflowEinoHandler(chain *MiddlewareChain, ts *state.ThreadState, log *zap.Logger) callbacks.Handler {
	if log == nil {
		log = zap.NewNop()
	}
	modelH := &ubc.ModelCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *model.CallbackInput) context.Context {
			if ts == nil || input == nil {
				return ctx
			}
			tid := ThreadIDFromContext(ctx)
			if len(input.Messages) > 0 {
				ts.Messages = input.Messages
			}
			injectPendingToolErrorAsToolMessage(ts, input)
			if err := runBeforeAgentPhase(ctx, chain, ts, tid); err != nil {
				log.Warn("deerflow middleware BeforeAgent phase", zap.Error(err))
			}
			if err := runBeforeModelPhase(ctx, chain, ts); err != nil {
				log.Warn("deerflow middleware BeforeModel phase", zap.Error(err))
			}
			input.Messages = ts.Messages
			return ctx
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			if ts == nil {
				return ctx
			}
			tid := ThreadIDFromContext(ctx)
			if output != nil && output.Message != nil {
				ts.Messages = append(ts.Messages, output.Message)
				if ld := findLoopDetection(chain); ld != nil && len(output.Message.ToolCalls) > 0 {
					warn, hard := ld.TrackAndCheck(tid, toolCallsToLoopToolCalls(output.Message.ToolCalls))
					if warn != "" {
						output.Message.Content = warn + "\n\n" + output.Message.Content
					}
					if hard {
						output.Message.ToolCalls = nil
					}
				}
			}
			if err := runAfterModelPhase(ctx, chain, ts); err != nil {
				log.Warn("deerflow middleware AfterModel phase", zap.Error(err))
			}
			return ctx
		},
	}
	toolH := &ubc.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			return WithToolCallbackInput(ctx, input)
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			if ts == nil || info == nil || chain == nil || output == nil {
				return ctx
			}
			tin := toolCallbackInputFromContext(ctx)
			toolName := strings.TrimSpace(info.Name)
			args := map[string]interface{}{}
			if tin != nil && tin.ArgumentsInJSON != "" {
				_ = json.Unmarshal([]byte(tin.ArgumentsInJSON), &args)
			}
			tci := &ToolCallInfo{
				ID:   toolCallIDFromToolCallback(ctx, tin),
				Name: toolName,
				Args: args,
			}
			inner := func(context.Context, *ToolCallInfo) (*ToolCallResult, error) {
				res := output.Response
				if output.ToolOutput != nil && len(output.ToolOutput.Parts) > 0 {
					if output.ToolOutput.Parts[0].Type == schema.ToolPartTypeText {
						res = output.ToolOutput.Parts[0].Text
					}
				}
				return &ToolCallResult{Content: res}, nil
			}
			composed := composeWrapToolCallChain(chain, inner)
			res, err := composed(ctx, tci)
			if err != nil {
				log.Warn("deerflow tool middleware chain", zap.Error(err))
				return ctx
			}
			if res != nil {
				if res.Content != "" {
					output.Response = res.Content
					// 文本覆盖时丢弃结构化片段，避免与模型可见内容不一致
					if output.ToolOutput != nil {
						output.ToolOutput = nil
					}
				}
				if len(res.StateUpdate) > 0 {
					state.ApplyMiddlewareUpdates(ts, res.StateUpdate)
				}
				if res.Interrupt {
					log.Info("deerflow tool interrupt", zap.String("tool", toolName), zap.String("tool_call_id", tci.ID))
				}
			}
			return ctx
		},
		OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if ts == nil || info == nil || chain == nil || err == nil {
				return ctx
			}
			tin := toolCallbackInputFromContext(ctx)
			toolName := strings.TrimSpace(info.Name)
			args := map[string]interface{}{}
			if tin != nil && tin.ArgumentsInJSON != "" {
				_ = json.Unmarshal([]byte(tin.ArgumentsInJSON), &args)
			}
			tci := &ToolCallInfo{
				ID:   toolCallIDFromToolCallback(ctx, tin),
				Name: toolName,
				Args: args,
			}
			inner := func(context.Context, *ToolCallInfo) (*ToolCallResult, error) {
				return nil, err
			}
			composed := composeWrapToolCallChain(chain, inner)
			res, wrapErr := composed(ctx, tci)
			if wrapErr != nil {
				log.Warn("deerflow tool error middleware chain", zap.Error(wrapErr))
				return ctx
			}
			if res != nil {
				if len(res.StateUpdate) > 0 {
					state.ApplyMiddlewareUpdates(ts, res.StateUpdate)
				}
				if res.Content != "" {
					ts.PendingToolErrorForModel = res.Content
				}
			}
			return ctx
		},
	}
	return ubc.NewHandlerHelper().ChatModel(modelH).Tool(toolH).Handler()
}

// RunBeforeAgentPhase 执行链上所有 BeforeAgent / BeforeAgent(threadID) / Uploads 四参 BeforeAgent。
func RunBeforeAgentPhase(ctx context.Context, chain *MiddlewareChain, ts *state.ThreadState, threadID string) error {
	return runBeforeAgentPhase(ctx, chain, ts, threadID)
}

// RunBeforeModelPhase 执行链上所有 BeforeModel。
func RunBeforeModelPhase(ctx context.Context, chain *MiddlewareChain, ts *state.ThreadState) error {
	return runBeforeModelPhase(ctx, chain, ts)
}

// RunAfterModelPhase 执行链上所有 AfterModel。
func RunAfterModelPhase(ctx context.Context, chain *MiddlewareChain, ts *state.ThreadState) error {
	return runAfterModelPhase(ctx, chain, ts)
}
