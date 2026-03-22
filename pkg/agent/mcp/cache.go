package mcp

import (
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"go.uber.org/zap"
)

// ============================================
// MCP Cache（一比一复刻 DeerFlow）
// ============================================

var (
	mcpToolsCache     []tool.BaseTool
	cacheInitialized  bool
	initializationLock sync.Mutex
	configMtime       *time.Time
	logger            *zap.Logger
)

// SetLogger 设置日志器
func SetLogger(l *zap.Logger) {
	logger = l
}

// getConfigMtime 获取配置文件修改时间
func getConfigMtime(configPath string) *time.Time {
	if configPath == "" {
		return nil
	}
	info, err := os.Stat(configPath)
	if err != nil {
		return nil
	}
	mtime := info.ModTime()
	return &mtime
}

// isCacheStale 检查缓存是否过期
// 一比一复刻 DeerFlow 的 _is_cache_stale
func isCacheStale(configPath string) bool {
	if !cacheInitialized {
		return false
	}

	currentMtime := getConfigMtime(configPath)

	// 如果之前或现在都无法获取 mtime，假设不过期
	if configMtime == nil || currentMtime == nil {
		return false
	}

	// 如果配置文件自缓存后被修改，则缓存过期
	if currentMtime.After(*configMtime) {
		if logger != nil {
			logger.Info("MCP config file has been modified, cache is stale",
				zap.Time("old_mtime", *configMtime),
				zap.Time("new_mtime", *currentMtime))
		}
		return true
	}

	return false
}

// InitializeMCPTools 初始化并缓存 MCP 工具
// 一比一复刻 DeerFlow 的 initialize_mcp_tools
func InitializeMCPTools(configPath string, loader func() ([]tool.BaseTool, error)) ([]tool.BaseTool, error) {
	initializationLock.Lock()
	defer initializationLock.Unlock()

	if cacheInitialized {
		if logger != nil {
			logger.Info("MCP tools already initialized")
		}
		return mcpToolsCache, nil
	}

	if logger != nil {
		logger.Info("Initializing MCP tools...")
	}

	var err error
	mcpToolsCache, err = loader()
	if err != nil {
		return nil, err
	}

	cacheInitialized = true
	configMtime = getConfigMtime(configPath)

	if logger != nil {
		logger.Info("MCP tools initialized",
			zap.Int("tool_count", len(mcpToolsCache)),
			zap.Time("config_mtime", *configMtime))
	}

	return mcpToolsCache, nil
}

// GetCachedMCPTools 获取缓存的 MCP 工具（带懒加载）
// 一比一复刻 DeerFlow 的 get_cached_mcp_tools
func GetCachedMCPTools(configPath string, loader func() ([]tool.BaseTool, error)) []tool.BaseTool {
	// 检查缓存是否因配置文件变更而过期
	if isCacheStale(configPath) {
		if logger != nil {
			logger.Info("MCP cache is stale, resetting for re-initialization...")
		}
		ResetMCPToolsCache()
	}

	if !cacheInitialized {
		if logger != nil {
			logger.Info("MCP tools not initialized, performing lazy initialization...")
		}
		_, err := InitializeMCPTools(configPath, loader)
		if err != nil {
			if logger != nil {
				logger.Error("Failed to lazy-initialize MCP tools", zap.Error(err))
			}
			return nil
		}
	}

	return mcpToolsCache
}

// ResetMCPToolsCache 重置 MCP 工具缓存
// 一比一复刻 DeerFlow 的 reset_mcp_tools_cache
func ResetMCPToolsCache() {
	initializationLock.Lock()
	defer initializationLock.Unlock()

	mcpToolsCache = nil
	cacheInitialized = false
	configMtime = nil

	if logger != nil {
		logger.Info("MCP tools cache reset")
	}
}

// GetCacheState 获取缓存状态（用于调试）
func GetCacheState() (initialized bool, toolCount int, mtime *time.Time) {
	return cacheInitialized, len(mcpToolsCache), configMtime
}
