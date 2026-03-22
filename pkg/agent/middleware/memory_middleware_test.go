package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestMemoryMiddleware_Name(t *testing.T) {
	m := NewDefaultMemoryMiddleware()

	if m.Name() != "memory" {
		t.Errorf("Name() = %v, want 'memory'", m.Name())
	}
}

func TestMemoryMiddleware_NewDefault(t *testing.T) {
	m := NewDefaultMemoryMiddleware()

	if m.enabled {
		t.Error("enabled = true, want false by default")
	}
}

func TestMemoryMiddleware_NewWithConfig(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewMemoryMiddleware(true, customLogger)

	if !m.enabled {
		t.Error("enabled = false, want true")
	}
	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}
	if m.manager == nil {
		t.Error("manager should not be nil")
	}
}

func TestMemoryMiddleware_AfterModel_Disabled(t *testing.T) {
	m := NewMemoryMiddleware(false, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestMemoryMiddleware_AfterModel_NilState(t *testing.T) {
	m := NewMemoryMiddleware(true, nil)
	ctx := context.Background()

	stateUpdate, err := m.AfterModel(ctx, nil)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestMemoryMiddleware_AfterModel_NilManager(t *testing.T) {
	m := NewMemoryMiddleware(true, nil)
	m.manager = nil // 手动设置为 nil
	ctx := context.Background()
	ts := state.NewThreadState()

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestMemoryMiddleware_AfterModel_Enabled(t *testing.T) {
	m := NewMemoryMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil (no state update expected)", stateUpdate)
	}
}

func TestMemoryMiddleware_AfterModel_WithMessages(t *testing.T) {
	m := NewMemoryMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 添加一些消息
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "Hello"},
		{Role: schema.Assistant, Content: "Hi"},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestMemoryMiddleware_Logger(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewMemoryMiddleware(true, customLogger)

	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}

	m2 := NewMemoryMiddleware(true, nil)
	if m2.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}
}

func TestMemoryMiddleware_IntegrationWithChain(t *testing.T) {
	config := DefaultMiddlewareConfig()
	config.MemoryEnabled = true // 需要启用
	chain := BuildLeadAgentMiddlewares(config)

	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "memory" {
			found = true
			break
		}
	}
	if !found {
		t.Error("MemoryMiddleware not found in chain (need MemoryEnabled=true)")
	}
}

// BenchmarkMemoryMiddleware_AfterModel 性能基准测试
func BenchmarkMemoryMiddleware_AfterModel(b *testing.B) {
	m := NewMemoryMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "Hello"},
		{Role: schema.Assistant, Content: "Hi"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.AfterModel(ctx, ts)
	}
}
