package middleware

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestToolErrorHandlingMiddleware_Name(t *testing.T) {
	m := NewDefaultToolErrorHandlingMiddleware()

	if m.Name() != "tool_error_handling" {
		t.Errorf("Name() = %v, want 'tool_error_handling'", m.Name())
	}
}

func TestToolErrorHandlingMiddleware_formatToolError(t *testing.T) {
	m := NewDefaultToolErrorHandlingMiddleware()

	err := errors.New("test error")
	msg := m.formatToolError("test_tool", err)

	expected := "Error executing tool 'test_tool': test error"
	if msg != expected {
		t.Errorf("formatToolError() = %v, want %v", msg, expected)
	}
}

func TestToolErrorHandlingMiddleware_WrapToolCall_Success(t *testing.T) {
	m := NewDefaultToolErrorHandlingMiddleware()
	ctx := context.Background()

	toolCall := &ToolCallInfo{
		ID:   "call-1",
		Name: "test_tool",
		Args: map[string]interface{}{"a": 1},
	}

	// 成功的 next handler
	next := func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error) {
		return &ToolCallResult{
			Content: "success",
		}, nil
	}

	result, err := m.WrapToolCall(ctx, toolCall, next)

	if err != nil {
		t.Errorf("WrapToolCall() error = %v, want nil", err)
	}
	if result == nil {
		t.Fatal("WrapToolCall() result = nil, want non-nil")
	}
	if result.Content != "success" {
		t.Errorf("WrapToolCall() content = %v, want 'success'", result.Content)
	}
	if result.Interrupt {
		t.Error("WrapToolCall() interrupt = true, want false")
	}
}

func TestToolErrorHandlingMiddleware_WrapToolCall_Error(t *testing.T) {
	m := NewDefaultToolErrorHandlingMiddleware()
	ctx := context.Background()

	toolCall := &ToolCallInfo{
		ID:   "call-1",
		Name: "test_tool",
		Args: map[string]interface{}{"a": 1},
	}

	testErr := errors.New("something went wrong")

	// 失败的 next handler
	next := func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error) {
		return nil, testErr
	}

	result, err := m.WrapToolCall(ctx, toolCall, next)

	if err != nil {
		t.Errorf("WrapToolCall() error = %v, want nil (error should be wrapped)", err)
	}
	if result == nil {
		t.Fatal("WrapToolCall() result = nil, want non-nil")
	}

	// 验证错误内容
	expectedContent := "Error executing tool 'test_tool': something went wrong"
	if result.Content != expectedContent {
		t.Errorf("WrapToolCall() content = %v, want %v", result.Content, expectedContent)
	}

	// 验证不中断
	if result.Interrupt {
		t.Error("WrapToolCall() interrupt = true, want false (should not interrupt on tool error)")
	}

	// 验证 StateUpdate
	if result.StateUpdate == nil {
		t.Error("WrapToolCall() StateUpdate = nil, want non-nil")
	}
	toolError, ok := result.StateUpdate["tool_error"]
	if !ok {
		t.Error("StateUpdate should contain 'tool_error'")
	}
	toolErrorMap, ok := toolError.(map[string]interface{})
	if !ok {
		t.Fatal("tool_error should be a map")
	}
	if toolErrorMap["tool_name"] != "test_tool" {
		t.Error("tool_error.tool_name mismatch")
	}
	if toolErrorMap["tool_call_id"] != "call-1" {
		t.Error("tool_error.tool_call_id mismatch")
	}
	if toolErrorMap["error"] != testErr.Error() {
		t.Error("tool_error.error mismatch")
	}
}

func TestToolErrorHandlingMiddleware_WrapToolCall_ErrorDifferentTools(t *testing.T) {
	m := NewDefaultToolErrorHandlingMiddleware()
	ctx := context.Background()

	testCases := []struct {
		name     string
		toolName string
		errMsg   string
	}{
		{"filesystem tool", "write_file", "permission denied"},
		{"search tool", "tavily_search", "rate limit exceeded"},
		{"shell tool", "execute_command", "command not found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			toolCall := &ToolCallInfo{
				ID:   "call-1",
				Name: tc.toolName,
				Args: map[string]interface{}{},
			}

			testErr := errors.New(tc.errMsg)
			next := func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error) {
				return nil, testErr
			}

			result, err := m.WrapToolCall(ctx, toolCall, next)

			if err != nil {
				t.Errorf("WrapToolCall() error = %v, want nil", err)
			}

			expectedContent := "Error executing tool '" + tc.toolName + "': " + tc.errMsg
			if result.Content != expectedContent {
				t.Errorf("WrapToolCall() content = %v, want %v", result.Content, expectedContent)
			}
		})
	}
}

func TestToolErrorHandlingMiddleware_Logger(t *testing.T) {
	// 验证 logger 参数可以为 nil
	customLogger := zap.NewExample()
	m := NewToolErrorHandlingMiddleware(customLogger)

	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}

	// nil logger 应该使用 NopLogger
	m2 := NewToolErrorHandlingMiddleware(nil)
	if m2.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}
}

func TestToolErrorHandlingMiddleware_WrapToolCall_WithStateUpdateOnSuccess(t *testing.T) {
	m := NewDefaultToolErrorHandlingMiddleware()
	ctx := context.Background()

	toolCall := &ToolCallInfo{
		ID:   "call-1",
		Name: "test_tool",
		Args: map[string]interface{}{"a": 1},
	}

	// 成功的 next handler，带 state update
	next := func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error) {
		return &ToolCallResult{
			Content: "success",
			StateUpdate: map[string]interface{}{
				"key": "value",
			},
		}, nil
	}

	result, err := m.WrapToolCall(ctx, toolCall, next)

	if err != nil {
		t.Errorf("WrapToolCall() error = %v, want nil", err)
	}
	if result.StateUpdate["key"] != "value" {
		t.Error("WrapToolCall() should preserve StateUpdate from next handler")
	}
}

func TestToolErrorHandlingMiddleware_IntegrationWithChain(t *testing.T) {
	// 验证能正确加入中间件链
	config := DefaultMiddlewareConfig()
	chain := BuildLeadAgentMiddlewares(config)

	// 验证链中有 ToolErrorHandlingMiddleware
	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "tool_error_handling" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ToolErrorHandlingMiddleware not found in chain")
	}
}

// BenchmarkToolErrorHandlingMiddleware_WrapToolCall_Success 性能基准测试（成功路径）
func BenchmarkToolErrorHandlingMiddleware_WrapToolCall_Success(b *testing.B) {
	m := NewDefaultToolErrorHandlingMiddleware()
	ctx := context.Background()
	toolCall := &ToolCallInfo{
		ID:   "call-1",
		Name: "test_tool",
		Args: map[string]interface{}{"a": 1},
	}
	next := func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error) {
		return &ToolCallResult{Content: "success"}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.WrapToolCall(ctx, toolCall, next)
	}
}

// BenchmarkToolErrorHandlingMiddleware_WrapToolCall_Error 性能基准测试（错误路径）
func BenchmarkToolErrorHandlingMiddleware_WrapToolCall_Error(b *testing.B) {
	m := NewDefaultToolErrorHandlingMiddleware()
	ctx := context.Background()
	toolCall := &ToolCallInfo{
		ID:   "call-1",
		Name: "test_tool",
		Args: map[string]interface{}{"a": 1},
	}
	testErr := errors.New("test error")
	next := func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error) {
		return nil, testErr
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.WrapToolCall(ctx, toolCall, next)
	}
}
