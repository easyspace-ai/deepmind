package middleware

import (
	"testing"
)

func TestLoopDetectionMiddleware_Name(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)

	if m.Name() != "loop_detection" {
		t.Errorf("Name() = %v, want 'loop_detection'", m.Name())
	}
}

func TestHashToolCalls_OrderIndependent(t *testing.T) {
	// 顺序不同但内容相同的工具调用应该有相同的哈希
	calls1 := []ToolCall{
		{Name: "tool1", Args: map[string]interface{}{"a": 1}},
		{Name: "tool2", Args: map[string]interface{}{"b": 2}},
	}
	calls2 := []ToolCall{
		{Name: "tool2", Args: map[string]interface{}{"b": 2}},
		{Name: "tool1", Args: map[string]interface{}{"a": 1}},
	}

	hash1 := hashToolCalls(calls1)
	hash2 := hashToolCalls(calls2)

	if hash1 != hash2 {
		t.Errorf("hashToolCalls() order-dependent: %v != %v", hash1, hash2)
	}
}

func TestHashToolCalls_DifferentContent(t *testing.T) {
	// 内容不同的工具调用应该有不同的哈希
	calls1 := []ToolCall{
		{Name: "tool1", Args: map[string]interface{}{"a": 1}},
	}
	calls2 := []ToolCall{
		{Name: "tool1", Args: map[string]interface{}{"a": 2}},
	}

	hash1 := hashToolCalls(calls1)
	hash2 := hashToolCalls(calls2)

	if hash1 == hash2 {
		t.Error("hashToolCalls() same hash for different content")
	}
}

func TestHashToolCalls_Empty(t *testing.T) {
	// 空工具调用
	hash := hashToolCalls([]ToolCall{})
	if hash == "" {
		t.Error("hashToolCalls(empty) = empty string, want non-empty")
	}
}

func TestHashToolCalls_SingleCall(t *testing.T) {
	// 单个工具调用
	calls := []ToolCall{
		{Name: "tool", Args: map[string]interface{}{"key": "value"}},
	}
	hash1 := hashToolCalls(calls)
	hash2 := hashToolCalls(calls)

	if hash1 != hash2 {
		t.Error("hashToolCalls() not deterministic")
	}
}

func TestHashToolCalls_WithID(t *testing.T) {
	// ID 不影响哈希（DeerFlow 行为）
	calls1 := []ToolCall{
		{Name: "tool", Args: map[string]interface{}{"a": 1}, ID: "id1"},
	}
	calls2 := []ToolCall{
		{Name: "tool", Args: map[string]interface{}{"a": 1}, ID: "id2"},
	}

	hash1 := hashToolCalls(calls1)
	hash2 := hashToolCalls(calls2)

	if hash1 != hash2 {
		t.Error("hashToolCalls() should ignore tool call ID")
	}
}

func TestLoopDetectionMiddleware_TrackAndCheck_NoWarning(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	threadID := "thread-1"

	// 第一次调用
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}
	msg, shouldStop := m.TrackAndCheck(threadID, calls)

	if msg != "" {
		t.Errorf("TrackAndCheck() msg = %v, want empty", msg)
	}
	if shouldStop {
		t.Error("TrackAndCheck() shouldStop = true, want false")
	}
}

func TestLoopDetectionMiddleware_TrackAndCheck_WarningThreshold(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	threadID := "thread-2"
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 调用 DefaultWarnThreshold 次（达到警告阈值）
	for i := 0; i < DefaultWarnThreshold-1; i++ {
		_, _ = m.TrackAndCheck(threadID, calls)
	}

	// 再调用一次应该触发警告
	msg, shouldStop := m.TrackAndCheck(threadID, calls)

	if msg == "" {
		t.Error("TrackAndCheck() msg = empty, want warning")
	}
	if msg != WarningMsg {
		t.Errorf("TrackAndCheck() msg = %v, want WarningMsg", msg)
	}
	if shouldStop {
		t.Error("TrackAndCheck() shouldStop = true, want false")
	}
}

func TestLoopDetectionMiddleware_TrackAndCheck_WarningOnlyOnce(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	threadID := "thread-3"
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 调用直到警告出现
	for i := 0; i < DefaultWarnThreshold; i++ {
		_, _ = m.TrackAndCheck(threadID, calls)
	}

	// 再调用一次，不应该再警告（因为已经警告过了）
	msg, shouldStop := m.TrackAndCheck(threadID, calls)

	if msg != "" {
		t.Error("TrackAndCheck() should not warn twice for same pattern")
	}
	if shouldStop {
		t.Error("TrackAndCheck() shouldStop = true, want false")
	}
}

func TestLoopDetectionMiddleware_TrackAndCheck_HardLimit(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	threadID := "thread-4"
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 调用直到硬限制
	for i := 0; i < DefaultHardLimit; i++ {
		_, _ = m.TrackAndCheck(threadID, calls)
	}

	// 第 5 次应该强制停止
	msg, shouldStop := m.TrackAndCheck(threadID, calls)

	if msg == "" {
		t.Error("TrackAndCheck() msg = empty, want hard stop message")
	}
	if msg != HardStopMsg {
		t.Errorf("TrackAndCheck() msg = %v, want HardStopMsg", msg)
	}
	if !shouldStop {
		t.Error("TrackAndCheck() shouldStop = false, want true")
	}
}

func TestLoopDetectionMiddleware_TrackAndCheck_DifferentThreads(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 线程 1：多次调用
	for i := 0; i < DefaultWarnThreshold; i++ {
		_, _ = m.TrackAndCheck("thread-5a", calls)
	}

	// 线程 2：第一次调用，不应该受影响
	msg, shouldStop := m.TrackAndCheck("thread-5b", calls)

	if msg != "" {
		t.Error("TrackAndCheck() different threads should not interfere")
	}
	if shouldStop {
		t.Error("TrackAndCheck() shouldStop = true, want false")
	}
}

func TestLoopDetectionMiddleware_TrackAndCheck_EmptyThreadID(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 空 threadID 应该使用 "default"
	msg, shouldStop := m.TrackAndCheck("", calls)

	if msg != "" {
		t.Error("TrackAndCheck(empty threadID) should work")
	}
	if shouldStop {
		t.Error("TrackAndCheck() shouldStop = true, want false")
	}
}

func TestLoopDetectionMiddleware_Reset_SingleThread(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	threadID := "thread-6"
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 多次调用
	for i := 0; i < DefaultWarnThreshold; i++ {
		_, _ = m.TrackAndCheck(threadID, calls)
	}

	// 重置
	m.Reset(threadID)

	// 再次调用，应该从头开始计数
	msg, shouldStop := m.TrackAndCheck(threadID, calls)

	if msg != "" {
		t.Error("TrackAndCheck() after reset should not warn")
	}
	if shouldStop {
		t.Error("TrackAndCheck() shouldStop = true, want false")
	}
}

func TestLoopDetectionMiddleware_Reset_All(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 多个线程
	for i := 0; i < DefaultWarnThreshold; i++ {
		_, _ = m.TrackAndCheck("thread-7a", calls)
		_, _ = m.TrackAndCheck("thread-7b", calls)
	}

	// 重置所有
	m.Reset("")

	// 验证已重置
	msg, _ := m.TrackAndCheck("thread-7a", calls)
	if msg != "" {
		t.Error("TrackAndCheck() after reset all should not warn")
	}
}

func TestLoopDetectionMiddleware_CustomConfig(t *testing.T) {
	m := NewLoopDetectionMiddlewareWithConfig(2, 3, 10, 50, nil)
	threadID := "thread-8"
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 验证自定义警告阈值
	_, _ = m.TrackAndCheck(threadID, calls)  // 1
	msg, _ := m.TrackAndCheck(threadID, calls) // 2 - 警告

	if msg != WarningMsg {
		t.Error("TrackAndCheck() with custom warn threshold should warn at 2")
	}
}

func TestLoopDetectionMiddleware_LRUEviction(t *testing.T) {
	m := NewLoopDetectionMiddlewareWithConfig(3, 5, 20, 2, nil)
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	// 跟踪 2 个线程（达到最大限制）
	_, _ = m.TrackAndCheck("thread-9a", calls)
	_, _ = m.TrackAndCheck("thread-9b", calls)

	// 第 3 个线程应该触发 LRU 淘汰
	_, _ = m.TrackAndCheck("thread-9c", calls)

	// 验证历史记录数量不超过限制
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.history) > 2 {
		t.Errorf("history size = %v, want <= 2", len(m.history))
	}
}

func TestLoopDetectionMiddleware_DifferentPatterns(t *testing.T) {
	m := NewLoopDetectionMiddleware(nil)
	threadID := "thread-10"

	// 交替调用不同模式，不应该触发警告
	for i := 0; i < DefaultWarnThreshold; i++ {
		var calls []ToolCall
		if i%2 == 0 {
			calls = []ToolCall{{Name: "toolA", Args: map[string]interface{}{"a": 1}}}
		} else {
			calls = []ToolCall{{Name: "toolB", Args: map[string]interface{}{"b": 2}}}
		}
		_, shouldStop := m.TrackAndCheck(threadID, calls)
		if shouldStop {
			t.Error("TrackAndCheck() shouldStop = true, want false for different patterns")
		}
	}
}

func TestCompareToolCalls(t *testing.T) {
	a := struct {
		Name string                 `json:"name"`
		Args map[string]interface{} `json:"args"`
	}{"toolA", map[string]interface{}{"a": 1}}
	b := struct {
		Name string                 `json:"name"`
		Args map[string]interface{} `json:"args"`
	}{"toolB", map[string]interface{}{"b": 2}}

	if compareToolCalls(a, b) >= 0 {
		t.Error("compareToolCalls(a, b) should be < 0")
	}
	if compareToolCalls(b, a) <= 0 {
		t.Error("compareToolCalls(b, a) should be > 0")
	}
	if compareToolCalls(a, a) != 0 {
		t.Error("compareToolCalls(a, a) should be 0")
	}
}

// BenchmarkLoopDetectionMiddleware_TrackAndCheck 性能基准测试
func BenchmarkLoopDetectionMiddleware_TrackAndCheck(b *testing.B) {
	m := NewLoopDetectionMiddleware(nil)
	calls := []ToolCall{{Name: "tool", Args: map[string]interface{}{"a": 1}}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.TrackAndCheck("bench-thread", calls)
	}
}

// BenchmarkHashToolCalls 哈希计算性能基准测试
func BenchmarkHashToolCalls(b *testing.B) {
	calls := []ToolCall{
		{Name: "tool1", Args: map[string]interface{}{"a": 1, "b": "test"}},
		{Name: "tool2", Args: map[string]interface{}{"c": 2.5, "d": true}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashToolCalls(calls)
	}
}
