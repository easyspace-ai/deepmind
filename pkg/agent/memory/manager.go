package memory

import (
	"strings"
	"sync"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// Manager DeerFlow 风格记忆去抖队列的轻量实现：聚合文本片段，供后续 LLM 事实提取接入。
type Manager struct {
	mu      sync.Mutex
	pending map[string]string
	logger  *zap.Logger
}

var defaultManager = &Manager{
	pending: make(map[string]string),
	logger:  zap.NewNop(),
}

// DefaultManager 返回进程级默认 Manager（测试可替换为独立实例）。
func DefaultManager() *Manager {
	return defaultManager
}

// SetLogger 设置默认 Manager 日志。
func SetLogger(l *zap.Logger) {
	if l == nil {
		l = zap.NewNop()
	}
	defaultManager.mu.Lock()
	defaultManager.logger = l
	defaultManager.mu.Unlock()
}

// EnqueueFromThreadState 从线程状态抽取最近用户/助手文本并入队（key 使用 workspace path 或占位 default）。
func (m *Manager) EnqueueFromThreadState(ts *state.ThreadState) {
	if m == nil || ts == nil {
		return
	}
	key := "default"
	if ts.ThreadData != nil && ts.ThreadData.WorkspacePath != "" {
		key = ts.ThreadData.WorkspacePath
	}
	var b strings.Builder
	for _, msg := range ts.Messages {
		if msg == nil {
			continue
		}
		if msg.Role != schema.User && msg.Role != schema.Assistant {
			continue
		}
		if msg.Content != "" {
			b.WriteString(string(msg.Role))
			b.WriteString(": ")
			b.WriteString(msg.Content)
			b.WriteByte('\n')
		}
	}
	if b.Len() == 0 {
		return
	}
	snippet := b.String()
	if len(snippet) > 8000 {
		snippet = snippet[len(snippet)-8000:]
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pending == nil {
		m.pending = make(map[string]string)
	}
	m.pending[key] = snippet
	log := m.logger
	if log == nil {
		log = zap.NewNop()
	}
	log.Debug("memory queue updated",
		zap.String("thread_key", key),
		zap.Int("bytes", len(snippet)))
}

// PendingSnapshot 返回当前待处理片段副本（测试/调试）。
func (m *Manager) PendingSnapshot() map[string]string {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]string, len(m.pending))
	for k, v := range m.pending {
		out[k] = v
	}
	return out
}

// ============================================
// 便捷函数：整合完整 Memory 系统
// ============================================

// UpdateMemoryFromConversation 便捷函数：从对话更新记忆
func UpdateMemoryFromConversation(messages []any, threadID string, agentName string) bool {
	updater := NewMemoryUpdater("", nil)
	queue := GetMemoryQueue()
	return updater.UpdateMemory(messages, threadID, agentName, queue.llmCaller)
}

// FormatMemoryForPrompt 便捷函数：格式化记忆用于提示词注入
func FormatMemoryForPrompt(memoryData *MemoryData) string {
	cfg := DefaultConfig()
	return FormatMemoryForInjection(memoryData, cfg.MaxInjectionTokens)
}
