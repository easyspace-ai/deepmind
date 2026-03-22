package middleware

import (
	"context"
	"testing"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/sandbox"
	"go.uber.org/zap"
)

// ============================================
// Mock Sandbox and SandboxProvider for testing
// ============================================

type mockSandbox struct {
	id string
}

func (m *mockSandbox) ID() string {
	return m.id
}

func (m *mockSandbox) ExecuteCommand(command string) (string, error) {
	return "output: " + command, nil
}

func (m *mockSandbox) ReadFile(path string) (string, error) {
	return "content of " + path, nil
}

func (m *mockSandbox) WriteFile(path, content string, append bool) error {
	return nil
}

func (m *mockSandbox) ListDir(path string, maxDepth int) ([]string, error) {
	return []string{"file1.txt", "file2.txt"}, nil
}

func (m *mockSandbox) UpdateFile(path string, content []byte) error {
	return nil
}

type mockSandboxProvider struct {
	sandboxes map[string]sandbox.Sandbox
	acquireCount int
	releaseCount int
}

func newMockSandboxProvider() *mockSandboxProvider {
	return &mockSandboxProvider{
		sandboxes: make(map[string]sandbox.Sandbox),
	}
}

func (m *mockSandboxProvider) Acquire(threadID string) (sandbox.Sandbox, error) {
	m.acquireCount++
	sb := &mockSandbox{id: "sandbox-" + threadID}
	m.sandboxes[threadID] = sb
	return sb, nil
}

func (m *mockSandboxProvider) Get(threadID string) (sandbox.Sandbox, bool) {
	sb, ok := m.sandboxes[threadID]
	return sb, ok
}

func (m *mockSandboxProvider) Release(threadID string) error {
	m.releaseCount++
	delete(m.sandboxes, threadID)
	return nil
}

// ============================================
// Tests
// ============================================

func TestSandboxMiddleware_Name(t *testing.T) {
	m := NewDefaultSandboxMiddleware()

	if m.Name() != "sandbox" {
		t.Errorf("Name() = %v, want 'sandbox'", m.Name())
	}
}

func TestSandboxMiddleware_BeforeAgent_EmptyThreadID(t *testing.T) {
	m := NewDefaultSandboxMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	// threadID 为空
	stateUpdate, err := m.BeforeAgent(ctx, ts, "")

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeAgent() stateUpdate = %v, want nil", stateUpdate)
	}
	if ts.Sandbox != nil {
		t.Error("Sandbox should be nil when threadID is empty")
	}
}

func TestSandboxMiddleware_BeforeAgent_NoProvider(t *testing.T) {
	m := NewDefaultSandboxMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	threadID := "thread-no-provider"
	stateUpdate, err := m.BeforeAgent(ctx, ts, threadID)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Fatal("stateUpdate should not be nil")
	}

	// 验证使用本地沙箱占位符
	if ts.Sandbox == nil {
		t.Fatal("Sandbox should not be nil")
	}
	if ts.Sandbox.SandboxID != "local" {
		t.Errorf("SandboxID = %v, want 'local'", ts.Sandbox.SandboxID)
	}

	// 验证 stateUpdate
	sandboxState, ok := stateUpdate["sandbox"]
	if !ok {
		t.Error("stateUpdate should contain 'sandbox'")
	}
	if sandboxState == nil {
		t.Error("sandbox in stateUpdate should not be nil")
	}
}

func TestSandboxMiddleware_BeforeAgent_WithProvider(t *testing.T) {
	provider := newMockSandboxProvider()
	logger := zap.NewNop()
	m := NewSandboxMiddleware(provider, logger)
	ctx := context.Background()
	ts := state.NewThreadState()

	threadID := "thread-with-provider"
	stateUpdate, err := m.BeforeAgent(ctx, ts, threadID)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Fatal("stateUpdate should not be nil")
	}

	// 验证 provider 的 Acquire 被调用
	if provider.acquireCount != 1 {
		t.Errorf("acquireCount = %v, want 1", provider.acquireCount)
	}

	// 验证沙箱已创建
	if ts.Sandbox == nil {
		t.Fatal("Sandbox should not be nil")
	}
	expectedSandboxID := "sandbox-" + threadID
	if ts.Sandbox.SandboxID != expectedSandboxID {
		t.Errorf("SandboxID = %v, want %v", ts.Sandbox.SandboxID, expectedSandboxID)
	}

	// 验证 Get 能找到沙箱
	sb, exists := provider.Get(threadID)
	if !exists {
		t.Error("Sandbox should exist in provider")
	}
	if sb.ID() != expectedSandboxID {
		t.Errorf("Sandbox ID from Get() = %v, want %v", sb.ID(), expectedSandboxID)
	}
}

func TestSandboxMiddleware_BeforeAgent_AlreadyExists(t *testing.T) {
	provider := newMockSandboxProvider()
	logger := zap.NewNop()
	m := NewSandboxMiddleware(provider, logger)
	ctx := context.Background()
	ts := state.NewThreadState()

	threadID := "thread-already-exists"

	// 第一次调用
	_, _ = m.BeforeAgent(ctx, ts, threadID)
	firstAcquireCount := provider.acquireCount

	// 第二次调用（应该不重复 acquire）
	ts2 := state.NewThreadState()
	ts2.Sandbox = &state.SandboxState{SandboxID: "sandbox-" + threadID}
	stateUpdate, err := m.BeforeAgent(ctx, ts2, threadID)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}

	// 验证没有重复 acquire
	if provider.acquireCount != firstAcquireCount {
		t.Errorf("acquireCount = %v, want %v (no duplicate acquire)", provider.acquireCount, firstAcquireCount)
	}

	// stateUpdate 应该是 nil（不需要更新）
	if stateUpdate != nil {
		t.Errorf("stateUpdate = %v, want nil when sandbox already exists", stateUpdate)
	}
}

func TestSandboxMiddleware_BeforeAgent_Idempotent(t *testing.T) {
	provider := newMockSandboxProvider()
	logger := zap.NewNop()
	m := NewSandboxMiddleware(provider, logger)
	ctx := context.Background()
	threadID := "thread-idempotent"

	// 使用同一个 state 多次调用
	ts := state.NewThreadState()

	// 第一次调用
	_, _ = m.BeforeAgent(ctx, ts, threadID)
	firstAcquireCount := provider.acquireCount

	// 第二次调用（同一个 state）
	_, _ = m.BeforeAgent(ctx, ts, threadID)

	// Acquire 应该只被调用一次
	if provider.acquireCount != firstAcquireCount {
		t.Errorf("acquireCount = %v, want %v (idempotent, same state)", provider.acquireCount, firstAcquireCount)
	}
}

func TestSandboxMiddleware_BeforeAgent_MultipleThreads(t *testing.T) {
	provider := newMockSandboxProvider()
	logger := zap.NewNop()
	m := NewSandboxMiddleware(provider, logger)
	ctx := context.Background()

	// 线程 1
	ts1 := state.NewThreadState()
	threadID1 := "thread-1"
	_, _ = m.BeforeAgent(ctx, ts1, threadID1)

	// 线程 2
	ts2 := state.NewThreadState()
	threadID2 := "thread-2"
	_, _ = m.BeforeAgent(ctx, ts2, threadID2)

	// 验证两个沙箱都创建了
	if provider.acquireCount != 2 {
		t.Errorf("acquireCount = %v, want 2", provider.acquireCount)
	}

	// 验证沙箱 ID 不同
	if ts1.Sandbox.SandboxID == ts2.Sandbox.SandboxID {
		t.Error("Sandbox IDs should be different for different threads")
	}
}

func TestSandboxMiddleware_IntegrationWithChain(t *testing.T) {
	// 验证 SandboxMiddleware 能正确加入中间件链
	provider := newMockSandboxProvider()
	config := DefaultMiddlewareConfig()
	config.SandboxProvider = provider

	chain := BuildLeadAgentMiddlewares(config)

	// 验证链中有 SandboxMiddleware
	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "sandbox" {
			found = true
			break
		}
	}
	if !found {
		t.Error("SandboxMiddleware not found in chain")
	}

	// 验证位置：应该是第三个（在 ThreadDataMiddleware 和 UploadsMiddleware 之后）
	if len(chain.Middlewares()) < 3 {
		t.Fatal("Chain should have at least 3 middlewares")
	}
	if chain.Middlewares()[2].Name() != "sandbox" {
		t.Error("SandboxMiddleware should be third in chain")
	}
}

func TestSandboxMiddleware_DefaultSandboxID(t *testing.T) {
	m := NewDefaultSandboxMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	threadID := "thread-default"
	_, _ = m.BeforeAgent(ctx, ts, threadID)

	if ts.Sandbox == nil {
		t.Fatal("Sandbox should not be nil")
	}
	if ts.Sandbox.SandboxID != "local" {
		t.Errorf("Default SandboxID = %v, want 'local'", ts.Sandbox.SandboxID)
	}
}

func TestSandboxMiddleware_Logger(t *testing.T) {
	// 验证 logger 参数可以为 nil
	provider := newMockSandboxProvider()

	// nil logger 应该使用 NopLogger
	m := NewSandboxMiddleware(provider, nil)
	if m.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}

	// 非 nil logger 应该被使用
	customLogger := zap.NewExample()
	m2 := NewSandboxMiddleware(provider, customLogger)
	if m2.logger != customLogger {
		t.Error("custom logger should be used")
	}
}

// BenchmarkSandboxMiddleware_BeforeAgent 性能基准测试
func BenchmarkSandboxMiddleware_BeforeAgent(b *testing.B) {
	provider := newMockSandboxProvider()
	logger := zap.NewNop()
	m := NewSandboxMiddleware(provider, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts := state.NewThreadState()
		threadID := "bench-thread"
		_, _ = m.BeforeAgent(ctx, ts, threadID)
	}
}
