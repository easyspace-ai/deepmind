package subagent

import (
	"sync"
)

// ============================================
// 事件类型
// ============================================

// TaskEventType 任务事件类型
type TaskEventType string

const (
	TaskEventTypeStarted    TaskEventType = "task_started"
	TaskEventTypeRunning    TaskEventType = "task_running"
	TaskEventTypeCompleted  TaskEventType = "task_completed"
	TaskEventTypeFailed     TaskEventType = "task_failed"
	TaskEventTypeTimedOut   TaskEventType = "task_timed_out"
)

// TaskEvent 任务事件
type TaskEvent struct {
	Type          TaskEventType     `json:"type"`
	TaskID        string            `json:"task_id"`
	Description   string            `json:"description,omitempty"`
	Message       map[string]any    `json:"message,omitempty"`
	MessageIndex  int               `json:"message_index,omitempty"`
	TotalMessages int               `json:"total_messages,omitempty"`
	Result        string            `json:"result,omitempty"`
	Error         string            `json:"error,omitempty"`
}

// ============================================
// 事件总线
// ============================================

// EventHandler 事件处理函数
type EventHandler func(event TaskEvent)

// TaskEventBus 任务事件总线
type TaskEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewTaskEventBus 创建任务事件总线
func NewTaskEventBus() *TaskEventBus {
	return &TaskEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe 订阅事件
func (b *TaskEventBus) Subscribe(taskID string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[taskID] = append(b.handlers[taskID], handler)
}

// Unsubscribe 取消订阅
func (b *TaskEventBus) Unsubscribe(taskID string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[taskID]
	for i, h := range handlers {
		if &h == &handler {
			b.handlers[taskID] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish 发布事件
func (b *TaskEventBus) Publish(event TaskEvent) {
	b.mu.RLock()
	handlers := append([]EventHandler{}, b.handlers[event.TaskID]...)
	b.mu.RUnlock()

	for _, handler := range handlers {
		func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// 忽略处理函数的 panic
				}
			}()
			h(event)
		}(handler)
	}
}

// Clear 清除任务的所有订阅
func (b *TaskEventBus) Clear(taskID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.handlers, taskID)
}

// 全局事件总线
var (
	globalEventBus = NewTaskEventBus()
)

// GetGlobalEventBus 获取全局事件总线
func GetGlobalEventBus() *TaskEventBus {
	return globalEventBus
}

// ============================================
// 事件创建函数
// ============================================

// NewTaskStartedEvent 创建任务启动事件
func NewTaskStartedEvent(taskID, description string) TaskEvent {
	return TaskEvent{
		Type:        TaskEventTypeStarted,
		TaskID:      taskID,
		Description: description,
	}
}

// NewTaskRunningEvent 创建任务运行中事件
func NewTaskRunningEvent(taskID string, message map[string]any, messageIndex, totalMessages int) TaskEvent {
	return TaskEvent{
		Type:          TaskEventTypeRunning,
		TaskID:        taskID,
		Message:       message,
		MessageIndex:  messageIndex,
		TotalMessages: totalMessages,
	}
}

// NewTaskCompletedEvent 创建任务完成事件
func NewTaskCompletedEvent(taskID, result string) TaskEvent {
	return TaskEvent{
		Type:   TaskEventTypeCompleted,
		TaskID: taskID,
		Result: result,
	}
}

// NewTaskFailedEvent 创建任务失败事件
func NewTaskFailedEvent(taskID, err string) TaskEvent {
	return TaskEvent{
		Type:   TaskEventTypeFailed,
		TaskID: taskID,
		Error:  err,
	}
}

// NewTaskTimedOutEvent 创建任务超时事件
func NewTaskTimedOutEvent(taskID, err string) TaskEvent {
	return TaskEvent{
		Type:   TaskEventTypeTimedOut,
		TaskID: taskID,
		Error:  err,
	}
}
