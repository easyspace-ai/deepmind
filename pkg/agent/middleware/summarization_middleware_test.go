package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

func TestSummarizationMiddleware_Name(t *testing.T) {
	m := NewDefaultSummarizationMiddleware()

	if m.Name() != "summarization" {
		t.Errorf("Name() = %v, want 'summarization'", m.Name())
	}
}

func TestSummarizationMiddleware_NewDefault(t *testing.T) {
	m := NewDefaultSummarizationMiddleware()

	if m.enabled {
		t.Error("enabled = true, want false by default")
	}
}

func TestSummarizationMiddleware_NewWithConfig(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewSummarizationMiddleware(true, customLogger)

	if !m.enabled {
		t.Error("enabled = false, want true")
	}
	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}
}

func TestSummarizationMiddleware_shouldSummarize_Disabled(t *testing.T) {
	m := NewSummarizationMiddleware(false, nil)
	ts := state.NewThreadState()

	result := m.shouldSummarize(ts)

	if result {
		t.Error("shouldSummarize() = true, want false (disabled)")
	}
}

func TestSummarizationMiddleware_shouldSummarize_UnderThreshold(t *testing.T) {
	m := NewSummarizationMiddleware(true, nil)
	ts := state.NewThreadState()

	// 添加 10 条消息（低于 20 的阈值）
	for i := 0; i < 10; i++ {
		ts.Messages = append(ts.Messages, &schema.Message{
			Role:    schema.User,
			Content: "message",
		})
	}

	result := m.shouldSummarize(ts)

	if result {
		t.Error("shouldSummarize() = true, want false (under threshold)")
	}
}

func TestSummarizationMiddleware_shouldSummarize_OverThreshold(t *testing.T) {
	m := NewSummarizationMiddleware(true, nil)
	ts := state.NewThreadState()

	// 添加 21 条消息（超过阈值）
	for i := 0; i < 21; i++ {
		ts.Messages = append(ts.Messages, &schema.Message{
			Role:    schema.User,
			Content: "message",
		})
	}

	result := m.shouldSummarize(ts)

	if !result {
		t.Error("shouldSummarize() = false, want true (over threshold)")
	}
}

func TestSummarizationMiddleware_buildSummary(t *testing.T) {
	m := NewSummarizationMiddleware(true, nil)
	ts := state.NewThreadState()

	summary := m.buildSummary(ts)

	if summary == "" {
		t.Error("buildSummary() = empty string, want non-empty")
	}
	if len(summary) == 0 {
		t.Error("buildSummary() should have content")
	}
}

func TestSummarizationMiddleware_BeforeModel_Disabled(t *testing.T) {
	m := NewSummarizationMiddleware(false, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestSummarizationMiddleware_BeforeModel_UnderThreshold(t *testing.T) {
	m := NewSummarizationMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 添加 10 条消息
	for i := 0; i < 10; i++ {
		ts.Messages = append(ts.Messages, &schema.Message{
			Role:    schema.User,
			Content: "message",
		})
	}

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeModel() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestSummarizationMiddleware_BeforeModel_OverThreshold(t *testing.T) {
	m := NewSummarizationMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 添加 25 条消息（超过阈值）
	for i := 0; i < 25; i++ {
		ts.Messages = append(ts.Messages, &schema.Message{
			Role:    schema.User,
			Content: "message",
		})
	}

	originalCount := len(ts.Messages)

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Fatal("BeforeModel() stateUpdate = nil, want non-nil")
	}

	// 验证 stateUpdate
	summarized, ok := stateUpdate["summarized"]
	if !ok {
		t.Error("stateUpdate should contain 'summarized'")
	}
	if summarized != true {
		t.Error("summarized should be true")
	}

	// 验证消息数量减少了
	if len(ts.Messages) >= originalCount {
		t.Errorf("message count = %v, want < %v", len(ts.Messages), originalCount)
	}

	// 验证第一条是摘要消息
	if len(ts.Messages) > 0 && ts.Messages[0].Role != schema.System {
		t.Error("first message should be system summary")
	}
}

func TestSummarizationMiddleware_BeforeModel_JustUnderTen(t *testing.T) {
	m := NewSummarizationMiddleware(true, nil)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 刚好 10 条消息，不应该摘要（需要 >10）
	for i := 0; i < 10; i++ {
		ts.Messages = append(ts.Messages, &schema.Message{
			Role:    schema.User,
			Content: "message",
		})
	}

	stateUpdate, err := m.BeforeModel(ctx, ts)

	if err != nil {
		t.Errorf("BeforeModel() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeModel() stateUpdate = %v, want nil (exactly 10 messages)", stateUpdate)
	}
}

func TestSummarizationMiddleware_Logger(t *testing.T) {
	customLogger := zap.NewExample()
	m := NewSummarizationMiddleware(true, customLogger)

	if m.logger != customLogger {
		t.Error("custom logger should be used")
	}

	m2 := NewSummarizationMiddleware(true, nil)
	if m2.logger == nil {
		t.Error("logger should not be nil (should use NopLogger)")
	}
}

func TestSummarizationMiddleware_IntegrationWithChain(t *testing.T) {
	config := DefaultMiddlewareConfig()
	config.SummarizationEnabled = true // 需要启用
	chain := BuildLeadAgentMiddlewares(config)

	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "summarization" {
			found = true
			break
		}
	}
	if !found {
		t.Error("SummarizationMiddleware not found in chain (need SummarizationEnabled=true)")
	}
}

// BenchmarkSummarizationMiddleware_shouldSummarize 性能基准测试
func BenchmarkSummarizationMiddleware_shouldSummarize(b *testing.B) {
	m := NewSummarizationMiddleware(true, nil)
	ts := state.NewThreadState()
	for i := 0; i < 25; i++ {
		ts.Messages = append(ts.Messages, &schema.Message{
			Role:    schema.User,
			Content: "message",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.shouldSummarize(ts)
	}
}

// BenchmarkSummarizationMiddleware_BeforeModel 性能基准测试
func BenchmarkSummarizationMiddleware_BeforeModel(b *testing.B) {
	m := NewSummarizationMiddleware(true, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts := state.NewThreadState()
		for j := 0; j < 25; j++ {
			ts.Messages = append(ts.Messages, &schema.Message{
				Role:    schema.User,
				Content: "message",
			})
		}
		_, _ = m.BeforeModel(ctx, ts)
	}
}
