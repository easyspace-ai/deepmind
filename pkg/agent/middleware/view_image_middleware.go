package middleware

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
)

// ViewImageMiddleware 图像处理中间件
// 一比一复刻 DeerFlow 的 ViewImageMiddleware
type ViewImageMiddleware struct {
	*BaseMiddleware
}

// NewViewImageMiddleware 创建图像处理中间件
func NewViewImageMiddleware() *ViewImageMiddleware {
	return &ViewImageMiddleware{
		BaseMiddleware: NewBaseMiddleware("view_image"),
	}
}

// getLastAssistantMessage 获取最后一条 AI 消息
func (m *ViewImageMiddleware) getLastAssistantMessage(state *state.ThreadState) (int, bool) {
	for i := len(state.Messages) - 1; i >= 0; i-- {
		if state.Messages[i].Role == schema.Assistant {
			return i, true
		}
	}
	return -1, false
}

// hasViewImageTool 检查是否有 view_image 工具调用
func (m *ViewImageMiddleware) hasViewImageTool(msg *schema.Message) bool {
	for _, tc := range msg.ToolCalls {
		if tc.Function.Name == "view_image" {
			return true
		}
	}
	return false
}

// allToolsCompleted 检查所有工具是否都完成
func (m *ViewImageMiddleware) allToolsCompleted(state *state.ThreadState, assistantIdx int) bool {
	assistantMsg := state.Messages[assistantIdx]

	toolCallIDs := make(map[string]bool)
	for _, tc := range assistantMsg.ToolCalls {
		if tc.ID != "" {
			toolCallIDs[tc.ID] = true
		}
	}

	if len(toolCallIDs) == 0 {
		return false
	}

	// 检查是否所有工具调用都有对应的 ToolMessage
	completedToolIDs := make(map[string]bool)
	for i := assistantIdx + 1; i < len(state.Messages); i++ {
		msg := state.Messages[i]
		if msg.Role == schema.Tool && msg.ToolCallID != "" {
			completedToolIDs[msg.ToolCallID] = true
		}
	}

	// 检查是否所有工具调用都完成了
	for id := range toolCallIDs {
		if !completedToolIDs[id] {
			return false
		}
	}
	return true
}

// shouldInjectImageMessage 检查是否应该注入图像消息
func (m *ViewImageMiddleware) shouldInjectImageMessage(state *state.ThreadState) bool {
	if len(state.Messages) == 0 {
		return false
	}

	// 获取最后一条 AI 消息
	assistantIdx, found := m.getLastAssistantMessage(state)
	if !found {
		return false
	}

	assistantMsg := state.Messages[assistantIdx]

	// 检查是否有 view_image 工具调用
	if !m.hasViewImageTool(assistantMsg) {
		return false
	}

	// 检查是否所有工具都完成了
	if !m.allToolsCompleted(state, assistantIdx) {
		return false
	}

	// 检查是否已经添加过图像消息
	for i := assistantIdx + 1; i < len(state.Messages); i++ {
		msg := state.Messages[i]
		if msg.Role == schema.User {
			contentStr := msg.Content
			if contains(contentStr, "Here are the images you've viewed") || contains(contentStr, "Here are the details of the images you've viewed") {
				return false
			}
		}
	}

	return true
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (indexOf(s, substr) >= 0))
}

// indexOf 返回子串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// createImageDetailsMessage 创建图像详情消息
func (m *ViewImageMiddleware) createImageDetailsMessage(state *state.ThreadState) string {
	if len(state.ViewedImages) == 0 {
		return "No images have been viewed."
	}

	// 简化处理：返回文本描述
	var parts []string
	parts = append(parts, "Here are the images you've viewed:")

	for path, data := range state.ViewedImages {
		parts = append(parts, "")
		parts = append(parts, "- **"+path+"** ("+data.MimeType+")")
	}

	return joinStrings(parts, "\n")
}

// joinStrings 连接字符串
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// injectImageMessage 注入图像消息
func (m *ViewImageMiddleware) injectImageMessage(state *state.ThreadState) (map[string]interface{}, bool) {
	if !m.shouldInjectImageMessage(state) {
		return nil, false
	}

	// 创建图像详情消息
	imageContent := m.createImageDetailsMessage(state)

	// 插入新的用户消息
	imageMsg := &schema.Message{
		Role:    schema.User,
		Content: imageContent,
	}
	state.Messages = append(state.Messages, imageMsg)

	return map[string]interface{}{
		"messages": state.Messages,
	}, true
}

// BeforeModel 在模型调用前运行
func (m *ViewImageMiddleware) BeforeModel(ctx context.Context, state *state.ThreadState) (map[string]interface{}, error) {
	update, injected := m.injectImageMessage(state)
	if !injected {
		return nil, nil
	}
	return update, nil
}
