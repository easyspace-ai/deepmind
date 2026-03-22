package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestTitleMiddleware_Name(t *testing.T) {
	m := NewDefaultTitleMiddleware()

	if m.Name() != "title" {
		t.Errorf("Name() = %v, want 'title'", m.Name())
	}
}

func TestTitleMiddleware_NewDefault(t *testing.T) {
	m := NewDefaultTitleMiddleware()

	if !m.enabled {
		t.Error("enabled = false, want true")
	}
}

func TestTitleMiddleware_NewWithConfig(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewTitleMiddleware(false, "gpt-4", customLogger)

	if m.enabled {
		t.Error("enabled = true, want false")
	}
	if m.modelName != "gpt-4" {
		t.Errorf("modelName = %v, want 'gpt-4'", m.modelName)
	}
	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}
}

func TestTitleMiddleware_min(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a smaller", 1, 2, 1},
		{"b smaller", 3, 2, 2},
		{"equal", 5, 5, 5},
		{"zero", 0, 10, 0},
		{"negative", -5, 3, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.want {
				t.Errorf("min(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.want)
			}
		})
	}
}

func TestTitleMiddleware_normalizeContent(t *testing.T) {
	m := NewDefaultTitleMiddleware()

	tests := []struct {
		name     string
		content  interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello world", "hello world"},
		{"empty string", "", ""},
		{"list of strings", []interface{}{"a", "b", "c"}, "a\nb\nc"},
		{"map with text", map[string]interface{}{"text": "map text"}, "map text"},
		{"map with content", map[string]interface{}{"content": "nested content"}, "nested content"},
		{"map with text and content", map[string]interface{}{"text": "text1", "content": "content1"}, "text1"},
		{"other type", 123, "123"},
		{"nested list", []interface{}{"a", []interface{}{"b", "c"}}, "a\nb\nc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.normalizeContent(tt.content)
			if result != tt.expected {
				t.Errorf("normalizeContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTitleMiddleware_shouldGenerateTitle_Disabled(t *testing.T) {
	m := NewTitleMiddleware(false, "", nil)
	ts := state.NewThreadState()

	result := m.shouldGenerateTitle(ts)

	if result {
		t.Error("shouldGenerateTitle() = true, want false (disabled)")
	}
}

func TestTitleMiddleware_shouldGenerateTitle_AlreadyHasTitle(t *testing.T) {
	m := NewDefaultTitleMiddleware()
	ts := state.NewThreadState()
	ts.Title = "Existing Title"

	result := m.shouldGenerateTitle(ts)

	if result {
		t.Error("shouldGenerateTitle() = true, want false (already has title)")
	}
}

func TestTitleMiddleware_shouldGenerateTitle_NotEnoughMessages(t *testing.T) {
	m := NewDefaultTitleMiddleware()
	ts := state.NewThreadState()

	// 只有 1 条消息
	ts.Messages = []*schema.Message{
		{Role: "user", Content: "hello"},
	}

	result := m.shouldGenerateTitle(ts)

	if result {
		t.Error("shouldGenerateTitle() = true, want false (not enough messages)")
	}
}

func TestTitleMiddleware_shouldGenerateTitle_Ready(t *testing.T) {
	m := NewDefaultTitleMiddleware()
	ts := state.NewThreadState()

	// 第一次完整对话
	ts.Messages = []*schema.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi"},
	}

	result := m.shouldGenerateTitle(ts)

	if !result {
		t.Error("shouldGenerateTitle() = false, want true (first complete conversation)")
	}
}

func TestTitleMiddleware_shouldGenerateTitle_NotFirstConversation(t *testing.T) {
	m := NewDefaultTitleMiddleware()
	ts := state.NewThreadState()

	// 多轮对话
	ts.Messages = []*schema.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi"},
		{Role: "user", Content: "more"},
		{Role: "assistant", Content: "ok"},
	}

	result := m.shouldGenerateTitle(ts)

	if result {
		t.Error("shouldGenerateTitle() = true, want false (not first conversation)")
	}
}

func TestTitleMiddleware_buildTitlePrompt(t *testing.T) {
	m := NewDefaultTitleMiddleware()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: "user", Content: "Hello, how are you?"},
		{Role: "assistant", Content: "I'm fine, thank you!"},
	}

	prompt, userMsg := m.buildTitlePrompt(ts)

	if userMsg != "Hello, how are you?" {
		t.Errorf("userMsg = %q, want 'Hello, how are you?'", userMsg)
	}
	if len(prompt) == 0 {
		t.Error("prompt should not be empty")
	}
}

func TestTitleMiddleware_buildTitlePrompt_TruncatesLongContent(t *testing.T) {
	m := NewDefaultTitleMiddleware()
	ts := state.NewThreadState()

	// 创建很长的消息
	longUserMsg := make([]byte, 1000)
	for i := range longUserMsg {
		longUserMsg[i] = 'a'
	}
	longAssistantMsg := make([]byte, 1000)
	for i := range longAssistantMsg {
		longAssistantMsg[i] = 'b'
	}

	ts.Messages = []*schema.Message{
		{Role: "user", Content: string(longUserMsg)},
		{Role: "assistant", Content: string(longAssistantMsg)},
	}

	prompt, userMsg := m.buildTitlePrompt(ts)

	if len(userMsg) > 500 {
		t.Errorf("userMsg length = %v, want <= 500", len(userMsg))
	}
	if len(prompt) == 0 {
		t.Error("prompt should not be empty")
	}
}

func TestTitleMiddleware_parseTitle(t *testing.T) {
	m := NewDefaultTitleMiddleware()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"simple", "My Title", "My Title"},
		{"with whitespace", "  My Title  ", "My Title"},
		{"with quotes", "\"My Title\"", "My Title"},
		{"with single quotes", "'My Title'", "My Title"},
		{"too long", string(make([]byte, 200)), string(make([]byte, 100))},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.parseTitle(tt.content)
			if len(result) > m.maxChars {
				t.Errorf("parseTitle() length = %v, want <= %v", len(result), m.maxChars)
			}
		})
	}
}

func TestTitleMiddleware_fallbackTitle(t *testing.T) {
	m := NewDefaultTitleMiddleware()

	tests := []struct {
		name     string
		userMsg  string
		check    func(string) bool
	}{
		{"empty", "", func(s string) bool { return s == "New Conversation" }},
		{"short", "Hello", func(s string) bool { return s == "Hello" }},
		{"long", string(make([]byte, 100)), func(s string) bool { return len(s) <= 53 && len(s) > 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.fallbackTitle(tt.userMsg)
			if !tt.check(result) {
				t.Errorf("fallbackTitle(%q) = %q, unexpected", tt.userMsg, result)
			}
		})
	}
}

func TestTitleMiddleware_AfterModel_NotReady(t *testing.T) {
	m := NewDefaultTitleMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	// 不满足生成标题的条件
	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
	if ts.Title != "" {
		t.Error("Title should not be set")
	}
}

func TestTitleMiddleware_AfterModel_GeneratesTitle(t *testing.T) {
	m := NewDefaultTitleMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: "user", Content: "Hello, world!"},
		{Role: "assistant", Content: "Hi there!"},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Fatal("AfterModel() stateUpdate = nil, want non-nil")
	}
	if ts.Title == "" {
		t.Error("Title should be set")
	}

	// 验证 stateUpdate
	title, ok := stateUpdate["title"]
	if !ok {
		t.Error("stateUpdate should contain 'title'")
	}
	if title != ts.Title {
		t.Error("title in stateUpdate should match ts.Title")
	}
}

func TestTitleMiddleware_AfterModel_Disabled(t *testing.T) {
	m := NewTitleMiddleware(false, "", nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	ts.Messages = []*schema.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil (disabled)", stateUpdate)
	}
	if ts.Title != "" {
		t.Error("Title should not be set when disabled")
	}
}

func TestTitleMiddleware_Logger(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewTitleMiddleware(true, "", customLogger)

	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}

	m2 := NewTitleMiddleware(true, "", nil)
	if m2.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}
}

func TestTitleMiddleware_IntegrationWithChain(t *testing.T) {
	config := DefaultMiddlewareConfig()
	chain := BuildLeadAgentMiddlewares(config)

	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "title" {
			found = true
			break
		}
	}
	if !found {
		t.Error("TitleMiddleware not found in chain")
	}
}

// BenchmarkTitleMiddleware_shouldGenerateTitle 性能基准测试
func BenchmarkTitleMiddleware_shouldGenerateTitle(b *testing.B) {
	m := NewDefaultTitleMiddleware()
	ts := state.NewThreadState()
	ts.Messages = []*schema.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.shouldGenerateTitle(ts)
	}
}
