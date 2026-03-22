package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClarificationMiddleware_Name(t *testing.T) {
	m := NewClarificationMiddleware()

	if m.Name() != "clarification" {
		t.Errorf("Name() = %v, want 'clarification'", m.Name())
	}
}

func TestClarificationMiddleware_isChinese(t *testing.T) {
	m := NewClarificationMiddleware()

	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"empty", "", false},
		{"english only", "Hello World", false},
		{"chinese only", "你好世界", true},
		{"mixed", "Hello 你好", true},
		{"punctuation", "Hello, 世界!", true},
		{"numbers", "123 测试", true},
		{"special chars", "Hello!@#", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.isChinese(tt.text)
			if result != tt.expected {
				t.Errorf("isChinese(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

func TestClarificationMiddleware_formatClarificationMessage_QuestionOnly(t *testing.T) {
	m := NewClarificationMiddleware()

	args := map[string]interface{}{
		"question": "What do you mean?",
	}

	msg := m.formatClarificationMessage(args)

	if len(msg) == 0 {
		t.Error("formatClarificationMessage() should not return empty string")
	}
	if !contains(msg, "What do you mean?") {
		t.Error("message should contain the question")
	}
	if !contains(msg, "❓") {
		t.Error("message should contain default icon")
	}
}

func TestClarificationMiddleware_formatClarificationMessage_WithContext(t *testing.T) {
	m := NewClarificationMiddleware()

	args := map[string]interface{}{
		"question": "What do you mean?",
		"context":  "I need more details",
	}

	msg := m.formatClarificationMessage(args)

	if !contains(msg, "I need more details") {
		t.Error("message should contain the context")
	}
	if !contains(msg, "What do you mean?") {
		t.Error("message should contain the question")
	}
}

func TestClarificationMiddleware_formatClarificationMessage_WithType(t *testing.T) {
	m := NewClarificationMiddleware()

	testCases := []struct {
		clarType string
		icon     string
	}{
		{"missing_info", "❓"},
		{"ambiguous_requirement", "🤔"},
		{"approach_choice", "🔀"},
		{"risk_confirmation", "⚠️"},
		{"suggestion", "💡"},
		{"unknown_type", "❓"}, // 默认
	}

	for _, tc := range testCases {
		t.Run(tc.clarType, func(t *testing.T) {
			args := map[string]interface{}{
				"question":           "Test",
				"clarification_type": tc.clarType,
			}
			msg := m.formatClarificationMessage(args)
			if !contains(msg, tc.icon) {
				t.Errorf("message should contain %s icon for type %s", tc.icon, tc.clarType)
			}
		})
	}
}

func TestClarificationMiddleware_formatClarificationMessage_WithOptions(t *testing.T) {
	m := NewClarificationMiddleware()

	args := map[string]interface{}{
		"question": "Which one?",
		"options":  []interface{}{"Option A", "Option B", "Option C"},
	}

	msg := m.formatClarificationMessage(args)

	if !contains(msg, "Options:") {
		t.Error("message should contain options header")
	}
	if !contains(msg, "1. Option A") {
		t.Error("message should contain option 1")
	}
	if !contains(msg, "2. Option B") {
		t.Error("message should contain option 2")
	}
	if !contains(msg, "3. Option C") {
		t.Error("message should contain option 3")
	}
}

func TestClarificationMiddleware_formatClarificationMessage_ChineseOptions(t *testing.T) {
	m := NewClarificationMiddleware()

	args := map[string]interface{}{
		"question": "选哪个？",
		"options":  []interface{}{"选项 A", "选项 B"},
	}

	msg := m.formatClarificationMessage(args)

	if !contains(msg, "选项：") {
		t.Error("message should contain Chinese options header")
	}
}

func TestClarificationMiddleware_formatClarificationMessage_EmptyQuestion_Chinese(t *testing.T) {
	m := NewClarificationMiddleware()

	args := map[string]interface{}{
		"question": "",
		"context":  "需要更多信息",
	}

	msg := m.formatClarificationMessage(args)

	if !contains(msg, "请补充") {
		t.Error("empty question with Chinese context should use Chinese fallback")
	}
}

func TestClarificationMiddleware_formatClarificationMessage_EmptyQuestion_English(t *testing.T) {
	m := NewClarificationMiddleware()

	args := map[string]interface{}{
		"question": "",
		"context":  "Need more info",
	}

	msg := m.formatClarificationMessage(args)

	if !contains(msg, "Please provide more details") {
		t.Error("empty question with English context should use English fallback")
	}
}

func TestClarificationMiddleware_handleClarification_NotAskClarification(t *testing.T) {
	m := NewClarificationMiddleware()

	toolCall := &ToolCallInfo{
		Name: "other_tool",
		Args: map[string]interface{}{},
	}

	update, shouldInterrupt := m.handleClarification(toolCall)

	if update != nil {
		t.Error("handleClarification() update should be nil for non-ask_clarification tool")
	}
	if shouldInterrupt {
		t.Error("handleClarification() shouldInterrupt should be false for non-ask_clarification tool")
	}
}

func TestClarificationMiddleware_handleClarification_AskClarification(t *testing.T) {
	m := NewClarificationMiddleware()

	toolCall := &ToolCallInfo{
		Name: "ask_clarification",
		Args: map[string]interface{}{
			"question": "What do you mean?",
		},
	}

	update, shouldInterrupt := m.handleClarification(toolCall)

	if update == nil {
		t.Fatal("handleClarification() update should not be nil")
	}
	if !shouldInterrupt {
		t.Error("handleClarification() shouldInterrupt should be true")
	}

	// 验证 update 内容
	clarMsg, ok := update["clarification_message"]
	if !ok {
		t.Error("update should contain 'clarification_message'")
	}
	if clarMsg == "" {
		t.Error("clarification_message should not be empty")
	}

	interrupt, ok := update["interrupt"]
	if !ok {
		t.Error("update should contain 'interrupt'")
	}
	if interrupt != true {
		t.Error("interrupt should be true")
	}
}

func TestClarificationMiddleware_handleClarification_NilArgs(t *testing.T) {
	m := NewClarificationMiddleware()

	toolCall := &ToolCallInfo{
		Name: "ask_clarification",
		Args: nil,
	}

	update, shouldInterrupt := m.handleClarification(toolCall)

	if update == nil {
		t.Fatal("handleClarification() update should not be nil with nil args")
	}
	if !shouldInterrupt {
		t.Error("handleClarification() shouldInterrupt should be true")
	}
}

func TestClarificationMiddleware_WrapToolCall_NotAskClarification(t *testing.T) {
	m := NewClarificationMiddleware()
	ctx := context.Background()

	toolCall := &ToolCallInfo{
		Name: "other_tool",
		Args: map[string]interface{}{},
	}

	called := false
	next := func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error) {
		called = true
		return &ToolCallResult{Content: "result"}, nil
	}

	result, err := m.WrapToolCall(ctx, toolCall, next)

	if err != nil {
		t.Errorf("WrapToolCall() error = %v, want nil", err)
	}
	if !called {
		t.Error("next handler should be called for non-ask_clarification tool")
	}
	if result.Content != "result" {
		t.Error("result content mismatch")
	}
}

func TestClarificationMiddleware_WrapToolCall_AskClarification(t *testing.T) {
	m := NewClarificationMiddleware()
	ctx := context.Background()

	toolCall := &ToolCallInfo{
		Name: "ask_clarification",
		Args: map[string]interface{}{
			"question": "What do you mean?",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error) {
		nextCalled = true
		return &ToolCallResult{Content: "result"}, nil
	}

	result, err := m.WrapToolCall(ctx, toolCall, next)

	if err != nil {
		t.Errorf("WrapToolCall() error = %v, want nil", err)
	}
	if nextCalled {
		t.Error("next handler should NOT be called for ask_clarification tool")
	}
	if result == nil {
		t.Fatal("WrapToolCall() result should not be nil")
	}
	if !result.Interrupt {
		t.Error("result.Interrupt should be true")
	}
	if len(result.Content) == 0 {
		t.Error("result.Content should not be empty")
	}
}

func TestClarificationMiddleware_FormatMessage_EmptyQuestionUsesLocale(t *testing.T) {
	m := NewClarificationMiddleware()
	msg := m.formatClarificationMessage(map[string]interface{}{
		"question": "",
		"context":  "需要更多信息",
	})
	require.Contains(t, msg, "请补充")
}

func TestClarificationMiddleware_HandleClarification(t *testing.T) {
	m := NewClarificationMiddleware()
	upd, hit := m.handleClarification(&ToolCallInfo{
		Name: "ask_clarification",
		Args: map[string]interface{}{
			"question":           "OK?",
			"clarification_type": "missing_info",
			"context":            "ctx",
			"options":            []interface{}{"A", "B"},
		},
	})
	require.True(t, hit)
	require.True(t, upd["interrupt"].(bool))
	require.Contains(t, upd["clarification_message"].(string), "OK?")
}

func TestClarificationMiddleware_IntegrationWithChain(t *testing.T) {
	config := DefaultMiddlewareConfig()
	chain := BuildLeadAgentMiddlewares(config)

	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "clarification" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ClarificationMiddleware not found in chain")
	}
}

// BenchmarkClarificationMiddleware_formatClarificationMessage 性能基准测试
func BenchmarkClarificationMiddleware_formatClarificationMessage(b *testing.B) {
	m := NewClarificationMiddleware()
	args := map[string]interface{}{
		"question": "What do you mean?",
		"context":  "Need more info",
		"options":  []interface{}{"A", "B", "C"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.formatClarificationMessage(args)
	}
}
