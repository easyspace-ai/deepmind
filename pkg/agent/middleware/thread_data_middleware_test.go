package middleware

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/config"
)

func TestThreadDataMiddleware_Name(t *testing.T) {
	m := NewDefaultThreadDataMiddleware()

	if m.Name() != "thread_data" {
		t.Errorf("Name() = %v, want 'thread_data'", m.Name())
	}
}

func TestThreadDataMiddleware_BeforeAgent_EmptyThreadID(t *testing.T) {
	m := NewDefaultThreadDataMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	// threadID 为空（DeerFlow 行为：返回错误）
	stateUpdate, err := m.BeforeAgent(ctx, ts, "")

	if err == nil {
		t.Error("BeforeAgent() error = nil, want non-nil")
	}
	if err != ErrThreadIDRequired {
		t.Errorf("BeforeAgent() error = %v, want ErrThreadIDRequired", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeAgent() stateUpdate = %v, want nil", stateUpdate)
	}
	if ts.ThreadData != nil {
		t.Error("ThreadData should be nil when threadID is empty")
	}
}

func TestThreadDataMiddleware_BeforeAgent_LazyInit(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()
	m := NewThreadDataMiddleware(tmpDir, true)
	ctx := context.Background()
	ts := state.NewThreadState()
	threadID := "test-thread-123"

	stateUpdate, err := m.BeforeAgent(ctx, ts, threadID)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Error("stateUpdate should not be nil")
	}

	// 验证 ThreadData 已设置
	if ts.ThreadData == nil {
		t.Fatal("ThreadData should not be nil")
	}

	// 验证路径
	paths := config.NewPaths(tmpDir)
	expectedWorkspace := paths.SandboxWorkDir(threadID)
	expectedUploads := paths.SandboxUploadsDir(threadID)
	expectedOutputs := paths.SandboxOutputsDir(threadID)

	if ts.ThreadData.WorkspacePath != expectedWorkspace {
		t.Errorf("WorkspacePath = %v, want %v", ts.ThreadData.WorkspacePath, expectedWorkspace)
	}
	if ts.ThreadData.UploadsPath != expectedUploads {
		t.Errorf("UploadsPath = %v, want %v", ts.ThreadData.UploadsPath, expectedUploads)
	}
	if ts.ThreadData.OutputsPath != expectedOutputs {
		t.Errorf("OutputsPath = %v, want %v", ts.ThreadData.OutputsPath, expectedOutputs)
	}

	// 验证 LazyInit: 目录不应被创建
	workspaceDir := paths.SandboxWorkDir(threadID)
	if _, err := os.Stat(workspaceDir); !os.IsNotExist(err) {
		t.Error("Workspace directory should not exist with LazyInit=true")
	}
}

func TestThreadDataMiddleware_BeforeAgent_EagerInit(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()
	m := NewThreadDataMiddleware(tmpDir, false)
	ctx := context.Background()
	ts := state.NewThreadState()
	threadID := "test-thread-456"

	stateUpdate, err := m.BeforeAgent(ctx, ts, threadID)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Error("stateUpdate should not be nil")
	}

	// 验证 ThreadData 已设置
	if ts.ThreadData == nil {
		t.Fatal("ThreadData should not be nil")
	}

	// 验证 EagerInit: 目录应该被创建
	paths := config.NewPaths(tmpDir)
	workspaceDir := paths.SandboxWorkDir(threadID)
	uploadsDir := paths.SandboxUploadsDir(threadID)
	outputsDir := paths.SandboxOutputsDir(threadID)

	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		t.Error("Workspace directory should exist with LazyInit=false")
	}
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		t.Error("Uploads directory should exist with LazyInit=false")
	}
	if _, err := os.Stat(outputsDir); os.IsNotExist(err) {
		t.Error("Outputs directory should exist with LazyInit=false")
	}
}

func TestThreadDataMiddleware_BeforeAgent_StateUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewThreadDataMiddleware(tmpDir, true)
	ctx := context.Background()
	ts := state.NewThreadState()
	threadID := "test-thread-789"

	stateUpdate, err := m.BeforeAgent(ctx, ts, threadID)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}

	// 验证 stateUpdate 包含 thread_data
	threadData, ok := stateUpdate["thread_data"]
	if !ok {
		t.Fatal("stateUpdate should contain 'thread_data'")
	}

	td, ok := threadData.(*state.ThreadDataState)
	if !ok {
		t.Fatal("thread_data should be *state.ThreadDataState")
	}

	paths := config.NewPaths(tmpDir)
	if td.WorkspacePath != paths.SandboxWorkDir(threadID) {
		t.Error("stateUpdate thread_data has wrong WorkspacePath")
	}
}

func TestThreadDataMiddleware_DefaultBaseDir(t *testing.T) {
	// 不指定 baseDir，使用默认
	m := NewDefaultThreadDataMiddleware()

	if m.baseDir != "" {
		t.Errorf("baseDir = %v, want empty string (uses default)", m.baseDir)
	}
	if !m.lazyInit {
		t.Error("lazyInit should be true by default")
	}
}

func TestThreadDataMiddleware_getThreadPaths(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewThreadDataMiddleware(tmpDir, true)
	threadID := "test-thread-paths"

	paths := m.getThreadPaths(threadID)

	// 验证返回的 map 包含所有必需的键
	requiredKeys := []string{"workspace_path", "uploads_path", "outputs_path"}
	for _, key := range requiredKeys {
		if _, ok := paths[key]; !ok {
			t.Errorf("getThreadPaths() missing key: %v", key)
		}
	}

	// 验证路径格式
	configPaths := config.NewPaths(tmpDir)
	if paths["workspace_path"] != configPaths.SandboxWorkDir(threadID) {
		t.Error("workspace_path mismatch")
	}
}

func TestThreadDataMiddleware_createThreadDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewThreadDataMiddleware(tmpDir, true)
	threadID := "test-thread-create"

	paths := m.createThreadDirectories(threadID)

	// 验证目录已创建
	configPaths := config.NewPaths(tmpDir)
	workspaceDir := configPaths.SandboxWorkDir(threadID)
	uploadsDir := configPaths.SandboxUploadsDir(threadID)
	outputsDir := configPaths.SandboxOutputsDir(threadID)

	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		t.Error("workspace directory not created")
	}
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		t.Error("uploads directory not created")
	}
	if _, err := os.Stat(outputsDir); os.IsNotExist(err) {
		t.Error("outputs directory not created")
	}

	// 验证返回的路径
	if paths["workspace_path"] != workspaceDir {
		t.Error("createThreadDirectories returned wrong workspace_path")
	}
}

func TestThreadDataMiddleware_MultipleCalls(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewThreadDataMiddleware(tmpDir, false)
	ctx := context.Background()
	threadID := "test-thread-multi"

	// 第一次调用
	ts1 := state.NewThreadState()
	_, _ = m.BeforeAgent(ctx, ts1, threadID)

	// 第二次调用（同一 threadID）
	ts2 := state.NewThreadState()
	_, _ = m.BeforeAgent(ctx, ts2, threadID)

	// 验证两次结果一致
	if ts1.ThreadData.WorkspacePath != ts2.ThreadData.WorkspacePath {
		t.Error("Multiple calls should return same paths")
	}

	// 验证目录只创建一次（不报错）
	workspaceDir := config.NewPaths(tmpDir).SandboxWorkDir(threadID)
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		t.Error("Directory should still exist after multiple calls")
	}
}

func TestThreadDataMiddleware_DifferentThreadIDs(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewThreadDataMiddleware(tmpDir, true)
	ctx := context.Background()

	// 线程 1
	ts1 := state.NewThreadState()
	_, _ = m.BeforeAgent(ctx, ts1, "thread-1")

	// 线程 2
	ts2 := state.NewThreadState()
	_, _ = m.BeforeAgent(ctx, ts2, "thread-2")

	// 验证路径不同
	if ts1.ThreadData.WorkspacePath == ts2.ThreadData.WorkspacePath {
		t.Error("Different thread IDs should have different paths")
	}

	// 验证路径包含各自的 threadID
	paths := config.NewPaths(tmpDir)
	if ts1.ThreadData.WorkspacePath != paths.SandboxWorkDir("thread-1") {
		t.Error("Thread 1 path incorrect")
	}
	if ts2.ThreadData.WorkspacePath != paths.SandboxWorkDir("thread-2") {
		t.Error("Thread 2 path incorrect")
	}
}

func TestThreadDataMiddleware_PathPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewThreadDataMiddleware(tmpDir, false)
	ctx := context.Background()
	ts := state.NewThreadState()
	threadID := "test-thread-perms"

	_, _ = m.BeforeAgent(ctx, ts, threadID)

	// 验证目录权限（0755）
	paths := config.NewPaths(tmpDir)
	workspaceDir := paths.SandboxWorkDir(threadID)

	info, err := os.Stat(workspaceDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0755 {
		t.Errorf("Directory permissions = %v, want 0755", mode)
	}
}

func TestThreadDataMiddleware_IntegrationWithChain(t *testing.T) {
	// 验证 ThreadDataMiddleware 能正确加入中间件链
	tmpDir := t.TempDir()
	config := DefaultMiddlewareConfig()
	config.BaseDir = tmpDir
	config.LazyInit = true

	chain := BuildLeadAgentMiddlewares(config)

	// 验证链中有 ThreadDataMiddleware
	found := false
	for _, m := range chain.Middlewares() {
		if m.Name() == "thread_data" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ThreadDataMiddleware not found in chain")
	}

	// 验证是第一个中间件
	if len(chain.Middlewares()) == 0 {
		t.Fatal("Chain should not be empty")
	}
	if chain.Middlewares()[0].Name() != "thread_data" {
		t.Error("ThreadDataMiddleware should be first in chain")
	}
}

// BenchmarkThreadDataMiddleware_BeforeAgent 性能基准测试
func BenchmarkThreadDataMiddleware_BeforeAgent(b *testing.B) {
	tmpDir := b.TempDir()
	m := NewThreadDataMiddleware(tmpDir, true)
	ctx := context.Background()
	threadID := "bench-thread"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts := state.NewThreadState()
		_, _ = m.BeforeAgent(ctx, ts, threadID)
	}
}

// BenchmarkThreadDataMiddleware_BeforeAgent_Eager 性能基准测试（EagerInit）
func BenchmarkThreadDataMiddleware_BeforeAgent_Eager(b *testing.B) {
	tmpDir := b.TempDir()
	m := NewThreadDataMiddleware(tmpDir, false)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		threadID := filepath.Join("bench-thread-eager", string(rune('a'+i%26)))
		ts := state.NewThreadState()
		_, _ = m.BeforeAgent(ctx, ts, threadID)
	}
}
