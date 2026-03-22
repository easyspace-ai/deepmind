package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

// 默认配置常量
const (
	DefaultWarnThreshold     = 3  // 警告阈值
	DefaultHardLimit         = 5  // 强制停止阈值
	DefaultWindowSize        = 20 // 滑动窗口大小
	DefaultMaxTrackedThreads = 100 // 最大跟踪线程数
)

// 警告消息
const (
	WarningMsg = "[LOOP DETECTED] You are repeating the same tool calls. " +
		"Stop calling tools and produce your final answer now. " +
		"If you cannot complete the task, summarize what you accomplished so far."

	HardStopMsg = "[FORCED STOP] Repeated tool calls exceeded the safety limit. " +
		"Producing final answer with results collected so far."
)

// ToolCall 工具调用
type ToolCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
	ID   string                 `json:"id,omitempty"`
}

// LoopDetectionMiddleware 循环检测中间件
// 一比一复刻 DeerFlow 的 LoopDetectionMiddleware
type LoopDetectionMiddleware struct {
	*BaseMiddleware
	warnThreshold     int
	hardLimit         int
	windowSize        int
	maxTrackedThreads int
	logger            *zap.Logger
	lock              sync.Mutex
	history           map[string][]string // thread_id -> [call_hash]
	warned            map[string]map[string]bool // thread_id -> call_hash -> warned
}

// NewLoopDetectionMiddleware 创建循环检测中间件
func NewLoopDetectionMiddleware(logger *zap.Logger) *LoopDetectionMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LoopDetectionMiddleware{
		BaseMiddleware:    NewBaseMiddleware("loop_detection"),
		warnThreshold:     DefaultWarnThreshold,
		hardLimit:         DefaultHardLimit,
		windowSize:        DefaultWindowSize,
		maxTrackedThreads: DefaultMaxTrackedThreads,
		logger:            logger,
		history:           make(map[string][]string),
		warned:            make(map[string]map[string]bool),
	}
}

// NewLoopDetectionMiddlewareWithConfig 带配置创建循环检测中间件
func NewLoopDetectionMiddlewareWithConfig(
	warnThreshold, hardLimit, windowSize, maxTrackedThreads int,
	logger *zap.Logger,
) *LoopDetectionMiddleware {
	m := NewLoopDetectionMiddleware(logger)
	m.warnThreshold = warnThreshold
	m.hardLimit = hardLimit
	m.windowSize = windowSize
	m.maxTrackedThreads = maxTrackedThreads
	return m
}

// hashToolCalls 计算工具调用的哈希值
// 一比一复刻 DeerFlow 的 _hash_tool_calls() - order-independent
func hashToolCalls(toolCalls []ToolCall) string {
	// 先标准化每个工具调用
	normalized := make([]struct {
		Name string                 `json:"name"`
		Args map[string]interface{} `json:"args"`
	}, len(toolCalls))

	for i, tc := range toolCalls {
		normalized[i].Name = tc.Name
		normalized[i].Args = tc.Args
	}

	// 排序：按名称和 args 的确定性序列化
	// 使用冒泡排序，比较 name 和 args 的 JSON 字符串
	for i := 0; i < len(normalized); i++ {
		for j := i + 1; j < len(normalized); j++ {
			if compareToolCalls(normalized[i], normalized[j]) > 0 {
				normalized[i], normalized[j] = normalized[j], normalized[i]
			}
		}
	}

	// 序列化并计算 MD5
	blob, _ := json.Marshal(normalized)
	hash := md5.Sum(blob)
	return hex.EncodeToString(hash[:])[:12]
}

// compareToolCalls 比较两个工具调用
func compareToolCalls(a, b struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}) int {
	if a.Name != b.Name {
		if a.Name < b.Name {
			return -1
		}
		return 1
	}
	// 比较 args 的 JSON
	aArgs, _ := json.Marshal(a.Args)
	bArgs, _ := json.Marshal(b.Args)
	aStr := string(aArgs)
	bStr := string(bArgs)
	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

// TrackAndCheck 跟踪工具调用并检查循环
// 返回：(warningMessage, shouldHardStop)
func (m *LoopDetectionMiddleware) TrackAndCheck(threadID string, toolCalls []ToolCall) (string, bool) {
	if threadID == "" {
		threadID = "default"
	}

	callHash := hashToolCalls(toolCalls)

	m.lock.Lock()
	defer m.lock.Unlock()

	// 获取或创建历史记录
	history, exists := m.history[threadID]
	if !exists {
		// LRU 淘汰：如果超过最大跟踪线程数，删除最早的
		if len(m.history) >= m.maxTrackedThreads {
			// 删除最早的（遍历 map，删除第一个）
			for id := range m.history {
				delete(m.history, id)
				delete(m.warned, id)
				m.logger.Debug("Evicted loop tracking for thread", zap.String("thread_id", id))
				break
			}
		}
		history = []string{}
	}

	// 添加新的哈希值
	history = append(history, callHash)
	// 保持窗口大小
	if len(history) > m.windowSize {
		history = history[len(history)-m.windowSize:]
	}
	m.history[threadID] = history

	// 计数
	count := 0
	for _, h := range history {
		if h == callHash {
			count++
		}
	}

	// 收集工具名称用于日志
	toolNames := make([]string, 0, len(toolCalls))
	for _, tc := range toolCalls {
		toolNames = append(toolNames, tc.Name)
	}

	// 检查硬限制
	if count >= m.hardLimit {
		m.logger.Error("Loop hard limit reached — forcing stop",
			zap.String("thread_id", threadID),
			zap.String("call_hash", callHash),
			zap.Int("count", count),
			zap.Strings("tools", toolNames),
		)
		return HardStopMsg, true
	}

	// 检查警告阈值
	if count >= m.warnThreshold {
		// 检查是否已经警告过
		warnedSet, exists := m.warned[threadID]
		if !exists {
			warnedSet = make(map[string]bool)
			m.warned[threadID] = warnedSet
		}
		if !warnedSet[callHash] {
			warnedSet[callHash] = true
			m.logger.Warn("Repetitive tool calls detected — injecting warning",
				zap.String("thread_id", threadID),
				zap.String("call_hash", callHash),
				zap.Int("count", count),
				zap.Strings("tools", toolNames),
			)
			return WarningMsg, false
		}
		// 已经警告过
		return "", false
	}

	return "", false
}

// Reset 重置跟踪状态
func (m *LoopDetectionMiddleware) Reset(threadID string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if threadID != "" {
		delete(m.history, threadID)
		delete(m.warned, threadID)
	} else {
		// 重置所有
		m.history = make(map[string][]string)
		m.warned = make(map[string]map[string]bool)
	}
}
