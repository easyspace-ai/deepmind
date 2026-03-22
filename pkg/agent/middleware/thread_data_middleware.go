package middleware

import (
	"context"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/config"
)

// ThreadDataMiddleware 创建线程目录中间件
// 一比一复刻 DeerFlow 的 ThreadDataMiddleware
type ThreadDataMiddleware struct {
	*BaseMiddleware
	baseDir   string
	lazyInit   bool
	paths       *config.Paths
}

// NewThreadDataMiddleware 创建线程数据中间件
func NewThreadDataMiddleware(baseDir string, lazyInit bool) *ThreadDataMiddleware {
	paths := config.NewPaths(baseDir)
	return &ThreadDataMiddleware{
		BaseMiddleware: NewBaseMiddleware("thread_data"),
		baseDir:        baseDir,
		lazyInit:       lazyInit,
		paths:          paths,
	}
}

// NewDefaultThreadDataMiddleware 使用默认配置创建线程数据中间件
func NewDefaultThreadDataMiddleware() *ThreadDataMiddleware {
	return NewThreadDataMiddleware("", true)
}

// getThreadPaths 获取线程路径
func (m *ThreadDataMiddleware) getThreadPaths(threadID string) map[string]string {
	return map[string]string{
		"workspace_path": m.paths.SandboxWorkDir(threadID),
		"uploads_path":   m.paths.SandboxUploadsDir(threadID),
		"outputs_path":   m.paths.SandboxOutputsDir(threadID),
	}
}

// createThreadDirectories 创建线程目录
func (m *ThreadDataMiddleware) createThreadDirectories(threadID string) map[string]string {
	m.paths.EnsureThreadDirs(threadID)
	return m.getThreadPaths(threadID)
}

// BeforeAgent 在 Agent 执行前运行
func (m *ThreadDataMiddleware) BeforeAgent(ctx context.Context, ts *state.ThreadState, threadID string) (map[string]interface{}, error) {
	if threadID == "" {
		return nil, ErrThreadIDRequired
	}

	var paths map[string]string
	if m.lazyInit {
		// Lazy initialization: only compute paths, don't create directories
		paths = m.getThreadPaths(threadID)
	} else {
		// Eager initialization: create directories immediately
		paths = m.createThreadDirectories(threadID)
	}

	ts.ThreadData = &state.ThreadDataState{
		WorkspacePath: paths["workspace_path"],
		UploadsPath:   paths["uploads_path"],
		OutputsPath:   paths["outputs_path"],
	}

	return map[string]interface{}{
		"thread_data": ts.ThreadData,
	}, nil
}
