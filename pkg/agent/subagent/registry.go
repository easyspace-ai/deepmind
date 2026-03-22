package subagent

import (
	"sync"

	"go.uber.org/zap"
)

// ============================================
// SubagentRegistry - 子代理注册表
// 一比一复刻 DeerFlow 的 registry
// ============================================

// SubagentRegistry 子代理注册表
type SubagentRegistry struct {
	mu         sync.RWMutex
	configs    map[string]*SubagentConfig
	timeouts   map[string]int // 配置的超时覆盖
	logger     *zap.Logger
}

// NewSubagentRegistry 创建子代理注册表
func NewSubagentRegistry(logger *zap.Logger) *SubagentRegistry {
	if logger == nil {
		logger = zap.NewNop()
	}

	r := &SubagentRegistry{
		configs:  make(map[string]*SubagentConfig),
		timeouts: make(map[string]int),
		logger:   logger,
	}

	// 注册内置子代理
	for name, config := range BuiltinSubagents {
		r.Register(name, config)
	}

	return r
}

// Register 注册子代理配置
func (r *SubagentRegistry) Register(name string, config *SubagentConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[name] = config
	r.logger.Debug("Subagent registered", zap.String("name", name))
}

// Get 获取子代理配置（应用超时覆盖）
func (r *SubagentRegistry) Get(name string) *SubagentConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config := r.configs[name]
	if config == nil {
		return nil
	}

	// 创建副本以避免修改原始配置
	result := *config

	// 应用超时覆盖
	if timeout, ok := r.timeouts[name]; ok && timeout != result.TimeoutSeconds {
		r.logger.Debug("Subagent timeout overridden",
			zap.String("name", name),
			zap.Int("original", result.TimeoutSeconds),
			zap.Int("override", timeout))
		result.TimeoutSeconds = timeout
	}

	return &result
}

// List 列出所有子代理配置
func (r *SubagentRegistry) List() []*SubagentConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*SubagentConfig, 0, len(r.configs))
	for name := range r.configs {
		results = append(results, r.Get(name))
	}
	return results
}

// Names 获取所有子代理名称
func (r *SubagentRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.configs))
	for name := range r.configs {
		names = append(names, name)
	}
	return names
}

// SetTimeoutOverride 设置超时覆盖
func (r *SubagentRegistry) SetTimeoutOverride(name string, seconds int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timeouts[name] = seconds
	r.logger.Debug("Subagent timeout override set",
		zap.String("name", name),
		zap.Int("seconds", seconds))
}

// ClearTimeoutOverride 清除超时覆盖
func (r *SubagentRegistry) ClearTimeoutOverride(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.timeouts, name)
}

// ============================================
// 全局注册表
// ============================================

var (
	globalRegistry *SubagentRegistry
	registryOnce   sync.Once
)

// GetGlobalRegistry 获取全局注册表
func GetGlobalRegistry() *SubagentRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewSubagentRegistry(nil)
	})
	return globalRegistry
}

// SetGlobalRegistryLogger 设置全局注册表日志器
func SetGlobalRegistryLogger(logger *zap.Logger) {
	GetGlobalRegistry().logger = logger
}

// ============================================
// 便捷函数
// ============================================

// GetSubagentConfig 获取子代理配置（便捷函数）
func GetSubagentConfig(name string) *SubagentConfig {
	return GetGlobalRegistry().Get(name)
}

// ListSubagents 列出所有子代理配置（便捷函数）
func ListSubagents() []*SubagentConfig {
	return GetGlobalRegistry().List()
}

// GetSubagentNames 获取所有子代理名称（便捷函数）
func GetSubagentNames() []string {
	return GetGlobalRegistry().Names()
}
