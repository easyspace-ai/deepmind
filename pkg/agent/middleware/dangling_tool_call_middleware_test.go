package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestDanglingToolCallMiddleware_Name(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()

	if m.Name() != "dangling_tool_call" {
		t.Errorf("Name() = %v, want 'dangling_tool_call'", m.Name())
	}
}

func TestDanglingToolCallMiddleware_buildPatchedMessages_NoDangling(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ts := state.NewThreadState()

	// 有工具调用，也有对应的响应
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "call-1", Function: schema.FunctionCall{Name: "test"}},
		}},
		{Role: schema.Tool, ToolCallID: "call-1", Content: "result"},
	}

	patchedState, needsPatch := m.buildPatchedMessages(ts)

	if needsPatch {
		t.Error("needsPatch = true, want false")
	}
	if patchedState != ts {
		t.Error("patchedState should be the same as input")
	}
}

func TestDanglingToolCallMiddleware_buildPatchedMessages_WithDangling(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ts := state.NewThreadState()

	// 有工具调用，但没有对应的响应
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "call-1", Function: schema.FunctionCall{Name: "test"}},
		}},
	}

	patchedState, needsPatch := m.buildPatchedMessages(ts)

	if !needsPatch {
		t.Error("needsPatch = false, want true")
	}
	if len(patchedState.Messages) != 2 {
		t.Errorf("len(patchedState.Messages) = %v, want 2", len(patchedState.Messages))
	}

	// 验证补丁消息
	patchedMsg := patchedState.Messages[1]
	if patchedMsg.Role != schema.Tool {
		t.Errorf("patchedMsg.Role = %v, want Tool", patchedMsg.Role)
	}
	if patchedMsg.ToolCallID != "call-1" {
		t.Errorf("patchedMsg.ToolCallID = %v, want 'call-1'", patchedMsg.ToolCallID)
	}
	if patchedMsg.ToolName != "test" {
		t.Errorf("patchedMsg.ToolName = %v, want 'test'", patchedMsg.ToolName)
	}
}

func TestDanglingToolCallMiddleware_buildPatchedMessages_MultipleDangling(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ts := state.NewThreadState()

	// 多个工具调用，部分有响应，部分没有
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "call-1", Function: schema.FunctionCall{Name: "test1"}},
			{ID: "call-2", Function: schema.FunctionCall{Name: "test2"}},
			{ID: "call-3", Function: schema.FunctionCall{Name: "test3"}},
		}},
		{Role: schema.Tool, ToolCallID: "call-2", Content: "result2"},
	}

	patchedState, needsPatch := m.buildPatchedMessages(ts)

	if !needsPatch {
		t.Error("needsPatch = false, want true")
	}
	if len(patchedState.Messages) != 4 {
		t.Errorf("len(patchedState.Messages) = %v, want 4", len(patchedState.Messages))
	}

	// 验证补丁消息的位置（在原 Assistant 消息之后）
	patchedIDs := make(map[string]bool)
	for _, msg := range patchedState.Messages {
		if msg.Role == schema.Tool && msg.ToolCallID != "call-2" {
			patchedIDs[msg.ToolCallID] = true
		}
	}
	if !patchedIDs["call-1"] {
		t.Error("call-1 should be patched")
	}
	if !patchedIDs["call-3"] {
		t.Error("call-3 should be patched")
	}
}

func TestDanglingToolCallMiddleware_buildPatchedMessages_MultipleAssistantMessages(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ts := state.NewThreadState()

	// 多条 Assistant 消息，都有挂起工具调用
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "call-1", Function: schema.FunctionCall{Name: "test1"}},
		}},
		{Role: schema.User, Content: "continue"},
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "call-2", Function: schema.FunctionCall{Name: "test2"}},
		}},
	}

	patchedState, needsPatch := m.buildPatchedMessages(ts)

	if !needsPatch {
		t.Error("needsPatch = false, want true")
	}
	if len(patchedState.Messages) != 5 {
		t.Errorf("len(patchedState.Messages) = %v, want 5", len(patchedState.Messages))
	}
}

func TestDanglingToolCallMiddleware_buildPatchedMessages_AlreadyPatched(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ts := state.NewThreadState()

	// 已经有补丁消息了
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "call-1", Function: schema.FunctionCall{Name: "test"}},
		}},
		{Role: schema.Tool, ToolCallID: "call-1", Content: "[Tool call was interrupted and did not return a result.]"},
	}

	patchedState, needsPatch := m.buildPatchedMessages(ts)

	if needsPatch {
		t.Error("needsPatch = true, want false")
	}
	if patchedState != ts {
		t.Error("patchedState should be the same as input")
	}
}

func TestDanglingToolCallMiddleware_buildPatchedMessages_NoToolCalls(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ts := state.NewThreadState()

	// 没有工具调用
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
		{Role: schema.Assistant, Content: "hi"},
	}

	patchedState, needsPatch := m.buildPatchedMessages(ts)

	if needsPatch {
		t.Error("needsPatch = true, want false")
	}
	if patchedState != ts {
		t.Error("patchedState should be the same as input")
	}
}

func TestDanglingToolCallMiddleware_buildPatchedMessages_EmptyToolID(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ts := state.NewThreadState()

	// 工具调用没有 ID
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "", Function: schema.FunctionCall{Name: "test"}},
		}},
	}

	patchedState, needsPatch := m.buildPatchedMessages(ts)

	if needsPatch {
		t.Error("needsPatch = true, want false (empty tool ID should not be patched)")
	}
	if patchedState != ts {
		t.Error("patchedState should be the same as input")
	}
}

func TestDanglingToolCallMiddleware_BeforeModel_NoPatchNeeded(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
	}

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestDanglingToolCallMiddleware_BeforeModel_WithPatch(t *testing.T) {
	m := NewDefaultDanglingToolCallMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "call-1", Function: schema.FunctionCall{Name: "test"}},
		}},
	}

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Fatal("BeforeModel() stateUpdate = nil, want non-nil")
	}

	// 验证 stateUpdate
	messages, ok := stateUpdate["messages"]
	if !ok {
		t.Error("stateUpdate should contain 'messages'")
	}
	if messages == nil {
		t.Error("messages in stateUpdate should not be nil")
	}

	// 验证原 state 已被更新
	if len(ts.Messages) != 2 {
		t.Errorf("len(ts.Messages) = %v, want 2", len(ts.Messages))
	}
}

func TestDanglingToolCallMiddleware_IntegrationWithChain(t *testing.T) {
	// 验证能正确加入中间件链
	config := DefaultMiddlewareConfig()
	chain := BuildLeadAgentMiddlewares(config)

	// 验证链中有 DanglingToolCallMiddleware
	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "dangling_tool_call" {
			found = true
			break
		}
	}
	if !found {
		t.Error("DanglingToolCallMiddleware not found in chain")
	}
}

func TestDanglingToolCallMiddleware_Logger(t *testing.T) {
	// 验证 logger 参数可以为 nil
	customLogger := zap.NewExample()
	m := NewDanglingToolCallMiddleware(customLogger)

	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}

	// nil logger 应该使用 NopLogger
	m2 := NewDanglingToolCallMiddleware(nil)
	if m2.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}
}

// BenchmarkDanglingToolCallMiddleware_buildPatchedMessages 性能基准测试
func BenchmarkDanglingToolCallMiddleware_buildPatchedMessages(b *testing.B) {
	m := NewDefaultDanglingToolCallMiddleware()

	ts := state.NewThreadState()
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "call-1", Function: schema.FunctionCall{Name: "test"}},
		}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.buildPatchedMessages(ts)
	}
}
