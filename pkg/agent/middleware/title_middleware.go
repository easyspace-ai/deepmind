package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// TitleMiddleware 自动生成标题中间件
// 一比一复刻 DeerFlow 的 TitleMiddleware
type TitleMiddleware struct {
	*BaseMiddleware
	enabled     bool
	modelName   string
	maxWords    int
	maxChars    int
	promptTemplate string
	logger      *zap.Logger
}

// NewTitleMiddleware 创建标题中间件
func NewTitleMiddleware(enabled bool, modelName string, logger *zap.Logger) *TitleMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TitleMiddleware{
		BaseMiddleware: NewBaseMiddleware("title"),
		enabled:        enabled,
		modelName:      modelName,
		maxWords:       10,
		maxChars:       100,
		promptTemplate: defaultTitlePromptTemplate,
		logger:         logger,
	}
}

// 默认标题生成提示词
const defaultTitlePromptTemplate = `Generate a concise title for this conversation (max {max_words} words):

User: {user_msg}
Assistant: {assistant_msg}

Title:`

// NewDefaultTitleMiddleware 使用默认配置创建标题中间件
func NewDefaultTitleMiddleware() *TitleMiddleware {
	return NewTitleMiddleware(true, "", nil)
}

// normalizeContent 标准化内容
func (m *TitleMiddleware) normalizeContent(content interface{}) string {
	if content == nil {
		return ""
	}
	if str, ok := content.(string); ok {
		return str
	}
	if list, ok := content.([]interface{}); ok {
		var parts []string
		for _, item := range list {
			parts = append(parts, m.normalizeContent(item))
		}
		return strings.Join(parts, "\n")
	}
	if mp, ok := content.(map[string]interface{}); ok {
		if text, ok := mp["text"].(string); ok {
			return text
		}
		if nested, ok := mp["content"]; ok {
			return m.normalizeContent(nested)
		}
	}
	return fmt.Sprintf("%v", content)
}

// shouldGenerateTitle 检查是否应该生成标题
func (m *TitleMiddleware) shouldGenerateTitle(state *state.ThreadState) bool {
	if !m.enabled {
		return false
	}

	// 检查是否已经有标题
	if state.Title != "" {
		return false
	}

	// 检查是否是第一次完整对话（至少 1 条用户消息 + 1 条 AI 响应）
	if len(state.Messages) < 2 {
		return false
	}

	// 统计用户和 AI 消息
	userCount := 0
	assistantCount := 0
	for _, msg := range state.Messages {
		if msg.Role == "user" {
			userCount++
		} else if msg.Role == "assistant" {
			assistantCount++
		}
	}

	// 第一次完整对话后生成标题
	return userCount == 1 && assistantCount >= 1
}

// buildTitlePrompt 构建标题生成提示词
func (m *TitleMiddleware) buildTitlePrompt(state *state.ThreadState) (prompt string, userMsg string) {
	var userContent string
	var assistantContent string

	// 提取第一条用户消息和第一条 AI 响应
	for _, msg := range state.Messages {
		if msg.Role == "user" && userContent == "" {
			userContent = m.normalizeContent(msg.Content)
		} else if msg.Role == "assistant" && assistantContent == "" {
			assistantContent = m.normalizeContent(msg.Content)
		}
	}

	// 截断到 500 字符
	if len(userContent) > 500 {
		userContent = userContent[:500]
	}
	if len(assistantContent) > 500 {
		assistantContent = assistantContent[:500]
	}

	prompt = strings.ReplaceAll(m.promptTemplate, "{max_words}", fmt.Sprintf("%d", m.maxWords))
	prompt = strings.ReplaceAll(prompt, "{user_msg}", userContent)
	prompt = strings.ReplaceAll(prompt, "{assistant_msg}", assistantContent)

	return prompt, userContent
}

// parseTitle 解析标题
func (m *TitleMiddleware) parseTitle(content string) string {
	title := strings.TrimSpace(content)
	title = strings.Trim(title, `"'`)
	if len(title) > m.maxChars {
		title = title[:m.maxChars]
	}
	return title
}

// fallbackTitle 回退标题
func (m *TitleMiddleware) fallbackTitle(userMsg string) string {
	fallbackChars := min(m.maxChars, 50)
	if len(userMsg) > fallbackChars {
		return strings.TrimSpace(userMsg[:fallbackChars]) + "..."
	}
	if userMsg == "" {
		return "New Conversation"
	}
	return userMsg
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AfterModel 在模型响应后运行
func (m *TitleMiddleware) AfterModel(ctx context.Context, state *state.ThreadState) (map[string]interface{}, error) {
	if !m.shouldGenerateTitle(state) {
		return nil, nil
	}

	_, userMsg := m.buildTitlePrompt(state)

	// 注意：实际使用时需要调用 LLM 生成标题（将 buildTitlePrompt 的完整 prompt 传入模型）
	// 这里简化处理，使用 fallback 标题
	title := m.fallbackTitle(userMsg)

	m.logger.Info("Generated thread title", zap.String("title", title))

	// 更新状态
	state.Title = title

	return map[string]interface{}{
		"title": title,
	}, nil
}
