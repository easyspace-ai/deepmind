package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
)

func TestViewImageMiddleware_Name(t *testing.T) {
	m := NewViewImageMiddleware()

	if m.Name() != "view_image" {
		t.Errorf("Name() = %v, want 'view_image'", m.Name())
	}
}

func TestViewImageMiddleware_contains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"empty string", "", "", true},
		{"empty substr", "hello", "", true},
		{"contains", "hello world", "world", true},
		{"not contains", "hello", "world", false},
		{"exact match", "hello", "hello", true},
		{"substring at start", "hello world", "hello", true},
		{"substring longer", "hello", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestViewImageMiddleware_indexOf(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{"found at start", "hello world", "hello", 0},
		{"found in middle", "hello world", "world", 6},
		{"not found", "hello", "world", -1},
		{"exact match", "hello", "hello", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOf(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("indexOf(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestViewImageMiddleware_joinStrings(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{"empty", []string{}, "", ""},
		{"single", []string{"a"}, "", "a"},
		{"multiple", []string{"a", "b", "c"}, ", ", "a, b, c"},
		{"newline", []string{"a", "b"}, "\n", "a\nb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("joinStrings() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestViewImageMiddleware_getLastAssistantMessage(t *testing.T) {
	m := NewViewImageMiddleware()

	tests := []struct {
		name           string
		messages       []*schema.Message
		expectedIndex  int
		expectedFound  bool
	}{
		{
			"no messages",
			[]*schema.Message{},
			-1,
			false,
		},
		{
			"no assistant messages",
			[]*schema.Message{
				{Role: schema.User, Content: "hello"},
			},
			-1,
			false,
		},
		{
			"one assistant message",
			[]*schema.Message{
				{Role: schema.User, Content: "hello"},
				{Role: schema.Assistant, Content: "hi"},
			},
			1,
			true,
		},
		{
			"multiple assistant messages",
			[]*schema.Message{
				{Role: schema.User, Content: "hello"},
				{Role: schema.Assistant, Content: "hi"},
				{Role: schema.User, Content: "more"},
				{Role: schema.Assistant, Content: "ok"},
			},
			3,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := state.NewThreadState()
			ts.Messages = tt.messages
			idx, found := m.getLastAssistantMessage(ts)
			if idx != tt.expectedIndex {
				t.Errorf("getLastAssistantMessage() index = %v, want %v", idx, tt.expectedIndex)
			}
			if found != tt.expectedFound {
				t.Errorf("getLastAssistantMessage() found = %v, want %v", found, tt.expectedFound)
			}
		})
	}
}

func TestViewImageMiddleware_hasViewImageTool(t *testing.T) {
	m := NewViewImageMiddleware()

	tests := []struct {
		name     string
		msg      *schema.Message
		expected bool
	}{
		{
			"no tool calls",
			&schema.Message{Role: schema.Assistant, Content: "hi"},
			false,
		},
		{
			"other tool",
			&schema.Message{
				Role: schema.Assistant,
				ToolCalls: []schema.ToolCall{
					{Function: schema.FunctionCall{Name: "write_file"}},
				},
			},
			false,
		},
		{
			"has view_image",
			&schema.Message{
				Role: schema.Assistant,
				ToolCalls: []schema.ToolCall{
					{Function: schema.FunctionCall{Name: "view_image"}},
				},
			},
			true,
		},
		{
			"multiple tools including view_image",
			&schema.Message{
				Role: schema.Assistant,
				ToolCalls: []schema.ToolCall{
					{Function: schema.FunctionCall{Name: "write_file"}},
					{Function: schema.FunctionCall{Name: "view_image"}},
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.hasViewImageTool(tt.msg)
			if result != tt.expected {
				t.Errorf("hasViewImageTool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestViewImageMiddleware_allToolsCompleted(t *testing.T) {
	m := NewViewImageMiddleware()

	tests := []struct {
		name         string
		messages     []*schema.Message
		assistantIdx int
		expected     bool
	}{
		{
			"no tool calls",
			[]*schema.Message{
				{Role: schema.Assistant, Content: "hi"},
			},
			0,
			false,
		},
		{
			"all completed",
			[]*schema.Message{
				{
					Role: schema.Assistant,
					ToolCalls: []schema.ToolCall{
						{ID: "call-1"},
					},
				},
				{Role: schema.Tool, ToolCallID: "call-1"},
			},
			0,
			true,
		},
		{
			"some not completed",
			[]*schema.Message{
				{
					Role: schema.Assistant,
					ToolCalls: []schema.ToolCall{
						{ID: "call-1"},
						{ID: "call-2"},
					},
				},
				{Role: schema.Tool, ToolCallID: "call-1"},
			},
			0,
			false,
		},
		{
			"multiple tool calls all completed",
			[]*schema.Message{
				{
					Role: schema.Assistant,
					ToolCalls: []schema.ToolCall{
						{ID: "call-1"},
						{ID: "call-2"},
					},
				},
				{Role: schema.Tool, ToolCallID: "call-1"},
				{Role: schema.Tool, ToolCallID: "call-2"},
			},
			0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := state.NewThreadState()
			ts.Messages = tt.messages
			result := m.allToolsCompleted(ts, tt.assistantIdx)
			if result != tt.expected {
				t.Errorf("allToolsCompleted() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestViewImageMiddleware_createImageDetailsMessage(t *testing.T) {
	m := NewViewImageMiddleware()

	tests := []struct {
		name         string
		viewedImages map[string]state.ViewedImageData
		checkContent func(string) bool
	}{
		{
			"no images",
			map[string]state.ViewedImageData{},
			func(s string) bool { return s == "No images have been viewed." },
		},
		{
			"one image",
			map[string]state.ViewedImageData{
				"/path/to/image.jpg": {MimeType: "image/jpeg"},
			},
			func(s string) bool {
				return contains(s, "Here are the images you've viewed") &&
					contains(s, "/path/to/image.jpg") &&
					contains(s, "image/jpeg")
			},
		},
		{
			"multiple images",
			map[string]state.ViewedImageData{
				"/img1.jpg": {MimeType: "image/jpeg"},
				"/img2.png": {MimeType: "image/png"},
			},
			func(s string) bool {
				return contains(s, "/img1.jpg") && contains(s, "/img2.png")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := state.NewThreadState()
			ts.ViewedImages = tt.viewedImages
			result := m.createImageDetailsMessage(ts)
			if !tt.checkContent(result) {
				t.Errorf("createImageDetailsMessage() = %q, unexpected content", result)
			}
		})
	}
}

func TestViewImageMiddleware_shouldInjectImageMessage_NoMessages(t *testing.T) {
	m := NewViewImageMiddleware()
	ts := state.NewThreadState()

	result := m.shouldInjectImageMessage(ts)

	if result {
		t.Error("shouldInjectImageMessage() = true, want false (no messages)")
	}
}

func TestViewImageMiddleware_shouldInjectImageMessage_NoViewImage(t *testing.T) {
	m := NewViewImageMiddleware()
	ts := state.NewThreadState()
	ts.Messages = []*schema.Message{
		{Role: schema.Assistant, Content: "hi"},
	}

	result := m.shouldInjectImageMessage(ts)

	if result {
		t.Error("shouldInjectImageMessage() = true, want false (no view_image tool)")
	}
}

func TestViewImageMiddleware_shouldInjectImageMessage_AlreadyInjected(t *testing.T) {
	m := NewViewImageMiddleware()
	ts := state.NewThreadState()
	ts.Messages = []*schema.Message{
		{
			Role: schema.Assistant,
			ToolCalls: []schema.ToolCall{
				{Function: schema.FunctionCall{Name: "view_image"}},
			},
		},
		{Role: schema.User, Content: "Here are the images you've viewed: ..."},
	}

	result := m.shouldInjectImageMessage(ts)

	if result {
		t.Error("shouldInjectImageMessage() = true, want false (already injected)")
	}
}

func TestViewImageMiddleware_BeforeModel_NoInjection(t *testing.T) {
	m := NewViewImageMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestViewImageMiddleware_IntegrationWithChain(t *testing.T) {
	config := DefaultMiddlewareConfig()
	chain := BuildLeadAgentMiddlewares(config)

	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "view_image" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ViewImageMiddleware not found in chain")
	}
}

// BenchmarkViewImageMiddleware_shouldInjectImageMessage 性能基准测试
func BenchmarkViewImageMiddleware_shouldInjectImageMessage(b *testing.B) {
	m := NewViewImageMiddleware()
	ts := state.NewThreadState()
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.shouldInjectImageMessage(ts)
	}
}
