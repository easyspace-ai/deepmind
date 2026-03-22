package middleware

import (
	"context"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// TodoListMiddleware 任务列表中间件
// 一比一复刻 DeerFlow 的 TodoMiddleware
type TodoListMiddleware struct {
	*BaseMiddleware
	enabled bool
	logger  *zap.Logger
}

// NewTodoListMiddleware 创建任务列表中间件
func NewTodoListMiddleware(enabled bool, logger *zap.Logger) *TodoListMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TodoListMiddleware{
		BaseMiddleware: NewBaseMiddleware("todo_list"),
		enabled:        enabled,
		logger:         logger,
	}
}

// NewDefaultTodoListMiddleware 使用默认配置创建任务列表中间件
func NewDefaultTodoListMiddleware() *TodoListMiddleware {
	return NewTodoListMiddleware(false, nil)
}

// AfterModel 任务列表追踪：记录当前 todos 数量，提醒模型使用 write_todos 维护状态。
func (m *TodoListMiddleware) AfterModel(ctx context.Context, ts *state.ThreadState) (map[string]interface{}, error) {
	if !m.enabled || ts == nil {
		return nil, nil
	}
	m.logger.Debug("todo_list middleware",
		zap.Int("todo_count", len(ts.Todos)),
		zap.String("hint", "use write_todos tool to update task list"))
	return nil, nil
}
