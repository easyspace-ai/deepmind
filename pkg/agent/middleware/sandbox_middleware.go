package middleware

import (
	"context"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/sandbox"
	"go.uber.org/zap"
)

// SandboxMiddleware 沙箱生命周期中间件
// 一比一复刻 DeerFlow 的 SandboxMiddleware
type SandboxMiddleware struct {
	*BaseMiddleware
	provider sandbox.SandboxProvider
	logger   *zap.Logger
}

// NewSandboxMiddleware 创建沙箱中间件
func NewSandboxMiddleware(provider sandbox.SandboxProvider, logger *zap.Logger) *SandboxMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SandboxMiddleware{
		BaseMiddleware: NewBaseMiddleware("sandbox"),
		provider:       provider,
		logger:         logger,
	}
}

// NewDefaultSandboxMiddleware 使用默认配置创建沙箱中间件
func NewDefaultSandboxMiddleware() *SandboxMiddleware {
	return NewSandboxMiddleware(nil, nil)
}

// BeforeAgent 在 Agent 执行前运行
func (m *SandboxMiddleware) BeforeAgent(ctx context.Context, ts *state.ThreadState, threadID string) (map[string]interface{}, error) {
	if threadID == "" {
		return nil, nil
	}

	// 检查是否已经有沙箱
	if ts.Sandbox != nil && ts.Sandbox.SandboxID != "" {
		// 沙箱已存在
		if m.provider != nil {
			if _, exists := m.provider.Get(threadID); exists {
				return nil, nil
			}
		}
	}

	// 懒加载获取沙箱
	if m.provider == nil {
		// 没有 provider，使用本地沙箱占位符
		ts.Sandbox = &state.SandboxState{
			SandboxID: "local",
		}
	} else {
		// 使用 provider 获取沙箱
		sb, err := m.provider.Acquire(threadID)
		if err != nil {
			m.logger.Error("Failed to acquire sandbox", zap.String("thread_id", threadID), zap.Error(err))
			return nil, err
		}
		ts.Sandbox = &state.SandboxState{
			SandboxID: sb.ID(),
		}
		m.logger.Info("Sandbox acquired", zap.String("thread_id", threadID), zap.String("sandbox_id", ts.Sandbox.SandboxID))
	}

	return map[string]interface{}{
		"sandbox": ts.Sandbox,
	}, nil
}
