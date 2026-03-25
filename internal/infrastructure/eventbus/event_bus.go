package eventbus

import (
	"context"
	"sync"
)

// EventHandler 事件处理器类型
type EventHandler func(context.Context, interface{}) error

// SimpleEventBus 简单的事件总线实现
type SimpleEventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
}

// NewSimpleEventBus 创建新的 SimpleEventBus
func NewSimpleEventBus() *SimpleEventBus {
	return &SimpleEventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

// Subscribe 订阅事件
func (eb *SimpleEventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

// Publish 发布事件
func (eb *SimpleEventBus) Publish(ctx context.Context, eventType string, event interface{}) error {
	eb.mu.RLock()
	handlers := eb.subscribers[eventType]
	// 复制一份以避免持行过程中的并发问题
	handlersCopy := make([]EventHandler, len(handlers))
	copy(handlersCopy, handlers)
	eb.mu.RUnlock()

	// 异步执行处理器
	for _, handler := range handlersCopy {
		go func(h EventHandler) {
			_ = h(ctx, event) // 忽略错误以继续处理
		}(handler)
	}

	return nil
}

// SubscribeAll 订阅所有事件（通用处理器）
func (eb *SimpleEventBus) SubscribeAll(handler EventHandler) {
	eb.Subscribe("*", handler)
}
