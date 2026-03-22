package mcp

import (
	"github.com/cloudwego/eino/components/tool"
	"github.com/weibaohui/nanobot-go/pkg/config"
	"go.uber.org/zap"
)

// ============================================
// MCP Tools Integration（一比一复刻 DeerFlow）
// ============================================

// MCPToolsLoader MCP 工具加载器函数类型
type MCPToolsLoader func() ([]tool.BaseTool, error)

// GetMCPTools 获取所有 MCP 工具
// 一比一复刻 DeerFlow 的 get_mcp_tools
func GetMCPTools(
	extensionsConfig *config.ExtensionsConfig,
	logger *zap.Logger,
	loader MCPToolsLoader,
) ([]tool.BaseTool, error) {

	if extensionsConfig == nil {
		extensionsConfig = &config.ExtensionsConfig{
			MCPServers: make(map[string]config.MCPServerConfig),
		}
	}

	// 构建服务器配置
	serversConfig, err := BuildServersConfig(extensionsConfig)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to build MCP servers config", zap.Error(err))
		}
		return nil, err
	}

	if len(serversConfig) == 0 {
		if logger != nil {
			logger.Info("No enabled MCP servers configured")
		}
		return []tool.BaseTool{}, nil
	}

	// 获取初始 OAuth headers
	oauthHeaders, err := GetInitialOAuthHeaders(extensionsConfig, logger)
	if err != nil {
		if logger != nil {
			logger.Warn("Failed to get initial OAuth headers", zap.Error(err))
		}
		// 继续执行，不因为 OAuth 失败而中断
	}

	// 注入 OAuth headers
	InjectOAuthHeaders(serversConfig, oauthHeaders)

	// 使用加载器加载工具
	if loader != nil {
		tools, err := loader()
		if err != nil {
			if logger != nil {
				logger.Error("Failed to load MCP tools", zap.Error(err))
			}
			return nil, err
		}

		if logger != nil {
			logger.Info("Successfully loaded MCP tools", zap.Int("tool_count", len(tools)))
		}

		return tools, nil
	}

	// 没有加载器时返回空列表
	if logger != nil {
		logger.Info("No MCP tools loader provided")
	}

	return []tool.BaseTool{}, nil
}

// GetCachedMCPToolsWithConfig 获取缓存的 MCP 工具（带配置）
func GetCachedMCPToolsWithConfig(
	configPath string,
	extensionsConfig *config.ExtensionsConfig,
	logger *zap.Logger,
	loader MCPToolsLoader,
) []tool.BaseTool {

	// 设置 logger
	SetLogger(logger)

	// 包装 loader
	wrappedLoader := func() ([]tool.BaseTool, error) {
		return GetMCPTools(extensionsConfig, logger, loader)
	}

	return GetCachedMCPTools(configPath, wrappedLoader)
}
