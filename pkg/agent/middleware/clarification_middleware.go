package middleware

import (
	"context"
	"fmt"
	"strings"
)

// ClarificationMiddleware 询问澄清中间件
// 一比一复刻 DeerFlow 的 ClarificationMiddleware
type ClarificationMiddleware struct {
	*BaseMiddleware
}

// NewClarificationMiddleware 创建询问澄清中间件
func NewClarificationMiddleware() *ClarificationMiddleware {
	return &ClarificationMiddleware{
		BaseMiddleware: NewBaseMiddleware("clarification"),
	}
}

// isChinese 检查是否包含中文字符
func (m *ClarificationMiddleware) isChinese(text string) bool {
	for _, r := range text {
		if '\u4e00' <= r && r <= '\u9fff' {
			return true
		}
	}
	return false
}

// formatClarificationMessage 格式化澄清消息
// 一比一复刻 DeerFlow 的 _format_clarification_message
func (m *ClarificationMiddleware) formatClarificationMessage(args map[string]interface{}) string {
	question, _ := args["question"].(string)
	question = strings.TrimSpace(question)
	clarificationType, _ := args["clarification_type"].(string)
	contextStr, _ := args["context"].(string)
	contextStr = strings.TrimSpace(contextStr)
	optionsIf, _ := args["options"].([]interface{})

	var options []string
	for _, optIf := range optionsIf {
		if opt, ok := optIf.(string); ok {
			options = append(options, opt)
		}
	}

	// 类型图标
	typeIcons := map[string]string{
		"missing_info":          "❓",
		"ambiguous_requirement": "🤔",
		"approach_choice":       "🔀",
		"risk_confirmation":     "⚠️",
		"suggestion":            "💡",
	}

	icon := typeIcons[clarificationType]
	if icon == "" {
		icon = "❓"
	}

	if question == "" {
		if m.isChinese(contextStr) {
			question = "请补充更多信息。"
		} else {
			question = "Please provide more details."
		}
	}

	// 构建消息
	var messageParts []string

	// 添加图标和问题
	if contextStr != "" {
		messageParts = append(messageParts, fmt.Sprintf("%s %s", icon, contextStr))
		messageParts = append(messageParts, fmt.Sprintf("\n%s", question))
	} else {
		messageParts = append(messageParts, fmt.Sprintf("%s %s", icon, question))
	}

	// 添加选项
	if len(options) > 0 {
		messageParts = append(messageParts, "")
		optHeader := "Options:"
		if m.isChinese(question) || m.isChinese(contextStr) {
			optHeader = "选项："
		}
		messageParts = append(messageParts, optHeader)
		for i, opt := range options {
			messageParts = append(messageParts, fmt.Sprintf("  %d. %s", i+1, opt))
		}
	}

	return strings.Join(messageParts, "\n")
}

// handleClarification 处理澄清请求
func (m *ClarificationMiddleware) handleClarification(toolCall *ToolCallInfo) (map[string]interface{}, bool) {
	if toolCall.Name != "ask_clarification" {
		return nil, false
	}

	args := toolCall.Args
	if args == nil {
		args = make(map[string]interface{})
	}

	// 格式化澄清消息
	formattedMessage := m.formatClarificationMessage(args)

	// 返回状态更新（中断执行）
	return map[string]interface{}{
		"clarification_message": formattedMessage,
		"interrupt":             true,
	}, true
}

// WrapToolCall 包装工具调用
func (m *ClarificationMiddleware) WrapToolCall(ctx context.Context, toolCall *ToolCallInfo, next ToolCallHandler) (*ToolCallResult, error) {
	update, shouldInterrupt := m.handleClarification(toolCall)
	if shouldInterrupt {
		return &ToolCallResult{
			Content:     update["clarification_message"].(string),
			StateUpdate: update,
			Interrupt:   true,
		}, nil
	}
	return next(ctx, toolCall)
}
