package memory

import (
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/require"
)

// ============================================
// 测试：上传提及过滤
// ============================================

func TestStripUploadSentences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
	}{
		{
			name:     "空字符串",
			input:    "",
		},
		{
			name:     "无上传提及",
			input:    "User likes programming.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripUploadSentences(tt.input)
			require.NotNil(t, result)
		})
	}
}

func TestStripUploadMentions(t *testing.T) {
	m := CreateEmptyMemory()
	m.User.WorkContext.Summary = "User works on projects."
	m.Facts = []Fact{
		{ID: "2", Content: "User likes Go", Category: "preference"},
	}

	result := StripUploadMentions(m)
	require.NotNil(t, result)
	require.Len(t, result.Facts, 1)
}

// ============================================
// 测试：对话格式化
// ============================================

func TestFormatConversationForUpdate(t *testing.T) {
	tests := []struct {
		name     string
		messages []*schema.Message
	}{
		{
			name:     "空消息列表",
			messages: nil,
		},
		{
			name: "用户和助手消息",
			messages: []*schema.Message{
				{Role: schema.User, Content: "Hello!"},
				{Role: schema.Assistant, Content: "Hi there!"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatConversationForUpdate(tt.messages)
			require.NotNil(t, result)
		})
	}
}

// ============================================
// 测试：记忆格式化
// ============================================

func TestFormatMemoryForInjection(t *testing.T) {
	tests := []struct {
		name      string
		memory    *MemoryData
		maxTokens int
	}{
		{
			name:      "空记忆",
			memory:    nil,
			maxTokens: 2000,
		},
		{
			name:      "只有 user context",
			memory:    CreateEmptyMemory(),
			maxTokens: 2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMemoryForInjection(tt.memory, tt.maxTokens)
			require.NotNil(t, result)
		})
	}
}

func TestBuildMemoryUpdatePrompt(t *testing.T) {
	m := CreateEmptyMemory()
	conversation := "User: Hello\nAssistant: Hi"

	prompt, err := BuildMemoryUpdatePrompt(m, conversation)
	require.NoError(t, err)
	require.NotEmpty(t, prompt)
}

func TestBuildFactExtractionPrompt(t *testing.T) {
	prompt := BuildFactExtractionPrompt("Hello world")
	require.NotEmpty(t, prompt)
}

func TestCountTokens(t *testing.T) {
	require.Equal(t, 0, countTokens(""))
	require.Greater(t, countTokens("Hello world"), 0)
}

func TestCoerceConfidence(t *testing.T) {
	require.Equal(t, 0.0, coerceConfidence(-0.5, 0.0))
	require.Equal(t, 1.0, coerceConfidence(1.5, 0.0))
	require.Equal(t, 0.5, coerceConfidence(0.5, 0.0))
}

func TestFormatMemoryForPrompt(t *testing.T) {
	m := CreateEmptyMemory()
	m.User.WorkContext.Summary = "Test"
	result := FormatMemoryForPrompt(m)
	require.NotNil(t, result)
}
