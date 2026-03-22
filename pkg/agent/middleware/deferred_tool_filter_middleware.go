package middleware

import (
	"context"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// DeferredToolFilterMiddleware 延迟工具过滤中间件
// 一比一复刻 DeerFlow 的 DeferredToolFilterMiddleware
// 从 request.tools 中移除延迟工具，使 LLM 只看到 active tool schemas
// 延迟工具通过 tool_search 在运行时发现
type DeferredToolFilterMiddleware struct {
	*BaseMiddleware
	logger        *zap.Logger
	deferredNames []string
}

// NewDeferredToolFilterMiddleware 创建延迟工具过滤中间件（无预配置工具名）。
func NewDeferredToolFilterMiddleware(logger *zap.Logger) *DeferredToolFilterMiddleware {
	return NewDeferredToolFilterMiddlewareWithNames(nil, logger)
}

// NewDeferredToolFilterMiddlewareWithNames 创建并附带延迟工具名列表（用于日志与后续 schema 过滤）。
func NewDeferredToolFilterMiddlewareWithNames(names []string, logger *zap.Logger) *DeferredToolFilterMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	cp := append([]string(nil), names...)
	return &DeferredToolFilterMiddleware{
		BaseMiddleware: NewBaseMiddleware("deferred_tool_filter"),
		logger:         logger,
		deferredNames:  cp,
	}
}

// NewDefaultDeferredToolFilterMiddleware 使用默认配置创建延迟工具过滤中间件
func NewDefaultDeferredToolFilterMiddleware() *DeferredToolFilterMiddleware {
	return NewDeferredToolFilterMiddleware(nil)
}

// FilterTools 过滤工具，移除延迟工具
// 一比一复刻 DeerFlow 的 _filter_tools
func (m *DeferredToolFilterMiddleware) FilterTools(toolNames []string) []string {
	// 构建延迟工具名称集合
	deferredNameSet := make(map[string]bool)
	for _, name := range m.deferredNames {
		deferredNameSet[name] = true
	}

	if len(deferredNameSet) == 0 {
		return toolNames
	}

	// 过滤掉延迟工具
	activeTools := make([]string, 0, len(toolNames))
	for _, name := range toolNames {
		if !deferredNameSet[name] {
			activeTools = append(activeTools, name)
		}
	}

	if len(activeTools) < len(toolNames) {
		m.logger.Debug("Filtered deferred tool schema(s) from model binding",
			zap.Int("filtered_count", len(toolNames)-len(activeTools)))
	}

	return activeTools
}

// BeforeModel 在模型调用前记录延迟工具上下文（完整 schema 过滤需在模型 Option 层配合实现）。
func (m *DeferredToolFilterMiddleware) BeforeModel(ctx context.Context, ts *state.ThreadState) (map[string]interface{}, error) {
	if len(m.deferredNames) == 0 {
		return nil, nil
	}
	m.logger.Debug("deferred tools registered for filter",
		zap.Strings("tool_names", m.deferredNames),
		zap.Int("message_count", len(ts.Messages)))
	return nil, nil
}
