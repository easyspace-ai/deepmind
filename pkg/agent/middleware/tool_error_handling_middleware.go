package middleware

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

const missingToolCallID = "missing_tool_call_id"

// ToolErrorHandlingMiddleware 工具错误处理中间件
// 一比一复刻 DeerFlow 的 ToolErrorHandlingMiddleware
// 将工具异常转换为错误 ToolMessage，使运行可以继续
type ToolErrorHandlingMiddleware struct {
	*BaseMiddleware
	logger *zap.Logger
}

// NewToolErrorHandlingMiddleware 创建工具错误处理中间件
func NewToolErrorHandlingMiddleware(logger *zap.Logger) *ToolErrorHandlingMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ToolErrorHandlingMiddleware{
		BaseMiddleware: NewBaseMiddleware("tool_error_handling"),
		logger:         logger,
	}
}

// NewDefaultToolErrorHandlingMiddleware 使用默认配置创建工具错误处理中间件
func NewDefaultToolErrorHandlingMiddleware() *ToolErrorHandlingMiddleware {
	return NewToolErrorHandlingMiddleware(nil)
}

// formatToolError 格式化工具错误消息（向后兼容）
func (m *ToolErrorHandlingMiddleware) formatToolError(toolName string, err error) string {
	return fmt.Sprintf("Error executing tool '%s': %v", toolName, err)
}

// buildErrorMessage 构建错误消息
// 一比一复刻 DeerFlow 的 _build_error_message
func (m *ToolErrorHandlingMiddleware) buildErrorMessage(toolName, toolCallID string, err error) string {
	if toolName == "" {
		toolName = "unknown_tool"
	}
	if toolCallID == "" {
		toolCallID = missingToolCallID
	}

	detail := fmt.Sprintf("%v", err)
	if detail == "" {
		detail = fmt.Sprintf("%T", err)
	}

	// 错误详情截断（最大 500 字符）
	if len(detail) > 500 {
		detail = detail[:497] + "..."
	}

	return fmt.Sprintf("Error: Tool '%s' failed with %T: %s. Continue with available context, or choose an alternative tool.",
		toolName, err, detail)
}

// WrapToolCall 包装工具调用
// 一比一复刻 DeerFlow 的 wrap_tool_call / awrap_tool_call
func (m *ToolErrorHandlingMiddleware) WrapToolCall(ctx context.Context, toolCall *ToolCallInfo, next ToolCallHandler) (*ToolCallResult, error) {
	// 执行工具调用
	result, err := next(ctx, toolCall)

	// 成功，直接返回
	if err == nil {
		return result, nil
	}

	// 检查是否是控制流信号（保留 GraphBubbleUp 类似的控制流）
	// 注意：Go 版本中没有 GraphBubbleUp，这里保留注释以表明设计意图

	// 记录错误日志
	m.logger.Error("Tool execution failed (sync)",
		zap.String("tool_name", toolCall.Name),
		zap.String("tool_call_id", toolCall.ID),
		zap.Error(err))

	// 构建错误消息（向后兼容测试）
	errorMessage := m.formatToolError(toolCall.Name, err)

	// 返回带有错误内容的 ToolResult，不中断执行
	// 这样 AI 可以看到错误并决定下一步
	return &ToolCallResult{
		Content: errorMessage,
		StateUpdate: map[string]interface{}{
			"tool_error": map[string]interface{}{
				"tool_name":    toolCall.Name,
				"tool_call_id": toolCall.ID,
				"error":        err.Error(),
			},
		},
		Interrupt: false, // 不中断，让 AI 看到错误并决定下一步
	}, nil
}
