package middleware

import (
	"testing"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestBaseMiddleware(t *testing.T) {
	m := NewBaseMiddleware("test")

	if m.Name() != "test" {
		t.Errorf("Name() = %v, want 'test'", m.Name())
	}
}

func TestMiddlewareChain(t *testing.T) {
	chain := NewMiddlewareChain()

	// 添加中间件
	m1 := NewBaseMiddleware("m1")
	m2 := NewBaseMiddleware("m2")
	chain.Add(m1).Add(m2)

	if len(chain.Middlewares()) != 2 {
		t.Errorf("Middlewares() len = %v, want 2", len(chain.Middlewares()))
	}

	// 测试 AddAll
	chain = NewMiddlewareChain()
	chain.AddAll(m1, m2)
	if len(chain.Middlewares()) != 2 {
		t.Errorf("AddAll() len = %v, want 2", len(chain.Middlewares()))
	}
}

func TestLoopDetectionMiddleware(t *testing.T) {
	logger := zap.NewNop()
	m := NewLoopDetectionMiddleware(logger)

	if m.Name() != "loop_detection" {
		t.Errorf("Name() = %v, want 'loop_detection'", m.Name())
	}
}

func TestEinoCallbackBridge(t *testing.T) {
	chain := NewMiddlewareChain()
	s := state.NewThreadState()
	bridge := NewEinoCallbackBridge(chain, s)

	if bridge.GetState() != s {
		t.Error("GetState() should return initial state")
	}

	// 测试 Handler() 不返回 nil
	if bridge.Handler() == nil {
		t.Error("Handler() should not return nil")
	}
}
