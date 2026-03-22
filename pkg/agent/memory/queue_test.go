package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================
// 测试：Queue（去抖队列）
// ============================================

func TestMemoryUpdateQueue_New(t *testing.T) {
	q := NewMemoryUpdateQueue(nil, nil, nil)
	require.NotNil(t, q)
	require.Equal(t, 0, q.PendingCount())
	require.False(t, q.IsProcessing())
}

func TestMemoryUpdateQueue_Clear(t *testing.T) {
	q := NewMemoryUpdateQueue(nil, nil, nil)
	q.Clear()
	require.Equal(t, 0, q.PendingCount())
	require.False(t, q.IsProcessing())
}

func TestMemoryUpdateQueue_SetLLMCaller(t *testing.T) {
	q := NewMemoryUpdateQueue(nil, nil, nil)
	q.SetLLMCaller(func(prompt string) (string, error) {
		return "{}", nil
	})
	require.NotNil(t, q.llmCaller)
}

func TestGetMemoryQueue_Singleton(t *testing.T) {
	q1 := GetMemoryQueue()
	q2 := GetMemoryQueue()
	require.Equal(t, q1, q2)
}

func TestResetMemoryQueue(t *testing.T) {
	ResetMemoryQueue()
	// 只是确保不崩溃
	require.True(t, true)
}

// ============================================
// 测试：配置
// ============================================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)
	require.Equal(t, 30, cfg.DebounceSeconds)
	require.Equal(t, 100, cfg.MaxFacts)
	require.Equal(t, 0.7, cfg.FactConfidenceThreshold)
	require.Equal(t, 2000, cfg.MaxInjectionTokens)
}

func TestMemoryConfig_Values(t *testing.T) {
	cfg := &MemoryConfig{
		Enabled:                true,
		InjectionEnabled:       true,
		DebounceSeconds:        10,
		MaxFacts:               50,
		FactConfidenceThreshold: 0.5,
		MaxInjectionTokens:     1000,
	}

	require.True(t, cfg.Enabled)
	require.True(t, cfg.InjectionEnabled)
	require.Equal(t, 10, cfg.DebounceSeconds)
	require.Equal(t, 50, cfg.MaxFacts)
	require.Equal(t, 0.5, cfg.FactConfidenceThreshold)
	require.Equal(t, 1000, cfg.MaxInjectionTokens)
}

// ============================================
// 测试：ConversationContext
// ============================================

func TestConversationContext_Values(t *testing.T) {
	ctx := &ConversationContext{
		ThreadID:  "test-thread",
		Messages:  []any{"msg1", "msg2"},
		AgentName: "test-agent",
		Timestamp: nowForTest(),
	}

	require.Equal(t, "test-thread", ctx.ThreadID)
	require.Len(t, ctx.Messages, 2)
	require.Equal(t, "test-agent", ctx.AgentName)
	require.False(t, ctx.Timestamp.IsZero())
}

func nowForTest() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}
