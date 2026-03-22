package subagent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// ============================================
// SubagentStatus 测试
// ============================================

func TestSubagentStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status SubagentStatus
		expect bool
	}{
		{"Pending", SubagentStatusPending, false},
		{"Running", SubagentStatusRunning, false},
		{"Completed", SubagentStatusCompleted, true},
		{"Failed", SubagentStatusFailed, true},
		{"TimedOut", SubagentStatusTimedOut, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.status.IsTerminal())
		})
	}
}

// ============================================
// SubagentResult 测试
// ============================================

func TestNewSubagentResult(t *testing.T) {
	result := NewSubagentResult("task-123", "trace-456")

	assert.Equal(t, "task-123", result.TaskID)
	assert.Equal(t, "trace-456", result.TraceID)
	assert.Equal(t, SubagentStatusPending, result.Status)
	assert.NotNil(t, result.AIMessages)
	assert.Empty(t, result.AIMessages)
}

// ============================================
// SubagentConfig 测试
// ============================================

func TestDefaultSubagentConfig(t *testing.T) {
	config := DefaultSubagentConfig()

	assert.Equal(t, "inherit", config.Model)
	assert.Equal(t, 50, config.MaxTurns)
	assert.Equal(t, 900, config.TimeoutSeconds)
	assert.Contains(t, config.DisallowedTools, "task")
}

// ============================================
// FilterTools 测试
// ============================================

func TestFilterTools(t *testing.T) {
	allTools := []string{"bash", "ls", "read_file", "write_file", "task", "ask_clarification"}

	tests := []struct {
		name       string
		allowed    []string
		disallowed []string
		expect     []string
	}{
		{
			"No filters",
			nil,
			nil,
			[]string{"bash", "ls", "read_file", "write_file", "task", "ask_clarification"},
		},
		{
			"Allowlist only",
			[]string{"bash", "ls"},
			nil,
			[]string{"bash", "ls"},
		},
		{
			"Denylist only",
			nil,
			[]string{"task", "ask_clarification"},
			[]string{"bash", "ls", "read_file", "write_file"},
		},
		{
			"Both filters",
			[]string{"bash", "ls", "task"},
			[]string{"task"},
			[]string{"bash", "ls"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterTools(allTools, tt.allowed, tt.disallowed)
			assert.ElementsMatch(t, tt.expect, result)
		})
	}
}

// ============================================
// ResolveModelName 测试
// ============================================

func TestResolveModelName(t *testing.T) {
	tests := []struct {
		name        string
		configModel string
		parentModel string
		expect      string
	}{
		{"Inherit", "inherit", "gpt-4", "gpt-4"},
		{"Specific model", "claude-3", "gpt-4", "claude-3"},
		{"Empty inherit", "inherit", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SubagentConfig{Model: tt.configModel}
			result := ResolveModelName(config, tt.parentModel)
			assert.Equal(t, tt.expect, result)
		})
	}
}

// ============================================
// 全局任务存储测试
// ============================================

func TestBackgroundTaskStorage(t *testing.T) {
	// 清理可能存在的任务
	backgroundTasksLock.Lock()
	clear(backgroundTasks)
	backgroundTasksLock.Unlock()

	// 创建测试任务
	result1 := NewSubagentResult("task-1", "trace-1")
	result2 := NewSubagentResult("task-2", "trace-2")

	backgroundTasksLock.Lock()
	backgroundTasks["task-1"] = result1
	backgroundTasks["task-2"] = result2
	backgroundTasksLock.Unlock()

	// 测试 GetBackgroundTaskResult
	t.Run("GetBackgroundTaskResult", func(t *testing.T) {
		found := GetBackgroundTaskResult("task-1")
		assert.Equal(t, result1.TaskID, found.TaskID)

		notFound := GetBackgroundTaskResult("non-existent")
		assert.Nil(t, notFound)
	})

	// 测试 ListBackgroundTasks
	t.Run("ListBackgroundTasks", func(t *testing.T) {
		list := ListBackgroundTasks()
		assert.Len(t, list, 2)
	})

	// 测试 CleanupBackgroundTask
	t.Run("CleanupBackgroundTask", func(t *testing.T) {
		logger := zap.NewNop()

		// 标记 task-1 为完成
		backgroundTasksLock.Lock()
		now := time.Now()
		backgroundTasks["task-1"].Status = SubagentStatusCompleted
		backgroundTasks["task-1"].CompletedAt = &now
		backgroundTasksLock.Unlock()

		// 清理已完成的任务
		CleanupBackgroundTask("task-1", logger)

		// task-1 应该被删除
		assert.Nil(t, GetBackgroundTaskResult("task-1"))

		// task-2 还在运行中，不应该被删除
		assert.NotNil(t, GetBackgroundTaskResult("task-2"))
	})
}

// ============================================
// SubagentRegistry 测试
// ============================================

func TestSubagentRegistry(t *testing.T) {
	logger := zap.NewNop()
	registry := NewSubagentRegistry(logger)

	// 测试内置子代理已注册
	t.Run("Builtin subagents registered", func(t *testing.T) {
		names := registry.Names()
		assert.Contains(t, names, "general-purpose")
		assert.Contains(t, names, "bash")
	})

	// 测试获取子代理配置
	t.Run("Get subagent config", func(t *testing.T) {
		config := registry.Get("general-purpose")
		assert.NotNil(t, config)
		assert.Equal(t, "general-purpose", config.Name)

		notFound := registry.Get("non-existent")
		assert.Nil(t, notFound)
	})

	// 测试列出子代理
	t.Run("List subagents", func(t *testing.T) {
		list := registry.List()
		assert.GreaterOrEqual(t, len(list), 2)
	})

	// 测试超时覆盖
	t.Run("Timeout override", func(t *testing.T) {
		registry.SetTimeoutOverride("bash", 300)

		config := registry.Get("bash")
		assert.Equal(t, 300, config.TimeoutSeconds)

		registry.ClearTimeoutOverride("bash")
		config = registry.Get("bash")
		assert.Equal(t, 900, config.TimeoutSeconds)
	})

	// 测试注册新子代理
	t.Run("Register new subagent", func(t *testing.T) {
		customConfig := &SubagentConfig{
			Name:        "custom-agent",
			Description: "Custom agent",
		}
		registry.Register("custom-agent", customConfig)

		found := registry.Get("custom-agent")
		assert.NotNil(t, found)
		assert.Equal(t, "Custom agent", found.Description)
	})
}

// ============================================
// 全局注册表测试
// ============================================

func TestGlobalRegistry(t *testing.T) {
	// 测试便捷函数
	t.Run("Convenience functions", func(t *testing.T) {
		config := GetSubagentConfig("general-purpose")
		assert.NotNil(t, config)

		list := ListSubagents()
		assert.GreaterOrEqual(t, len(list), 2)

		names := GetSubagentNames()
		assert.Contains(t, names, "general-purpose")
	})
}

// ============================================
// 事件系统测试
// ============================================

func TestTaskEventBus(t *testing.T) {
	bus := NewTaskEventBus()

	var receivedEvent TaskEvent
	handler := func(event TaskEvent) {
		receivedEvent = event
	}

	// 测试订阅和发布
	t.Run("Subscribe and publish", func(t *testing.T) {
		bus.Subscribe("task-1", handler)

		event := NewTaskStartedEvent("task-1", "Test task")
		bus.Publish(event)

		assert.Equal(t, TaskEventTypeStarted, receivedEvent.Type)
		assert.Equal(t, "task-1", receivedEvent.TaskID)
		assert.Equal(t, "Test task", receivedEvent.Description)
	})

	// 测试清除订阅
	t.Run("Clear subscription", func(t *testing.T) {
		bus.Clear("task-1")

		receivedEvent = TaskEvent{} // 重置
		event := NewTaskCompletedEvent("task-1", "Done")
		bus.Publish(event)

		assert.Empty(t, receivedEvent.Type) // 不应该收到事件
	})
}

func TestEventCreation(t *testing.T) {
	tests := []struct {
		name     string
		event    TaskEvent
		expected TaskEventType
	}{
		{
			"Started",
			NewTaskStartedEvent("t1", "desc"),
			TaskEventTypeStarted,
		},
		{
			"Running",
			NewTaskRunningEvent("t1", map[string]any{"text": "hello"}, 1, 1),
			TaskEventTypeRunning,
		},
		{
			"Completed",
			NewTaskCompletedEvent("t1", "result"),
			TaskEventTypeCompleted,
		},
		{
			"Failed",
			NewTaskFailedEvent("t1", "error"),
			TaskEventTypeFailed,
		},
		{
			"TimedOut",
			NewTaskTimedOutEvent("t1", "timeout"),
			TaskEventTypeTimedOut,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.Type)
			assert.Equal(t, "t1", tt.event.TaskID)
		})
	}
}

// ============================================
// WorkerPool 测试
// ============================================

func TestWorkerPool(t *testing.T) {
	pool := NewWorkerPool(2, "test-pool")
	pool.SetLogger(zap.NewNop())
	defer pool.Stop()

	// 测试任务执行
	t.Run("Execute tasks", func(t *testing.T) {
		resultChan := make(chan int, 3)

		for i := 0; i < 3; i++ {
			val := i
			pool.Submit(func() {
				resultChan <- val
			})
		}

		// 收集结果
		results := make(map[int]bool)
		for i := 0; i < 3; i++ {
			select {
			case val := <-resultChan:
				results[val] = true
			case <-time.After(2 * time.Second):
				t.Fatal("Timeout waiting for task")
			}
		}

		assert.True(t, results[0])
		assert.True(t, results[1])
		assert.True(t, results[2])
	})
}

// ============================================
// SubagentExecutor 测试
// ============================================

func TestSubagentExecutor(t *testing.T) {
	logger := zap.NewNop()

	config := &SubagentConfig{
		Name:           "test-agent",
		SystemPrompt:   "You are a test agent",
		MaxTurns:       10,
		TimeoutSeconds: 30,
	}

	executor := NewSubagentExecutor(
		config,
		[]string{"bash", "ls"},
		"gpt-4",
		nil,
		nil,
		"thread-1",
		"trace-1",
		logger,
	)

	// 测试同步执行
	t.Run("Execute sync", func(t *testing.T) {
		result := executor.Execute("Test task", nil)

		assert.NotNil(t, result)
		assert.Equal(t, SubagentStatusCompleted, result.Status)
		assert.NotEmpty(t, result.Result)
	})

	// 测试异步执行
	t.Run("Execute async", func(t *testing.T) {
		StartPools()
		defer StopPools()

		taskID := executor.ExecuteAsync("Async test task", "")
		assert.NotEmpty(t, taskID)

		// 检查任务是否被创建
		result := GetBackgroundTaskResult(taskID)
		assert.NotNil(t, result)
	})
}

// ============================================
// 便捷函数测试
// ============================================

func TestConvenienceFunctions(t *testing.T) {
	// 测试 NewSubagentExecutorContext
	t.Run("NewSubagentExecutorContext", func(t *testing.T) {
		config := DefaultSubagentConfig()
		ctx := NewSubagentExecutorContext(
			config,
			"gpt-4",
			nil,
			nil,
			"thread-1",
			"trace-1",
			[]string{"bash"},
		)

		assert.Equal(t, config, ctx.Config)
		assert.Equal(t, "gpt-4", ctx.ParentModel)
		assert.Equal(t, "thread-1", ctx.ThreadID)
		assert.Equal(t, "trace-1", ctx.TraceID)
		assert.Contains(t, ctx.AvailableTools, "bash")
	})
}
