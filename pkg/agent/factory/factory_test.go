package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weibaohui/nanobot-go/pkg/sandbox"
	"go.uber.org/zap"
)

type stubSandboxProvider struct {
	releasedThread string
}

func (s *stubSandboxProvider) Acquire(string) (sandbox.Sandbox, error) { return nil, nil }
func (s *stubSandboxProvider) Get(string) (sandbox.Sandbox, bool)      { return nil, false }
func (s *stubSandboxProvider) Release(threadID string) error {
	s.releasedThread = threadID
	return nil
}

// ============================================
// LeadAgentConfig 测试
// ============================================

func TestDefaultLeadAgentConfig(t *testing.T) {
	config := DefaultLeadAgentConfig()

	assert.True(t, config.LazyInit)
	assert.True(t, config.ThinkingEnabled)
	assert.True(t, config.TitleEnabled)
	assert.False(t, config.SummarizationEnabled)
	assert.False(t, config.TodoListEnabled)
	assert.False(t, config.MemoryEnabled)
	assert.True(t, config.SubagentEnabled)
	assert.False(t, config.IsPlanMode)
	assert.False(t, config.IsBootstrap)
	assert.Equal(t, 3, config.MaxConcurrentSubagents)
	assert.NotNil(t, config.Logger)
}

// ============================================
// LeadAgentFactory 测试
// ============================================

func TestNewLeadAgentFactory(t *testing.T) {
	config := DefaultLeadAgentConfig()
	factory := NewLeadAgentFactory(config)

	assert.NotNil(t, factory)
	assert.Equal(t, config, factory.config)
}

func TestLeadAgentFactory_BuildMiddlewareConfig(t *testing.T) {
	config := DefaultLeadAgentConfig()
	config.TitleEnabled = true
	config.SummarizationEnabled = true
	config.TodoListEnabled = true
	config.MemoryEnabled = true
	config.SubagentEnabled = true
	config.MaxConcurrentSubagents = 5

	factory := NewLeadAgentFactory(config)
	mwConfig := factory.BuildMiddlewareConfig()

	assert.Equal(t, config.TitleEnabled, mwConfig.TitleEnabled)
	assert.Equal(t, config.SummarizationEnabled, mwConfig.SummarizationEnabled)
	assert.Equal(t, config.TodoListEnabled, mwConfig.TodoListEnabled)
	assert.Equal(t, config.MemoryEnabled, mwConfig.MemoryEnabled)
	assert.Equal(t, config.SubagentEnabled, mwConfig.SubagentEnabled)
	assert.Equal(t, config.MaxConcurrentSubagents, mwConfig.MaxConcurrentSubagents)
}

func TestLeadAgentFactory_BuildMiddlewareConfig_SandboxProvider(t *testing.T) {
	prov := &stubSandboxProvider{}
	cfg := DefaultLeadAgentConfig()
	cfg.SandboxProvider = prov
	f := NewLeadAgentFactory(cfg)
	mw := f.BuildMiddlewareConfig()
	assert.Equal(t, prov, mw.SandboxProvider)
}

func TestLeadAgentFactory_CleanupThread_Release(t *testing.T) {
	prov := &stubSandboxProvider{}
	cfg := DefaultLeadAgentConfig()
	cfg.SandboxProvider = prov
	f := NewLeadAgentFactory(cfg)
	f.CleanupThread("thread-xyz")
	assert.Equal(t, "thread-xyz", prov.releasedThread)
}

func TestLeadAgentFactory_BuildSystemPrompt(t *testing.T) {
	config := DefaultLeadAgentConfig()
	factory := NewLeadAgentFactory(config)

	prompt := factory.BuildSystemPrompt()

	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "<role>")
	assert.Contains(t, prompt, "<thinking_style>")
}

func TestLeadAgentFactory_BuildMiddlewares(t *testing.T) {
	config := DefaultLeadAgentConfig()
	factory := NewLeadAgentFactory(config)

	chain := factory.BuildMiddlewares()

	assert.NotNil(t, chain)
	// 中间件链应该包含多个中间件
	assert.Greater(t, len(chain.Middlewares()), 0)
}

// ============================================
// MakeLeadAgent 测试
// ============================================

func TestMakeLeadAgent(t *testing.T) {
	config := DefaultLeadAgentConfig()
	config.Logger = zap.NewNop()

	factory, err := MakeLeadAgent(config)

	assert.NoError(t, err)
	assert.NotNil(t, factory)

	// 清理
	factory.Cleanup()
}

func TestMakeLeadAgent_NilConfig(t *testing.T) {
	factory, err := MakeLeadAgent(nil)

	assert.NoError(t, err)
	assert.NotNil(t, factory)

	factory.Cleanup()
}

// ============================================
// RuntimeConfig 测试
// ============================================

func TestRuntimeConfig(t *testing.T) {
	config := DefaultLeadAgentConfig()
	factory := NewLeadAgentFactory(config)

	// 测试默认值
	assert.True(t, factory.config.ThinkingEnabled)
	assert.False(t, factory.config.IsPlanMode)
	assert.True(t, factory.config.SubagentEnabled)
	assert.Equal(t, 3, factory.config.MaxConcurrentSubagents)

	// 应用运行时配置
	thinkingEnabled := false
	isPlanMode := true
	subagentEnabled := false
	maxConcurrent := 10
	modelName := "gpt-4"

	rc := &RuntimeConfig{
		ThinkingEnabled:     &thinkingEnabled,
		ModelName:         &modelName,
		IsPlanMode:        &isPlanMode,
		SubagentEnabled:   &subagentEnabled,
		MaxConcurrentSubagents: &maxConcurrent,
	}

	factory.ApplyRuntimeConfig(rc)

	// 验证配置已更新
	assert.False(t, factory.config.ThinkingEnabled)
	assert.Equal(t, "gpt-4", factory.config.ModelName)
	assert.True(t, factory.config.IsPlanMode)
	assert.False(t, factory.config.SubagentEnabled)
	assert.Equal(t, 10, factory.config.MaxConcurrentSubagents)
}

func TestRuntimeConfig_Partial(t *testing.T) {
	config := DefaultLeadAgentConfig()
	factory := NewLeadAgentFactory(config)

	originalThinking := factory.config.ThinkingEnabled

	// 只更新部分配置
	isPlanMode := true
	rc := &RuntimeConfig{
		IsPlanMode: &isPlanMode,
	}

	factory.ApplyRuntimeConfig(rc)

	// 验证只更新了指定的配置
	assert.Equal(t, originalThinking, factory.config.ThinkingEnabled) // 保持不变
	assert.True(t, factory.config.IsPlanMode) // 已更新
}

func TestRuntimeConfig_Nil(t *testing.T) {
	config := DefaultLeadAgentConfig()
	factory := NewLeadAgentFactory(config)

	originalConfig := *factory.config

	// 应用 nil 配置
	factory.ApplyRuntimeConfig(nil)

	// 验证没有变化
	assert.Equal(t, originalConfig, *factory.config)
}
