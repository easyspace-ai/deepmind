package deerflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	"github.com/weibaohui/nanobot-go/pkg/agent/prompts"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/sandbox"
)

// ToolGroup 工具组类型
type ToolGroup string

const (
	ToolGroupSandbox   ToolGroup = "sandbox"
	ToolGroupBuiltin   ToolGroup = "builtin"
	ToolGroupMCP       ToolGroup = "mcp"
	ToolGroupCommunity ToolGroup = "community"
	ToolGroupSubagent  ToolGroup = "subagent"
)

// ToolConfig 工具配置
type ToolConfig struct {
	// Sandbox 沙箱实例
	Sandbox sandbox.Sandbox
	// ThreadState 线程状态
	ThreadState *state.ThreadState
	// SubagentEnabled 是否启用子代理
	SubagentEnabled bool
	// MaxConcurrentSubagents 最大并发子代理数
	MaxConcurrentSubagents int
}

// GetAvailableTools 获取可用工具
// 一比一复刻 DeerFlow 的 get_available_tools
func GetAvailableTools(config *ToolConfig, groups []ToolGroup) ([]tool.BaseTool, error) {
	if config == nil {
		config = &ToolConfig{}
	}

	var tools []tool.BaseTool

	// 如果没有指定组，默认返回所有组
	if len(groups) == 0 {
		groups = []ToolGroup{
			ToolGroupSandbox,
			ToolGroupBuiltin,
		}
		if config.SubagentEnabled {
			groups = append(groups, ToolGroupSubagent)
		}
	}

	for _, group := range groups {
		groupTools, err := getToolsByGroup(group, config)
		if err != nil {
			return nil, fmt.Errorf("get tools for group %s failed: %w", group, err)
		}
		tools = append(tools, groupTools...)
	}

	return tools, nil
}

// getToolsByGroup 根据组获取工具
func getToolsByGroup(group ToolGroup, config *ToolConfig) ([]tool.BaseTool, error) {
	switch group {
	case ToolGroupSandbox:
		return getSandboxTools(config)
	case ToolGroupBuiltin:
		return getBuiltinTools(config)
	case ToolGroupCommunity:
		return getCommunityTools(config)
	case ToolGroupSubagent:
		return getSubagentTools(config)
	default:
		return nil, nil
	}
}

// getCommunityTools 获取社区工具
// 注意：社区工具需要额外的配置，这里返回空列表
// 使用 community.GetCommunityTools() 来获取具体的社区工具
func getCommunityTools(config *ToolConfig) ([]tool.BaseTool, error) {
	// Community tools 需要单独配置提供商和 API key
	// 这里返回空列表，具体使用请参考 pkg/agent/tools/community/
	return nil, nil
}

// getSandboxTools 获取 Sandbox 工具
func getSandboxTools(config *ToolConfig) ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	// 如果有沙箱实例，创建带路径翻译的工具
	if config.Sandbox != nil {
		tools = append(tools, NewBashTool(config.Sandbox))
		tools = append(tools, NewLsTool(config.Sandbox))
		tools = append(tools, NewReadFileTool(config.Sandbox))
		tools = append(tools, NewWriteFileTool(config.Sandbox))
		tools = append(tools, NewStrReplaceTool(config.Sandbox))
	} else {
		// 没有沙箱时返回基础版本
		tools = append(tools, NewBasicBashTool())
		tools = append(tools, NewBasicLsTool())
		tools = append(tools, NewBasicReadFileTool())
		tools = append(tools, NewBasicWriteFileTool())
		tools = append(tools, NewBasicStrReplaceTool())
	}

	return tools, nil
}

// getBuiltinTools 获取内置工具
func getBuiltinTools(config *ToolConfig) ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	tools = append(tools, NewPresentFilesTool(config))
	tools = append(tools, NewAskClarificationTool())
	tools = append(tools, NewViewImageTool(config))
	tools = append(tools, NewWriteTodosTool(config))

	return tools, nil
}

// getSubagentTools 获取子代理工具
func getSubagentTools(config *ToolConfig) ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	if config.SubagentEnabled {
		tools = append(tools, NewTaskTool(nil))
	}

	return tools, nil
}

// ============================================
// 工具基础结构
// ============================================

// BaseDeerFlowTool DeerFlow 工具基类
type BaseDeerFlowTool struct {
	name        string
	description string
	paramsSchema map[string]interface{}
}

// NewBaseDeerFlowTool 创建 DeerFlow 工具基类
func NewBaseDeerFlowTool(name, description string, paramsSchema map[string]interface{}) *BaseDeerFlowTool {
	return &BaseDeerFlowTool{
		name:        name,
		description: description,
		paramsSchema: paramsSchema,
	}
}

// Info 实现 tool.BaseTool 接口
func (t *BaseDeerFlowTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	info := &schema.ToolInfo{
		Name: t.name,
		Desc: t.description,
	}
	if len(t.paramsSchema) == 0 {
		return info, nil
	}
	raw, err := json.Marshal(map[string]interface{}{
		"type":       "object",
		"properties": t.paramsSchema,
	})
	if err != nil {
		return nil, fmt.Errorf("deerflow tool %q params: %w", t.name, err)
	}
	var js jsonschema.Schema
	if err := json.Unmarshal(raw, &js); err != nil {
		return nil, fmt.Errorf("deerflow tool %q jsonschema: %w", t.name, err)
	}
	info.ParamsOneOf = schema.NewParamsOneOfByJSONSchema(&js)
	return info, nil
}

// 具体工具构造函数见 sandbox_tools.go、builtin_tools.go、task_tool.go。

// ============================================
// 预设函数
// ============================================

// BuildLeadAgentTools 构建 Lead Agent 工具列表
func BuildLeadAgentTools(config *ToolConfig) ([]tool.BaseTool, error) {
	return GetAvailableTools(config, nil)
}

// BuildGeneralPurposeSubagentTools 构建通用子代理工具列表
func BuildGeneralPurposeSubagentTools(config *ToolConfig) ([]tool.BaseTool, error) {
	// 通用子代理：所有工具除了 task()
	return GetAvailableTools(config, []ToolGroup{
		ToolGroupSandbox,
		ToolGroupBuiltin,
	})
}

// BuildBashSubagentTools 构建 Bash 子代理工具列表
func BuildBashSubagentTools(config *ToolConfig) ([]tool.BaseTool, error) {
	// Bash 子代理：只有 Sandbox 工具
	return GetAvailableTools(config, []ToolGroup{
		ToolGroupSandbox,
	})
}

// BuildLeadAgentPromptWithTools 构建带工具提示的 Lead Agent 提示词
func BuildLeadAgentPromptWithTools(config *ToolConfig, agentConfig *prompts.LeadAgentConfig) (*prompts.Prompt, []tool.BaseTool, error) {
	// 构建工具列表
	tools, err := BuildLeadAgentTools(config)
	if err != nil {
		return nil, nil, err
	}

	// 构建提示词
	prompt := prompts.BuildLeadAgentPrompt(agentConfig)

	return prompt, tools, nil
}
