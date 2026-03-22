package community

import (
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/weibaohui/nanobot-go/pkg/agent/tools/community/firecrawl"
	"github.com/weibaohui/nanobot-go/pkg/agent/tools/community/imagesearch"
	"github.com/weibaohui/nanobot-go/pkg/agent/tools/community/jinaai"
	"github.com/weibaohui/nanobot-go/pkg/agent/tools/community/tavily"
)

// ============================================
// Community Tools Factory（一比一复刻 DeerFlow）
// ============================================

// CommunityToolProvider 社区工具类型
type CommunityToolProvider string

const (
	// ProviderTavily Tavily 提供商
	ProviderTavily CommunityToolProvider = "tavily"
	// ProviderJinaAI Jina AI 提供商
	ProviderJinaAI CommunityToolProvider = "jina_ai"
	// ProviderFirecrawl Firecrawl 提供商
	ProviderFirecrawl CommunityToolProvider = "firecrawl"
	// ProviderImageSearch Image Search 提供商
	ProviderImageSearch CommunityToolProvider = "image_search"
)

// CommunityToolConfig 社区工具配置
type CommunityToolConfig struct {
	// Provider 工具提供商
	Provider CommunityToolProvider
	// APIKey API 密钥（如需要）
	APIKey string
	// MaxResults 最大结果数
	MaxResults int
	// Timeout 超时时间（秒）
	Timeout int
}

// GetCommunityTools 获取社区工具
// 一比一复刻 DeerFlow 的社区工具组装
func GetCommunityTools(config *CommunityToolConfig) ([]tool.BaseTool, error) {
	if config == nil {
		return nil, fmt.Errorf("community tool config is required")
	}

	var tools []tool.BaseTool

	switch config.Provider {
	case ProviderTavily:
		// Tavily：提供 web_search 和 web_fetch
		tools = append(tools, tavily.NewWebSearchTool(config.APIKey, config.MaxResults))
		tools = append(tools, tavily.NewWebFetchTool(config.APIKey))

	case ProviderJinaAI:
		// Jina AI：只提供 web_fetch
		tools = append(tools, jinaai.NewWebFetchTool(config.APIKey, config.Timeout))

	case ProviderFirecrawl:
		// Firecrawl：提供 web_search 和 web_fetch
		tools = append(tools, firecrawl.NewWebSearchTool(config.APIKey, config.MaxResults))
		tools = append(tools, firecrawl.NewWebFetchTool(config.APIKey))

	case ProviderImageSearch:
		// Image Search：只提供 image_search
		tools = append(tools, imagesearch.NewImageSearchTool(config.MaxResults))

	default:
		return nil, fmt.Errorf("unknown community tool provider: %s", config.Provider)
	}

	return tools, nil
}

// GetAllCommunityTools 获取所有社区工具（用于测试）
func GetAllCommunityTools() ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	// Tavily（无 API key 模式）
	tools = append(tools, tavily.NewWebSearchTool("", 5))
	tools = append(tools, tavily.NewWebFetchTool(""))

	// Jina AI（无 API key 模式）
	tools = append(tools, jinaai.NewWebFetchTool("", 10))

	// Firecrawl（无 API key 模式）
	tools = append(tools, firecrawl.NewWebSearchTool("", 5))
	tools = append(tools, firecrawl.NewWebFetchTool(""))

	// Image Search
	tools = append(tools, imagesearch.NewImageSearchTool(5))

	return tools, nil
}

// ToolProviderNames 获取所有支持的工具提供商名称
func ToolProviderNames() []string {
	return []string{
		string(ProviderTavily),
		string(ProviderJinaAI),
		string(ProviderFirecrawl),
		string(ProviderImageSearch),
	}
}
