package middleware

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// SubagentLimitMiddleware 子代理限制中间件
// 一比一复刻 DeerFlow 的 SubagentLimitMiddleware
type SubagentLimitMiddleware struct {
	*BaseMiddleware
	maxConcurrentSubagents int
	logger                 *zap.Logger
}

// NewSubagentLimitMiddleware 创建子代理限制中间件
func NewSubagentLimitMiddleware(maxConcurrent int, logger *zap.Logger) *SubagentLimitMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	if maxConcurrent <= 0 {
		maxConcurrent = 3 // 默认 3
	}
	return &SubagentLimitMiddleware{
		BaseMiddleware:         NewBaseMiddleware("subagent_limit"),
		maxConcurrentSubagents: maxConcurrent,
		logger:                 logger,
	}
}

// NewDefaultSubagentLimitMiddleware 使用默认配置创建子代理限制中间件
func NewDefaultSubagentLimitMiddleware() *SubagentLimitMiddleware {
	return NewSubagentLimitMiddleware(3, nil)
}

// truncateTaskToolCalls 截断多余的 task 工具调用
func (m *SubagentLimitMiddleware) truncateTaskToolCalls(toolCalls []schema.ToolCall) []schema.ToolCall {
	if len(toolCalls) <= m.maxConcurrentSubagents {
		return toolCalls
	}

	result := make([]schema.ToolCall, 0, m.maxConcurrentSubagents)
	taskCount := 0

	for _, tc := range toolCalls {
		if tc.Function.Name == "task" {
			if taskCount >= m.maxConcurrentSubagents {
				m.logger.Debug("Skipping excess task tool call",
					zap.String("tool_call_id", tc.ID))
				continue
			}
			taskCount++
		}
		result = append(result, tc)
	}

	return result
}

// AfterModel 在模型响应后运行
func (m *SubagentLimitMiddleware) AfterModel(ctx context.Context, state *state.ThreadState) (map[string]interface{}, error) {
	if len(state.Messages) == 0 {
		return nil, nil
	}

	// 获取最后一条助手消息
	lastMsgIndex := len(state.Messages) - 1
	lastMsg := state.Messages[lastMsgIndex]
	if lastMsg.Role != schema.Assistant {
		return nil, nil
	}

	toolCalls := lastMsg.ToolCalls
	if len(toolCalls) == 0 {
		return nil, nil
	}

	taskCount := 0
	for _, tc := range toolCalls {
		if tc.Function.Name == "task" {
			taskCount++
		}
	}

	if taskCount <= m.maxConcurrentSubagents {
		// 在限制内，不需要截断
		return nil, nil
	}

	// 超过限制，截断多余的 task 工具调用
	m.logger.Warn("Too many concurrent subagents, truncating excess task tool calls",
		zap.Int("count", taskCount),
		zap.Int("limit", m.maxConcurrentSubagents),
		zap.Int("excess", taskCount-m.maxConcurrentSubagents))

	// 截断工具调用
	truncatedToolCalls := m.truncateTaskToolCalls(toolCalls)

	// 创建新的消息列表
	newMessages := make([]*schema.Message, len(state.Messages))
	copy(newMessages, state.Messages)

	newMsg := *lastMsg
	newMsg.ToolCalls = truncatedToolCalls
	newMessages[lastMsgIndex] = &newMsg

	// 更新状态
	state.Messages = newMessages

	return map[string]interface{}{
		"messages":       newMessages,
		"tool_calls":     truncatedToolCalls,
		"truncated":      true,
		"original_count": taskCount,
		"limit":          m.maxConcurrentSubagents,
	}, nil
}
