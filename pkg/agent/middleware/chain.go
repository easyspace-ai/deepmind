package middleware

import (
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/config"
	"github.com/weibaohui/nanobot-go/pkg/sandbox"
	"go.uber.org/zap"
)

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	// 基础配置
	BaseDir string
	LazyInit bool

	// SandboxProvider 沙箱提供者（可选）。为 nil 时使用 NewDefaultSandboxMiddleware()（无真实 acquire，仅占位逻辑）。
	SandboxProvider sandbox.SandboxProvider

	// DeferredToolNames 延迟加载的工具名（DeferredToolFilterMiddleware；后续可扩展为从 schema 中隐藏）。
	DeferredToolNames []string

	// 标题生成配置
	TitleEnabled bool
	TitleModelName string

	// 摘要配置
	SummarizationEnabled bool

	// 任务列表配置
	TodoListEnabled bool

	// 记忆配置
	MemoryEnabled bool

	// 子代理配置
	SubagentEnabled bool
	MaxConcurrentSubagents int

	// 循环检测配置
	LoopWarnThreshold int
	LoopHardLimit int

	// Logger
	Logger *zap.Logger
}

// DefaultMiddlewareConfig 默认中间件配置
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		LazyInit:           true,
		TitleEnabled:       true,
		LoopWarnThreshold:  DefaultWarnThreshold,
		LoopHardLimit:      DefaultHardLimit,
		MaxConcurrentSubagents: 3,
		Logger:             zap.NewNop(),
	}
}

// middlewareConfigLogger 返回非 nil 的 zap.Logger。
func middlewareConfigLogger(c *MiddlewareConfig) *zap.Logger {
	if c == nil || c.Logger == nil {
		return zap.NewNop()
	}
	return c.Logger
}

// resolveLoopThresholds 解析循环检测阈值，0 或负数时使用包内默认。
func resolveLoopThresholds(c *MiddlewareConfig) (warn, hard int) {
	warn, hard = DefaultWarnThreshold, DefaultHardLimit
	if c == nil {
		return warn, hard
	}
	if c.LoopWarnThreshold > 0 {
		warn = c.LoopWarnThreshold
	}
	if c.LoopHardLimit > 0 {
		hard = c.LoopHardLimit
	}
	return warn, hard
}

// BuildLeadAgentMiddlewares 构建 Lead Agent 中间件链
// 一比一复刻 DeerFlow 的中间件顺序（14个）
// 严格按照 DeerFlow Python 的 _build_middlewares() 顺序
func BuildLeadAgentMiddlewares(config *MiddlewareConfig) *MiddlewareChain {
	return BuildLeadAgentMiddlewaresWithConfig(config, nil)
}

// BuildLeadAgentMiddlewaresWithConfig 构建 Lead Agent 中间件链（带 appConfig 支持）
// 一比一复刻 DeerFlow 的中间件顺序（14个）
// 严格按照 DeerFlow Python 的 _build_middlewares() 顺序
func BuildLeadAgentMiddlewaresWithConfig(config *MiddlewareConfig, appConfigGetter func() *config.AppConfig) *MiddlewareChain {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	log := middlewareConfigLogger(config)
	warnTh, hardLim := resolveLoopThresholds(config)

	chain := NewMiddlewareChain()

	// ============================================
	// 1. Runtime 基础链 (build_lead_runtime_middlewares)
	// ============================================

	// 1. ThreadDataMiddleware - 创建线程目录 ✅
	chain.Add(NewThreadDataMiddleware(config.BaseDir, config.LazyInit))

	// 2. UploadsMiddleware - 上传文件追踪 ✅
	chain.Add(NewUploadsMiddleware(config.BaseDir))

	// 3. SandboxMiddleware - 沙箱生命周期 ✅
	if config.SandboxProvider != nil {
		chain.Add(NewSandboxMiddleware(config.SandboxProvider, log))
	} else {
		chain.Add(NewDefaultSandboxMiddleware())
	}

	// 4. DanglingToolCallMiddleware - 处理挂起工具调用 ✅
	chain.Add(NewDanglingToolCallMiddleware(log))

	// 5. ToolErrorHandlingMiddleware - 工具错误处理 ✅ (在 runtime 链中)
	chain.Add(NewToolErrorHandlingMiddleware(log))

	// ============================================
	// 2. Lead 专用链
	// ============================================

	// 6. SummarizationMiddleware - 上下文摘要（可选）✅
	if config.SummarizationEnabled {
		chain.Add(NewSummarizationMiddleware(true, log))
	}

	// 7. TodoListMiddleware - 任务追踪（可选）✅
	if config.TodoListEnabled {
		chain.Add(NewTodoListMiddleware(true, log))
	}

	// 8. TitleMiddleware - 自动生成标题 ✅
	if config.TitleEnabled {
		chain.Add(NewTitleMiddleware(true, config.TitleModelName, log))
	}

	// 9. MemoryMiddleware - 记忆提取（可选）✅
	if config.MemoryEnabled {
		chain.Add(NewMemoryMiddleware(true, log))
	}

	// 10. ViewImageMiddleware - 图像处理 ✅ (条件：模型支持 vision)
	chain.Add(NewViewImageMiddleware())

	// 11. DeferredToolFilterMiddleware - 延迟工具过滤 ✅ (条件：tool_search.enabled)
	if appConfigGetter != nil {
		appCfg := appConfigGetter()
		if appCfg != nil && appCfg.ToolSearch.Enabled {
			chain.Add(NewDeferredToolFilterMiddlewareWithNames(config.DeferredToolNames, log))
		}
	} else {
		// 如果没有配置 getter，仍然添加（向后兼容）
		chain.Add(NewDeferredToolFilterMiddlewareWithNames(config.DeferredToolNames, log))
	}

	// 12. SubagentLimitMiddleware - 限制并发（可选）✅
	if config.SubagentEnabled {
		maxConc := config.MaxConcurrentSubagents
		if maxConc <= 0 {
			maxConc = 3
		}
		chain.Add(NewSubagentLimitMiddleware(maxConc, log))
	}

	// 13. LoopDetectionMiddleware - 循环检测 ✅
	chain.Add(NewLoopDetectionMiddlewareWithConfig(
		warnTh, hardLim, DefaultWindowSize, DefaultMaxTrackedThreads, log,
	))

	// 14. ClarificationMiddleware - 询问澄清（必须最后！）✅
	chain.Add(NewClarificationMiddleware())

	return chain
}

// BuildSubagentMiddlewares 构建子代理中间件链
func BuildSubagentMiddlewares(config *MiddlewareConfig) *MiddlewareChain {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	log := middlewareConfigLogger(config)
	warnTh, hardLim := resolveLoopThresholds(config)

	chain := NewMiddlewareChain()

	// 子代理使用简化版中间件链
	// 1. ThreadDataMiddleware
	chain.Add(NewThreadDataMiddleware(config.BaseDir, config.LazyInit))

	// 2. LoopDetectionMiddleware
	chain.Add(NewLoopDetectionMiddlewareWithConfig(
		warnTh, hardLim, DefaultWindowSize, DefaultMaxTrackedThreads, log,
	))

	return chain
}

// ============================================
// 中间件执行上下文
// ============================================

// MiddlewareContext 中间件执行上下文
type MiddlewareContext struct {
	ThreadID string
	State    *state.ThreadState
	Config   *MiddlewareConfig
}

// NewMiddlewareContext 创建中间件执行上下文
func NewMiddlewareContext(threadID string, state *state.ThreadState, config *MiddlewareConfig) *MiddlewareContext {
	return &MiddlewareContext{
		ThreadID: threadID,
		State:    state,
		Config:   config,
	}
}
