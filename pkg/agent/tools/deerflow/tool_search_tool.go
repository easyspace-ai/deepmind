package deerflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
)

// ============================================
// ToolSearch 工具（一比一复刻 DeerFlow）
// ============================================

// ToolSearchTool tool_search 工具
type ToolSearchTool struct {
	*BaseDeerFlowTool
}

// NewToolSearchTool 创建 tool_search 工具
func NewToolSearchTool() tool.BaseTool {
	return &ToolSearchTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"tool_search",
			"Fetches full schema definitions for deferred tools so they can be called.\n\nDeferred tools appear by name in <available-deferred-tools> in the system prompt. Until fetched, only the name is known — there is no parameter schema, so the tool cannot be invoked. This tool takes a query, matches it against the deferred tool list, and returns the matched tools' complete definitions. Once a tool's schema appears in that result, it is callable.\n\nQuery forms:\n  - \"select:Read,Edit,Grep\" — fetch these exact tools by name\n  - \"notebook jupyter\" — keyword search, up to max_results best matches\n  - \"+slack send\" — require \"slack\" in the name, rank by remaining terms",
			map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Query to find deferred tools. Use \"select:<tool_name> for direct selection, or keywords to search.",
				},
			},
		),
	}
}

// Invoke 执行 tool_search 工具
func (t *ToolSearchTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "No query provided.", nil
	}

	registry := GetDeferredRegistry()
	if registry == nil {
		return "No deferred tools available.", nil
	}

	matchedTools := registry.Search(query)
	if len(matchedTools) == 0 {
		return fmt.Sprintf("No tools found matching: %s", query), nil
	}

	// 转换为 OpenAI function 格式
	toolDefs := make([]map[string]interface{}, 0, len(matchedTools))
	for _, toolObj := range matchedTools {
		toolDef, err := toolToOpenAIFunction(ctx, toolObj)
		if err == nil && toolDef != nil {
			toolDefs = append(toolDefs, toolDef)
		}
		if len(toolDefs) >= MaxResults {
			break
		}
	}

	result, err := json.MarshalIndent(toolDefs, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	return string(result), nil
}

// toolToOpenAIFunction 将工具转换为 OpenAI function 格式
func toolToOpenAIFunction(ctx context.Context, t tool.BaseTool) (map[string]interface{}, error) {
	info, err := t.Info(ctx)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"name":        info.Name,
		"description": info.Desc,
	}

	// 处理参数 schema
	if info.ParamsOneOf != nil {
		params := make(map[string]interface{})
		// 简化处理：从 ParamsOneOf 提取 schema
		// 这里使用简化的 schema 结构
		params["type"] = "object"
		params["properties"] = map[string]interface{}{}
		result["parameters"] = params
	} else {
		// 默认空参数
		result["parameters"] = map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	return result, nil
}
