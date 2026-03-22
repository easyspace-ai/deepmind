package mcp

import (
	"fmt"

	"github.com/weibaohui/nanobot-go/pkg/config"
)

// ============================================
// MCP Client Config（一比一复刻 DeerFlow）
// ============================================

// ServerParams MCP 服务器参数
type ServerParams map[string]interface{}

// BuildServerParams 构建服务器参数
// 一比一复刻 DeerFlow 的 build_server_params
func BuildServerParams(serverName string, config *config.MCPServerConfig) (ServerParams, error) {
	transportType := config.Type
	if transportType == "" {
		transportType = "stdio"
	}

	params := ServerParams{
		"transport": transportType,
	}

	switch transportType {
	case "stdio":
		if config.Command == "" {
			return nil, fmt.Errorf("MCP server '%s' with stdio transport requires 'command' field", serverName)
		}
		params["command"] = config.Command
		if len(config.Args) > 0 {
			params["args"] = config.Args
		}
		if len(config.Env) > 0 {
			params["env"] = config.Env
		}

	case "sse", "http":
		if config.URL == "" {
			return nil, fmt.Errorf("MCP server '%s' with %s transport requires 'url' field", serverName, transportType)
		}
		params["url"] = config.URL
		if len(config.Headers) > 0 {
			params["headers"] = config.Headers
		}

	default:
		return nil, fmt.Errorf("MCP server '%s' has unsupported transport type: %s", serverName, transportType)
	}

	return params, nil
}

// BuildServersConfig 构建多服务器配置
// 一比一复刻 DeerFlow 的 build_servers_config
func BuildServersConfig(extensionsConfig *config.ExtensionsConfig) (map[string]ServerParams, error) {
	enabledServers := getEnabledMCPServers(extensionsConfig)

	if len(enabledServers) == 0 {
		return map[string]ServerParams{}, nil
	}

	serversConfig := make(map[string]ServerParams)
	for serverName, serverConfig := range enabledServers {
		params, err := BuildServerParams(serverName, &serverConfig)
		if err != nil {
			return nil, err
		}
		serversConfig[serverName] = params
	}

	return serversConfig, nil
}

// getEnabledMCPServers 获取启用的 MCP 服务器
func getEnabledMCPServers(extensionsConfig *config.ExtensionsConfig) map[string]config.MCPServerConfig {
	enabled := make(map[string]config.MCPServerConfig)
	for name, cfg := range extensionsConfig.MCPServers {
		if cfg.Enabled {
			enabled[name] = cfg
		}
	}
	return enabled
}

// InjectOAuthHeaders 注入 OAuth headers 到服务器配置
func InjectOAuthHeaders(serversConfig map[string]ServerParams, oauthHeaders map[string]string) {
	for serverName, authHeader := range oauthHeaders {
		params, ok := serversConfig[serverName]
		if !ok {
			continue
		}

		transport, _ := params["transport"].(string)
		if transport != "sse" && transport != "http" {
			continue
		}

		// 获取现有 headers
		var headers map[string]string
		if existingHeaders, ok := params["headers"].(map[string]string); ok {
			headers = make(map[string]string)
			for k, v := range existingHeaders {
				headers[k] = v
			}
		} else {
			headers = make(map[string]string)
		}

		// 添加 Authorization header
		headers["Authorization"] = authHeader
		params["headers"] = headers
	}
}
