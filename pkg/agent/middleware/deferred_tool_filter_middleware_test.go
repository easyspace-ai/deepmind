package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestDeferredToolFilterMiddleware_Name(t *testing.T) {
	m := NewDefaultDeferredToolFilterMiddleware()

	if m.Name() != "deferred_tool_filter" {
		t.Errorf("Name() = %v, want 'deferred_tool_filter'", m.Name())
	}
}

func TestDeferredToolFilterMiddleware_NewDefault(t *testing.T) {
	m := NewDefaultDeferredToolFilterMiddleware()

	if len(m.deferredNames) != 0 {
		t.Errorf("deferredNames length = %v, want 0", len(m.deferredNames))
	}
}

func TestDeferredToolFilterMiddleware_NewWithNames(t *testing.T) {
	names := []string{"tool1", "tool2", "tool3"}
	customLogger := zap.NewExample()
	m := NewDeferredToolFilterMiddlewareWithNames(names, customLogger)

	if len(m.deferredNames) != 3 {
		t.Errorf("deferredNames length = %v, want 3", len(m.deferredNames))
	}
	if m.deferredNames[0] != "tool1" {
		t.Error("first tool name mismatch")
	}
	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}
}

func TestDeferredToolFilterMiddleware_NewWithNames_CopiesSlice(t *testing.T) {
	names := []string{"tool1", "tool2"}
	m := NewDeferredToolFilterMiddlewareWithNames(names, nil)

	// 修改原切片，不应影响 middleware
	names[0] = "modified"

	if m.deferredNames[0] != "tool1" {
		t.Error("middleware should copy the slice, not reference it")
	}
}

func TestDeferredToolFilterMiddleware_NewWithNames_NilSlice(t *testing.T) {
	m := NewDeferredToolFilterMiddlewareWithNames(nil, nil)

	if len(m.deferredNames) != 0 {
		t.Errorf("deferredNames length = %v, want 0", len(m.deferredNames))
	}
}

func TestDeferredToolFilterMiddleware_BeforeModel_NoDeferredTools(t *testing.T) {
	m := NewDefaultDeferredToolFilterMiddleware()
	ctx := context.Background()
	ts := state.NewThreadState()

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestDeferredToolFilterMiddleware_BeforeModel_WithDeferredTools(t *testing.T) {
	m := NewDeferredToolFilterMiddlewareWithNames([]string{"tool1", "tool2"}, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 添加一些消息
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "Hello"},
	}

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeModel() stateUpdate = %v, want nil (no state update expected)", stateUpdate)
	}
}

func TestDeferredToolFilterMiddleware_Logger(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewDeferredToolFilterMiddlewareWithNames([]string{"tool1"}, customLogger)

	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}

	m2 := NewDeferredToolFilterMiddleware(nil)
	if m2.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}
}

func TestDeferredToolFilterMiddleware_IntegrationWithChain(t *testing.T) {
	config := DefaultMiddlewareConfig()
	chain := BuildLeadAgentMiddlewares(config)

	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "deferred_tool_filter" {
			found = true
			break
		}
	}
	if !found {
		t.Error("DeferredToolFilterMiddleware not found in chain")
	}
}

// BenchmarkDeferredToolFilterMiddleware_BeforeModel 性能基准测试
func BenchmarkDeferredToolFilterMiddleware_BeforeModel(b *testing.B) {
	m := NewDeferredToolFilterMiddlewareWithNames([]string{"tool1", "tool2", "tool3"}, nil)
	ctx := context.Background()
	ts := state.NewThreadState()
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "Hello"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.BeforeModel(ctx, ts)
	}
}
