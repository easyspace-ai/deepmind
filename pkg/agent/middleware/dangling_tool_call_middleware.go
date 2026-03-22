package middleware

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// DanglingToolCallMiddleware 处理挂起工具调用中间件
// 一比一复刻 DeerFlow 的 DanglingToolCallMiddleware
type DanglingToolCallMiddleware struct {
	*BaseMiddleware
	logger *zap.Logger
}

// NewDanglingToolCallMiddleware 创建挂起工具调用中间件
func NewDanglingToolCallMiddleware(logger *zap.Logger) *DanglingToolCallMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DanglingToolCallMiddleware{
		BaseMiddleware: NewBaseMiddleware("dangling_tool_call"),
		logger:         logger,
	}
}

// NewDefaultDanglingToolCallMiddleware 使用默认配置创建
func NewDefaultDanglingToolCallMiddleware() *DanglingToolCallMiddleware {
	return NewDanglingToolCallMiddleware(nil)
}

// buildPatchedMessages 构建补丁消息
// 一比一复刻 DeerFlow 的 _build_patched_messages
func (m *DanglingToolCallMiddleware) buildPatchedMessages(state *state.ThreadState) (*state.ThreadState, bool) {
	// 收集所有已存在的 ToolMessage 的 tool_call_id
	existingToolMsgIDs := make(map[string]bool)
	for _, msg := range state.Messages {
		if msg.Role == schema.Tool && msg.ToolCallID != "" {
			existingToolMsgIDs[msg.ToolCallID] = true
		}
	}

	// 检查是否需要打补丁
	needsPatch := false
	for _, msg := range state.Messages {
		if msg.Role != schema.Assistant {
			continue
		}
		for _, tc := range msg.ToolCalls {
			if tc.ID != "" && !existingToolMsgIDs[tc.ID] {
				needsPatch = true
				break
			}
		}
		if needsPatch {
			break
		}
	}

	if !needsPatch {
		return state, false
	}

	// 构建新消息列表，插入补丁
	var patchedMessages []*schema.Message
	patchedIDs := make(map[string]bool)
	patchCount := 0

	for _, msg := range state.Messages {
		patchedMessages = append(patchedMessages, msg)

		if msg.Role != schema.Assistant {
			continue
		}

		for _, tc := range msg.ToolCalls {
			if tc.ID != "" && !existingToolMsgIDs[tc.ID] && !patchedIDs[tc.ID] {
				toolName := tc.Function.Name
				placeholderMsg := &schema.Message{
					Role:       schema.Tool,
					Content:    "[Tool call was interrupted and did not return a result.]",
					ToolCallID: tc.ID,
					ToolName:   toolName,
				}
				patchedMessages = append(patchedMessages, placeholderMsg)
				patchedIDs[tc.ID] = true
				patchCount++
			}
		}
	}

	if patchCount > 0 {
		m.logger.Warn("Injecting placeholder ToolMessage(s) for dangling tool calls",
			zap.Int("count", patchCount))
	}

	newState := *state
	newState.Messages = patchedMessages
	return &newState, true
}

// BeforeModel 在模型调用前运行
func (m *DanglingToolCallMiddleware) BeforeModel(ctx context.Context, state *state.ThreadState) (map[string]interface{}, error) {
	patchedState, needsPatch := m.buildPatchedMessages(state)
	if !needsPatch {
		return nil, nil
	}

	// 更新状态
	*state = *patchedState

	return map[string]interface{}{
		"messages": patchedState.Messages,
	}, nil
}
