package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/require"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestSubagentLimitMiddleware_Name(t *testing.T) {
	m := NewDefaultSubagentLimitMiddleware()

	if m.Name() != "subagent_limit" {
		t.Errorf("Name() = %v, want 'subagent_limit'", m.Name())
	}
}

func TestSubagentLimitMiddleware_NewDefault(t *testing.T) {
	m := NewDefaultSubagentLimitMiddleware()

	if m.maxConcurrentSubagents != 3 {
		t.Errorf("maxConcurrentSubagents = %v, want 3", m.maxConcurrentSubagents)
	}
}

func TestSubagentLimitMiddleware_NewWithCustomLimit(t *testing.T) {
	m := NewSubagentLimitMiddleware(5, nil)

	if m.maxConcurrentSubagents != 5 {
		t.Errorf("maxConcurrentSubagents = %v, want 5", m.maxConcurrentSubagents)
	}
}

func TestSubagentLimitMiddleware_NewWithZeroLimit(t *testing.T) {
	m := NewSubagentLimitMiddleware(0, nil)

	if m.maxConcurrentSubagents != 3 {
		t.Errorf("maxConcurrentSubagents = %v, want 3 (default)", m.maxConcurrentSubagents)
	}
}

func TestSubagentLimitMiddleware_NewWithNegativeLimit(t *testing.T) {
	m := NewSubagentLimitMiddleware(-1, nil)

	if m.maxConcurrentSubagents != 3 {
		t.Errorf("maxConcurrentSubagents = %v, want 3 (default)", m.maxConcurrentSubagents)
	}
}

func TestSubagentLimitMiddleware_truncateTaskToolCalls_UnderLimit(t *testing.T) {
	m := NewSubagentLimitMiddleware(3, nil)

	toolCalls := []schema.ToolCall{
		{ID: "1", Function: schema.FunctionCall{Name: "task"}},
		{ID: "2", Function: schema.FunctionCall{Name: "task"}},
	}

	result := m.truncateTaskToolCalls(toolCalls)

	if len(result) != 2 {
		t.Errorf("len(result) = %v, want 2", len(result))
	}
}

func TestSubagentLimitMiddleware_truncateTaskToolCalls_AtLimit(t *testing.T) {
	m := NewSubagentLimitMiddleware(3, nil)

	toolCalls := []schema.ToolCall{
		{ID: "1", Function: schema.FunctionCall{Name: "task"}},
		{ID: "2", Function: schema.FunctionCall{Name: "task"}},
		{ID: "3", Function: schema.FunctionCall{Name: "task"}},
	}

	result := m.truncateTaskToolCalls(toolCalls)

	if len(result) != 3 {
		t.Errorf("len(result) = %v, want 3", len(result))
	}
}

func TestSubagentLimitMiddleware_truncateTaskToolCalls_OverLimit(t *testing.T) {
	m := NewSubagentLimitMiddleware(2, nil)

	toolCalls := []schema.ToolCall{
		{ID: "1", Function: schema.FunctionCall{Name: "task"}},
		{ID: "2", Function: schema.FunctionCall{Name: "task"}},
		{ID: "3", Function: schema.FunctionCall{Name: "task"}}, // 这个应该被截断
		{ID: "4", Function: schema.FunctionCall{Name: "task"}}, // 这个应该被截断
	}

	result := m.truncateTaskToolCalls(toolCalls)

	if len(result) != 2 {
		t.Errorf("len(result) = %v, want 2", len(result))
	}
	if result[0].ID != "1" {
		t.Error("first task should be preserved")
	}
	if result[1].ID != "2" {
		t.Error("second task should be preserved")
	}
}

func TestSubagentLimitMiddleware_truncateTaskToolCalls_PreservesNonTaskTools(t *testing.T) {
	m := NewSubagentLimitMiddleware(2, nil)

	toolCalls := []schema.ToolCall{
		{ID: "1", Function: schema.FunctionCall{Name: "task"}},
		{ID: "2", Function: schema.FunctionCall{Name: "write_file"}}, // 非 task 工具
		{ID: "3", Function: schema.FunctionCall{Name: "task"}},
		{ID: "4", Function: schema.FunctionCall{Name: "task"}},   // 这个应该被截断
		{ID: "5", Function: schema.FunctionCall{Name: "read_file"}}, // 非 task 工具
	}

	result := m.truncateTaskToolCalls(toolCalls)

	if len(result) != 4 {
		t.Errorf("len(result) = %v, want 4", len(result))
	}

	// 验证非 task 工具都被保留
	foundWrite := false
	foundRead := false
	for _, tc := range result {
		if tc.Function.Name == "write_file" {
			foundWrite = true
		}
		if tc.Function.Name == "read_file" {
			foundRead = true
		}
	}
	if !foundWrite {
		t.Error("write_file should be preserved")
	}
	if !foundRead {
		t.Error("read_file should be preserved")
	}
}

func TestSubagentLimitMiddleware_AfterModel_NoMessages(t *testing.T) {
	m := NewDefaultSubagentLimitMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	// 没有消息
	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestSubagentLimitMiddleware_AfterModel_LastNotAssistant(t *testing.T) {
	m := NewDefaultSubagentLimitMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestSubagentLimitMiddleware_AfterModel_NoToolCalls(t *testing.T) {
	m := NewDefaultSubagentLimitMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, Content: "hi"},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestSubagentLimitMiddleware_AfterModel_UnderLimit(t *testing.T) {
	m := NewSubagentLimitMiddleware(3, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "1", Function: schema.FunctionCall{Name: "task"}},
			{ID: "2", Function: schema.FunctionCall{Name: "task"}},
		}},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil (under limit)", stateUpdate)
	}
}

func TestSubagentLimitMiddleware_AfterModel_Truncates(t *testing.T) {
	m := NewSubagentLimitMiddleware(1, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "1", Type: "function", Function: schema.FunctionCall{Name: "task", Arguments: "{}"}},
			{ID: "2", Type: "function", Function: schema.FunctionCall{Name: "task", Arguments: "{}"}},
		}},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)
	require.NoError(t, err)
	require.NotNil(t, stateUpdate)

	last := ts.Messages[len(ts.Messages)-1]
	require.Len(t, last.ToolCalls, 1)
	require.Equal(t, "1", last.ToolCalls[0].ID)
}

func TestSubagentLimitMiddleware_AfterModel_StateUpdate(t *testing.T) {
	m := NewSubagentLimitMiddleware(2, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, ToolCalls: []schema.ToolCall{
			{ID: "1", Function: schema.FunctionCall{Name: "task"}},
			{ID: "2", Function: schema.FunctionCall{Name: "task"}},
			{ID: "3", Function: schema.FunctionCall{Name: "task"}},
		}},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Fatal("AfterModel() stateUpdate = nil, want non-nil")
	}

	// 验证 stateUpdate 内容
	if stateUpdate["truncated"] != true {
		t.Error("stateUpdate.truncated should be true")
	}
	if stateUpdate["original_count"] != 3 {
		t.Error("stateUpdate.original_count should be 3")
	}
	if stateUpdate["limit"] != 2 {
		t.Error("stateUpdate.limit should be 2")
	}
}

func TestSubagentLimitMiddleware_Logger(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewSubagentLimitMiddleware(3, customLogger)

	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}

	m2 := NewSubagentLimitMiddleware(3, nil)
	if m2.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}
}

func TestSubagentLimitMiddleware_IntegrationWithChain(t *testing.T) {
	config := DefaultMiddlewareConfig()
	config.SubagentEnabled = true // 需要启用
	chain := BuildLeadAgentMiddlewares(config)

	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "subagent_limit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("SubagentLimitMiddleware not found in chain (need SubagentEnabled=true)")
	}
}

// BenchmarkSubagentLimitMiddleware_truncateTaskToolCalls 性能基准测试
func BenchmarkSubagentLimitMiddleware_truncateTaskToolCalls(b *testing.B) {
	m := NewSubagentLimitMiddleware(3, nil)
	toolCalls := []schema.ToolCall{
		{ID: "1", Function: schema.FunctionCall{Name: "task"}},
		{ID: "2", Function: schema.FunctionCall{Name: "task"}},
		{ID: "3", Function: schema.FunctionCall{Name: "task"}},
		{ID: "4", Function: schema.FunctionCall{Name: "task"}},
		{ID: "5", Function: schema.FunctionCall{Name: "task"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.truncateTaskToolCalls(toolCalls)
	}
}
