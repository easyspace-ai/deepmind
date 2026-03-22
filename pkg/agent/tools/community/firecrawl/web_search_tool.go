package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ============================================
// Firecrawl Web Search Tool（一比一复刻 DeerFlow）
// ============================================

// WebSearchTool web_search 工具
type WebSearchTool struct {
	apiKey     string
	maxResults int
}

// NewWebSearchTool 创建 web_search 工具
func NewWebSearchTool(apiKey string, maxResults int) *WebSearchTool {
	if maxResults <= 0 {
		maxResults = 5
	}
	return &WebSearchTool{
		apiKey:     apiKey,
		maxResults: maxResults,
	}
}

// Name 返回工具名称
func (t *WebSearchTool) Name() string {
	return "web_search"
}

// Info 返回工具信息
func (t *WebSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: "Search the web.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.DataType("string"),
				Desc:     "The query to search for.",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具
func (t *WebSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if args.Query == "" {
		return "No query provided.", nil
	}

	client := NewClient(t.apiKey)
	resp, err := client.Search(args.Query, t.maxResults)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	// 归一化结果（一比一复刻 DeerFlow）
	webResults := resp.Web
	normalizedResults := make([]map[string]interface{}, 0, len(webResults))
	for _, item := range webResults {
		normalizedResults = append(normalizedResults, map[string]interface{}{
			"title":   item.Title,
			"url":     item.URL,
			"snippet": item.Description,
		})
	}

	jsonResults, err := json.MarshalIndent(normalizedResults, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	return string(jsonResults), nil
}

// Invoke 实现 Eino tool.BaseTool 接口
func (t *WebSearchTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "No query provided.", nil
	}

	client := NewClient(t.apiKey)
	resp, err := client.Search(query, t.maxResults)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	webResults := resp.Web
	normalizedResults := make([]map[string]interface{}, 0, len(webResults))
	for _, item := range webResults {
		normalizedResults = append(normalizedResults, map[string]interface{}{
			"title":   item.Title,
			"url":     item.URL,
			"snippet": item.Description,
		})
	}

	return normalizedResults, nil
}
