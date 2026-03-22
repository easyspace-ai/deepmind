package middleware

import (
	"context"
	"testing"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestTodoListMiddleware_Name(t *testing.T) {
	m := NewDefaultTodoListMiddleware()

	if m.Name() != "todo_list" {
		t.Errorf("Name() = %v, want 'todo_list'", m.Name())
	}
}

func TestTodoListMiddleware_NewDefault(t *testing.T) {
	m := NewDefaultTodoListMiddleware()

	if m.enabled {
		t.Error("enabled = true, want false by default")
	}
}

func TestTodoListMiddleware_NewWithConfig(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewTodoListMiddleware(true, customLogger)

	if !m.enabled {
		t.Error("enabled = false, want true")
	}
	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}
}

func TestTodoListMiddleware_AfterModel_Disabled(t *testing.T) {
	m := NewTodoListMiddleware(false, nil)
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

func TestTodoListMiddleware_AfterModel_NilState(t *testing.T) {
	m := NewTodoListMiddleware(true, nil)
	ctx := context.Background()

	stateUpdate, err := m.AfterModel(ctx, nil)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestTodoListMiddleware_AfterModel_Enabled(t *testing.T) {
	m := NewTodoListMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 添加一些 todos
	ts.Todos = []state.TodoItem{
		{ID: "1", Description: "Task 1", Status: "pending"},
		{ID: "2", Description: "Task 2", Status: "completed"},
	}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil (no state update expected)", stateUpdate)
	}
}

func TestTodoListMiddleware_AfterModel_EmptyTodos(t *testing.T) {
	m := NewTodoListMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 空 todos
	ts.Todos = []state.TodoItem{}

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestTodoListMiddleware_AfterModel_NilTodos(t *testing.T) {
	m := NewTodoListMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	// nil todos
	ts.Todos = nil

	stateUpdate, err := m.AfterModel(ctx, ts)

	if err != nil {
		t.Errorf("AfterModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("AfterModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestTodoListMiddleware_Logger(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewTodoListMiddleware(true, customLogger)

	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}

	m2 := NewTodoListMiddleware(true, nil)
	if m2.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}
}

func TestTodoListMiddleware_IntegrationWithChain(t *testing.T) {
	config := DefaultMiddlewareConfig()
	config.TodoListEnabled = true // 需要启用
	chain := BuildLeadAgentMiddlewares(config)

	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "todo_list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("TodoListMiddleware not found in chain (need TodoListEnabled=true)")
	}
}

// BenchmarkTodoListMiddleware_AfterModel 性能基准测试
func BenchmarkTodoListMiddleware_AfterModel(b *testing.B) {
	m := NewTodoListMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()
	ts.Todos = []state.TodoItem{
		{ID: "1", Description: "Task 1", Status: "pending"},
		{ID: "2", Description: "Task 2", Status: "completed"},
		{ID: "3", Description: "Task 3", Status: "pending"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.AfterModel(ctx, ts)
	}
}
