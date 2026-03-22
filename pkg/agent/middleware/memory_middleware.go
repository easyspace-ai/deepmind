package middleware

import (
	"context"

	"github.com/weibaohui/nanobot-go/pkg/agent/memory"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// MemoryMiddleware 记忆提取中间件
// 一比一复刻 DeerFlow 的 MemoryMiddleware
type MemoryMiddleware struct {
	*BaseMiddleware
	enabled bool
	logger  *zap.Logger
	manager *memory.Manager
}

// NewMemoryMiddleware 创建记忆提取中间件
func NewMemoryMiddleware(enabled bool, logger *zap.Logger) *MemoryMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MemoryMiddleware{
		BaseMiddleware: NewBaseMiddleware("memory"),
		enabled:        enabled,
		logger:         logger,
		manager:        memory.DefaultManager(),
	}
}

// NewDefaultMemoryMiddleware 使用默认配置创建记忆提取中间件
func NewDefaultMemoryMiddleware() *MemoryMiddleware {
	return NewMemoryMiddleware(false, nil)
}

// AfterModel 将最近对话摘要入队，供 memory.Manager 去抖持久化（见 pkg/agent/memory）。
func (m *MemoryMiddleware) AfterModel(ctx context.Context, ts *state.ThreadState) (map[string]interface{}, error) {
	if !m.enabled || ts == nil || m.manager == nil {
		return nil, nil
	}
	m.manager.EnqueueFromThreadState(ts)
	return nil, nil
}
