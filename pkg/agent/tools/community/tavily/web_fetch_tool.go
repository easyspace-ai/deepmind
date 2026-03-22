package tavily

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ============================================
// Tavily Web Fetch Tool（一比一复刻 DeerFlow）
// ============================================

// WebFetchTool web_fetch 工具
type WebFetchTool struct {
	apiKey string
}

// NewWebFetchTool 创建 web_fetch 工具
func NewWebFetchTool(apiKey string) *WebFetchTool {
	return &WebFetchTool{
		apiKey: apiKey,
	}
}

// Name 返回工具名称
func (t *WebFetchTool) Name() string {
	return "web_fetch"
}

// Info 返回工具信息
func (t *WebFetchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: `Fetch the contents of a web page at a given URL.
Only fetch EXACT URLs that have been provided directly by the user or have been returned in results from the web_search and web_fetch tools.
This tool can NOT access content that requires authentication, such as private Google Docs or pages behind login walls.
Do NOT add www. to URLs that do NOT have them.
URLs must include the schema: https://example.com is a valid URL while example.com is an invalid URL.`,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type:     schema.DataType("string"),
				Desc:     "The URL to fetch the contents of.",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具
func (t *WebFetchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if args.URL == "" {
		return "No URL provided.", nil
	}

	client := NewClient(t.apiKey)
	resp, err := client.Extract([]string{args.URL})
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	// 一比一复刻 DeerFlow 的错误处理和返回格式
	if len(resp.FailedResults) > 0 {
		return fmt.Sprintf("Error: %s", resp.FailedResults[0].Error), nil
	}

	if len(resp.Results) > 0 {
		result := resp.Results[0]
		content := result.RawContent
		if len(content) > 4096 {
			content = content[:4096]
		}
		return fmt.Sprintf("# %s\n\n%s", result.Title, content), nil
	}

	return "Error: No results found", nil
}

// Invoke 实现 Eino tool.BaseTool 接口
func (t *WebFetchTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	url, _ := args["url"].(string)
	if url == "" {
		return "No URL provided.", nil
	}

	client := NewClient(t.apiKey)
	resp, err := client.Extract([]string{url})
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	if len(resp.FailedResults) > 0 {
		return fmt.Sprintf("Error: %s", resp.FailedResults[0].Error), nil
	}

	if len(resp.Results) > 0 {
		result := resp.Results[0]
		content := result.RawContent
		if len(content) > 4096 {
			content = content[:4096]
		}
		return fmt.Sprintf("# %s\n\n%s", result.Title, content), nil
	}

	return "Error: No results found", nil
}
