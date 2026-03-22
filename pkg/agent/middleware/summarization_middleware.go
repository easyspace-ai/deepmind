package middleware

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// SummarizationMiddleware 上下文摘要中间件
// 一比一复刻 DeerFlow 的 SummarizationMiddleware
type SummarizationMiddleware struct {
	*BaseMiddleware
	enabled bool
	logger  *zap.Logger
}

// NewSummarizationMiddleware 创建上下文摘要中间件
func NewSummarizationMiddleware(enabled bool, logger *zap.Logger) *SummarizationMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SummarizationMiddleware{
		BaseMiddleware: NewBaseMiddleware("summarization"),
		enabled:        enabled,
		logger:         logger,
	}
}

// NewDefaultSummarizationMiddleware 使用默认配置创建上下文摘要中间件
func NewDefaultSummarizationMiddleware() *SummarizationMiddleware {
	return NewSummarizationMiddleware(false, nil)
}

// shouldSummarize 检查是否应该进行摘要
func (m *SummarizationMiddleware) shouldSummarize(state *state.ThreadState) bool {
	if !m.enabled {
		return false
	}
	// 简化逻辑：消息数量超过阈值时进行摘要
	// 完整实现需要检查 token 数量等
	return len(state.Messages) > 20
}

// buildSummary 构建摘要
func (m *SummarizationMiddleware) buildSummary(state *state.ThreadState) string {
	// 简化实现：返回固定格式摘要
	// 完整实现需要调用 LLM 进行摘要
	return "[Conversation summarized - older messages compacted]"
}

// BeforeModel 在模型调用前运行
func (m *SummarizationMiddleware) BeforeModel(ctx context.Context, state *state.ThreadState) (map[string]interface{}, error) {
	if !m.shouldSummarize(state) {
		return nil, nil
	}

	oldCount := len(state.Messages)
	m.logger.Info("Triggering context summarization",
		zap.Int("message_count", oldCount))

	summary := m.buildSummary(state)

	// 保留最近的消息，替换旧消息为摘要
	// 简化实现：只保留最后 10 条消息
	if oldCount > 10 {
		summaryMessage := &schema.Message{
			Role:    schema.System,
			Content: summary,
		}

		recentMessages := state.Messages[oldCount-10:]

		newMessages := make([]*schema.Message, 0, 1+len(recentMessages))
		newMessages = append(newMessages, summaryMessage)
		newMessages = append(newMessages, recentMessages...)

		state.Messages = newMessages

		return map[string]interface{}{
			"messages":      newMessages,
			"summarized":    true,
			"summary_count": oldCount - len(newMessages),
		}, nil
	}

	return nil, nil
}
