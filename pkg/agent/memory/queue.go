package memory

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// ============================================
// Memory Update Queue（一比一复刻 DeerFlow）
// ============================================

// ConversationContext 对话上下文
type ConversationContext struct {
	ThreadID   string
	Messages   []any
	AgentName  string
	Timestamp  time.Time
}

// MemoryUpdateQueue 记忆更新去抖队列
type MemoryUpdateQueue struct {
	mu         sync.Mutex
	queue      []*ConversationContext
	timer      *time.Timer
	processing bool
	llmCaller  func(prompt string) (string, error)
	logger     *zap.Logger
	updater    *MemoryUpdater
}

var (
	globalQueue     *MemoryUpdateQueue
	globalQueueOnce sync.Once
	globalQueueMu   sync.Mutex
)

// GetMemoryQueue 获取全局记忆更新队列（单例）
func GetMemoryQueue() *MemoryUpdateQueue {
	globalQueueOnce.Do(func() {
		globalQueue = NewMemoryUpdateQueue(nil, nil, nil)
	})
	return globalQueue
}

// ResetMemoryQueue 重置全局记忆队列（用于测试）
func ResetMemoryQueue() {
	globalQueueMu.Lock()
	defer globalQueueMu.Unlock()
	if globalQueue != nil {
		globalQueue.Clear()
	}
	globalQueue = nil
	globalQueueOnce = sync.Once{}
}

// NewMemoryUpdateQueue 创建记忆更新队列
func NewMemoryUpdateQueue(
	llmCaller func(prompt string) (string, error),
	updater *MemoryUpdater,
	logger *zap.Logger,
) *MemoryUpdateQueue {
	if logger == nil {
		logger = zap.NewNop()
	}
	if updater == nil {
		updater = NewMemoryUpdater("", logger)
	}
	return &MemoryUpdateQueue{
		queue:     make([]*ConversationContext, 0),
		llmCaller: llmCaller,
		logger:    logger,
		updater:   updater,
	}
}

// SetLLMCaller 设置 LLM 调用函数
func (q *MemoryUpdateQueue) SetLLMCaller(caller func(prompt string) (string, error)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.llmCaller = caller
}

// Add 添加对话到更新队列
// 如果同一 threadID 已有待处理的更新，替换为新的
func (q *MemoryUpdateQueue) Add(threadID string, messages []any, agentName string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	cfg := DefaultConfig()
	if !cfg.Enabled {
		return
	}

	ctx := &ConversationContext{
		ThreadID:  threadID,
		Messages:  messages,
		AgentName: agentName,
		Timestamp: time.Now().UTC(),
	}

	// 替换相同 threadID 的待处理条目
	filtered := make([]*ConversationContext, 0, len(q.queue))
	for _, c := range q.queue {
		if c.ThreadID != threadID {
			filtered = append(filtered, c)
		}
	}
	filtered = append(filtered, ctx)
	q.queue = filtered

	// 重置去抖计时器
	q.resetTimerLocked(cfg.DebounceSeconds)

	q.logger.Debug("Memory update queued",
		zap.String("thread_id", threadID),
		zap.Int("queue_size", len(q.queue)))
}

// resetTimerLocked 重置去抖计时器（必须在持有锁的情况下调用）
func (q *MemoryUpdateQueue) resetTimerLocked(debounceSeconds int) {
	// 取消现有计时器
	if q.timer != nil {
		q.timer.Stop()
	}

	// 启动新计时器
	q.timer = time.AfterFunc(time.Duration(debounceSeconds)*time.Second, q.processQueue)
}

// processQueue 处理队列中的所有对话上下文
func (q *MemoryUpdateQueue) processQueue() {
	q.mu.Lock()

	if q.processing {
		// 正在处理中，重新调度
		cfg := DefaultConfig()
		q.resetTimerLocked(cfg.DebounceSeconds)
		q.mu.Unlock()
		return
	}

	if len(q.queue) == 0 {
		q.mu.Unlock()
		return
	}

	q.processing = true
	contextsToProcess := make([]*ConversationContext, len(q.queue))
	copy(contextsToProcess, q.queue)
	q.queue = q.queue[:0]
	q.timer = nil
	q.mu.Unlock()

	q.logger.Debug("Processing queued memory updates", zap.Int("count", len(contextsToProcess)))

	defer func() {
		q.mu.Lock()
		q.processing = false
		q.mu.Unlock()
	}()

	for _, ctx := range contextsToProcess {
		q.logger.Debug("Updating memory for thread", zap.String("thread_id", ctx.ThreadID))

		success := q.updater.UpdateMemory(ctx.Messages, ctx.ThreadID, ctx.AgentName, q.llmCaller)
		if success {
			q.logger.Debug("Memory updated successfully", zap.String("thread_id", ctx.ThreadID))
		} else {
			q.logger.Debug("Memory update skipped/failed", zap.String("thread_id", ctx.ThreadID))
		}

		// 多个更新之间的小延迟，避免速率限制
		if len(contextsToProcess) > 1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Flush 强制立即处理队列
func (q *MemoryUpdateQueue) Flush() {
	q.mu.Lock()
	if q.timer != nil {
		q.timer.Stop()
		q.timer = nil
	}
	q.mu.Unlock()

	q.processQueue()
}

// Clear 清空队列而不处理
func (q *MemoryUpdateQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.timer != nil {
		q.timer.Stop()
		q.timer = nil
	}
	q.queue = q.queue[:0]
	q.processing = false
}

// PendingCount 获取待处理更新数量
func (q *MemoryUpdateQueue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

// IsProcessing 检查是否正在处理
func (q *MemoryUpdateQueue) IsProcessing() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.processing
}

// ============================================
// 便捷配置访问
// ============================================

// MemoryConfig 记忆配置（简化访问）
type MemoryConfig struct {
	Enabled                bool
	InjectionEnabled       bool
	DebounceSeconds        int
	MaxFacts               int
	FactConfidenceThreshold float64
	MaxInjectionTokens     int
}

// DefaultConfig 获取默认记忆配置
func DefaultConfig() *MemoryConfig {
	cfg := configDefault()
	return &MemoryConfig{
		Enabled:                cfg.Enabled,
		InjectionEnabled:       cfg.InjectionEnabled,
		DebounceSeconds:        cfg.DebounceSeconds,
		MaxFacts:               cfg.MaxFacts,
		FactConfidenceThreshold: cfg.FactConfidenceThreshold,
		MaxInjectionTokens:     cfg.MaxInjectionTokens,
	}
}

// 内部配置获取器（可被测试替换）
var configDefault = func() *MemoryConfig {
	// 从 pkg/config 获取
	importedCfg := defaultConfigImport()
	return &MemoryConfig{
		Enabled:                importedCfg.Enabled,
		InjectionEnabled:       importedCfg.InjectionEnabled,
		DebounceSeconds:        importedCfg.DebounceSeconds,
		MaxFacts:               importedCfg.MaxFacts,
		FactConfidenceThreshold: importedCfg.FactConfidenceThreshold,
		MaxInjectionTokens:     importedCfg.MaxInjectionTokens,
	}
}

// 从 pkg/config 导入的配置类型
type importedMemoryConfig struct {
	Enabled                bool
	InjectionEnabled       bool
	DebounceSeconds        int
	MaxFacts               int
	FactConfidenceThreshold float64
	MaxInjectionTokens     int
}

var defaultConfigImport = func() *importedMemoryConfig {
	// 直接返回默认值，避免循环导入
	return &importedMemoryConfig{
		Enabled:                false,
		InjectionEnabled:       true,
		DebounceSeconds:        30,
		MaxFacts:               100,
		FactConfidenceThreshold: 0.7,
		MaxInjectionTokens:     2000,
	}
}

// SetConfigDefaultForTest 为测试设置配置默认值（仅测试使用）
func SetConfigDefaultForTest(cfg *MemoryConfig) {
	orig := configDefault
	configDefault = func() *MemoryConfig {
		return cfg
	}
	// 返回恢复函数
	_ = orig
}
