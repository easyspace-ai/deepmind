package factory

import (
	"github.com/weibaohui/nanobot-go/pkg/agent/middleware"
	"github.com/weibaohui/nanobot-go/pkg/agent/prompts"
	"github.com/weibaohui/nanobot-go/pkg/agent/subagent"
	"github.com/weibaohui/nanobot-go/pkg/config"
	"github.com/weibaohui/nanobot-go/pkg/sandbox"
	"go.uber.org/zap"
)

// ============================================
// LeadAgentConfig - Lead Agent 配置
// ============================================

// LeadAgentConfig Lead Agent 配置
type LeadAgentConfig struct {
	// 基础配置
	BaseDir           string
	LazyInit          bool
	ThinkingEnabled   bool
	ModelName         string

	// 特性开关
	TitleEnabled      bool
	SummarizationEnabled bool
	TodoListEnabled  bool
	MemoryEnabled     bool
	SubagentEnabled   bool
	IsPlanMode        bool
	IsBootstrap       bool

	// 子代理配置
	MaxConcurrentSubagents int

	// Agent 名称
	AgentName         string

	// 日志
	Logger            *zap.Logger

	// AppConfigGetter 应用配置获取器（可选，用于中间件链构建）
	AppConfigGetter func() *config.AppConfig

	// SandboxProvider 可选；非空时 BuildMiddlewareConfig 会传给中间件，并可在会话结束时 CleanupThread 释放。
	SandboxProvider sandbox.SandboxProvider
}

// DefaultLeadAgentConfig 默认 Lead Agent 配置
func DefaultLeadAgentConfig() *LeadAgentConfig {
	return &LeadAgentConfig{
		LazyInit:              true,
		ThinkingEnabled:       true,
		TitleEnabled:          true,
		SummarizationEnabled: false,
		TodoListEnabled:       false,
		MemoryEnabled:         false,
		SubagentEnabled:       true,
		IsPlanMode:            false,
		IsBootstrap:           false,
		MaxConcurrentSubagents: 3,
		Logger:                 zap.NewNop(),
	}
}

// ============================================
// LeadAgentFactory - Lead Agent 工厂
// ============================================

// LeadAgentFactory Lead Agent 工厂
// 一比一复刻 DeerFlow 的 make_lead_agent
type LeadAgentFactory struct {
	config          *LeadAgentConfig
	logger          *zap.Logger
	appConfig       *config.AppConfig
	appConfigGetter func() *config.AppConfig
}

// NewLeadAgentFactory 创建 Lead Agent 工厂
func NewLeadAgentFactory(config *LeadAgentConfig) *LeadAgentFactory {
	if config == nil {
		config = DefaultLeadAgentConfig()
	}
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	return &LeadAgentFactory{
		config:          config,
		logger:          config.Logger,
		appConfigGetter: config.AppConfigGetter,
	}
}

// BuildMiddlewareConfig 构建中间件配置
func (f *LeadAgentFactory) BuildMiddlewareConfig() *middleware.MiddlewareConfig {
	return &middleware.MiddlewareConfig{
		BaseDir:              f.config.BaseDir,
		LazyInit:             f.config.LazyInit,
		SandboxProvider:      f.config.SandboxProvider,
		TitleEnabled:         f.config.TitleEnabled,
		TitleModelName:       "", // TODO: 从配置中获取
		SummarizationEnabled: f.config.SummarizationEnabled,
		TodoListEnabled:      f.config.TodoListEnabled,
		MemoryEnabled:        f.config.MemoryEnabled,
		SubagentEnabled:      f.config.SubagentEnabled,
		MaxConcurrentSubagents: f.config.MaxConcurrentSubagents,
		LoopWarnThreshold:   middleware.DefaultWarnThreshold,
		LoopHardLimit:       middleware.DefaultHardLimit,
		Logger:               f.logger,
	}
}

// BuildSystemPrompt 构建系统提示词
func (f *LeadAgentFactory) BuildSystemPrompt() string {
	pc := &prompts.LeadAgentConfig{
		AgentName:              f.config.AgentName,
		SubagentEnabled:        f.config.SubagentEnabled,
		MaxConcurrentSubagents: f.config.MaxConcurrentSubagents,
	}
	if pc.AgentName == "" {
		pc.AgentName = "DeerFlow 2.0"
	}
	if pc.MaxConcurrentSubagents <= 0 {
		pc.MaxConcurrentSubagents = 3
	}
	return prompts.BuildLeadAgentPromptString(pc)
}

// BuildMiddlewares 构建中间件链
func (f *LeadAgentFactory) BuildMiddlewares() *middleware.MiddlewareChain {
	mwConfig := f.BuildMiddlewareConfig()
	if f.appConfigGetter != nil {
		return middleware.BuildLeadAgentMiddlewaresWithConfig(mwConfig, f.appConfigGetter)
	}
	return middleware.BuildLeadAgentMiddlewares(mwConfig)
}

// ============================================
// 便捷函数
// ============================================

// MakeLeadAgent 便捷函数：创建 Lead Agent
// 一比一复刻 DeerFlow 的 make_lead_agent
func MakeLeadAgent(config *LeadAgentConfig) (*LeadAgentFactory, error) {
	if config == nil {
		config = DefaultLeadAgentConfig()
	}

	// 初始化子代理系统
	subagent.SetGlobalRegistryLogger(config.Logger)
	subagent.SetPoolsLogger(config.Logger)
	subagent.StartPools()

	factory := NewLeadAgentFactory(config)

	config.Logger.Info("LeadAgentFactory created",
		zap.String("agent_name", config.AgentName),
		zap.Bool("thinking_enabled", config.ThinkingEnabled),
		zap.String("model_name", config.ModelName),
		zap.Bool("is_plan_mode", config.IsPlanMode),
		zap.Bool("subagent_enabled", config.SubagentEnabled),
		zap.Int("max_concurrent_subagents", config.MaxConcurrentSubagents))

	return factory, nil
}

// ============================================
// 运行时配置覆盖
// ============================================

// RuntimeConfig 运行时配置
type RuntimeConfig struct {
	ThinkingEnabled     *bool   `json:"thinking_enabled,omitempty"`
	ModelName         *string `json:"model_name,omitempty"`
	IsPlanMode        *bool   `json:"is_plan_mode,omitempty"`
	SubagentEnabled   *bool   `json:"subagent_enabled,omitempty"`
	MaxConcurrentSubagents *int `json:"max_concurrent_subagents,omitempty"`
}

// ApplyRuntimeConfig 应用运行时配置覆盖
func (f *LeadAgentFactory) ApplyRuntimeConfig(rc *RuntimeConfig) {
	if rc == nil {
		return
	}

	if rc.ThinkingEnabled != nil {
		f.config.ThinkingEnabled = *rc.ThinkingEnabled
	}
	if rc.ModelName != nil {
		f.config.ModelName = *rc.ModelName
	}
	if rc.IsPlanMode != nil {
		f.config.IsPlanMode = *rc.IsPlanMode
	}
	if rc.SubagentEnabled != nil {
		f.config.SubagentEnabled = *rc.SubagentEnabled
	}
	if rc.MaxConcurrentSubagents != nil {
		f.config.MaxConcurrentSubagents = *rc.MaxConcurrentSubagents
	}

	f.logger.Debug("Runtime config applied",
		zap.Any("config", rc))
}

// ============================================
// 资源清理
// ============================================

// Cleanup 清理资源
func (f *LeadAgentFactory) Cleanup() {
	subagent.StopPools()
	f.logger.Debug("LeadAgentFactory cleanup completed")
}

// CleanupThread 释放指定线程的沙箱（会话/WebSocket 断开时由调用方触发；无 Provider 时为 no-op）。
func (f *LeadAgentFactory) CleanupThread(threadID string) {
	if f == nil || threadID == "" || f.config == nil || f.config.SandboxProvider == nil {
		return
	}
	if err := f.config.SandboxProvider.Release(threadID); err != nil {
		f.logger.Debug("sandbox release",
			zap.String("thread_id", threadID),
			zap.Error(err))
	}
}
