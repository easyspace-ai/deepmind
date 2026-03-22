package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ============================================
// 配置常量
// ============================================

const (
	// ConfigEnvVar 配置文件路径环境变量
	ConfigEnvVar = "DEER_FLOW_CONFIG_PATH"
	// ConfigFileName 配置文件名
	ConfigFileName = "config.yaml"
	// ConfigExampleFileName 示例配置文件名
	ConfigExampleFileName = "config.example.yaml"
)

// ============================================
// AppConfig - 主配置
// ============================================

// AppConfig 应用主配置
// 一比一复刻 DeerFlow 的 AppConfig
type AppConfig struct {
	// ConfigVersion 配置版本
	ConfigVersion int `yaml:"config_version,omitempty" json:"config_version,omitempty"`

	// Models 模型配置列表
	Models []ModelConfig `yaml:"models,omitempty" json:"models,omitempty"`

	// Sandbox 沙箱配置
	Sandbox SandboxConfig `yaml:"sandbox" json:"sandbox"`

	// Tools 工具配置列表
	Tools []ToolConfig `yaml:"tools,omitempty" json:"tools,omitempty"`

	// ToolGroups 工具分组配置
	ToolGroups []ToolGroupConfig `yaml:"tool_groups,omitempty" json:"tool_groups,omitempty"`

	// Skills 技能配置
	Skills SkillsConfig `yaml:"skills,omitempty" json:"skills,omitempty"`

	// Extensions 扩展配置（MCP + Skills 状态）
	Extensions ExtensionsConfig `yaml:"extensions,omitempty" json:"extensions,omitempty"`

	// ToolSearch 工具搜索配置
	ToolSearch ToolSearchConfig `yaml:"tool_search,omitempty" json:"tool_search,omitempty"`

	// Title 标题生成配置
	Title TitleConfig `yaml:"title,omitempty" json:"title,omitempty"`

	// Summarization 摘要配置
	Summarization SummarizationConfig `yaml:"summarization,omitempty" json:"summarization,omitempty"`

	// Subagents 子代理配置
	Subagents SubagentsConfig `yaml:"subagents,omitempty" json:"subagents,omitempty"`

	// Memory 记忆配置
	Memory MemoryConfig `yaml:"memory,omitempty" json:"memory,omitempty"`

	// Checkpointer 检查点配置
	Checkpointer *CheckpointerConfig `yaml:"checkpointer,omitempty" json:"checkpointer,omitempty"`
}

// DefaultAppConfig 默认应用配置
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ConfigVersion: 1,
		Sandbox:       DefaultSandboxConfig(),
		Skills:        DefaultSkillsConfig(),
		Extensions:    DefaultExtensionsConfig(),
		ToolSearch:    DefaultToolSearchConfig(),
		Title:         DefaultTitleConfig(),
		Summarization: DefaultSummarizationConfig(),
		Subagents:     DefaultSubagentsConfig(),
		Memory:        DefaultMemoryConfig(),
	}
}

// ============================================
// ModelConfig - 模型配置
// ============================================

// ModelConfig 模型配置
type ModelConfig struct {
	// Name 模型名称
	Name string `yaml:"name" json:"name"`

	// Use 实现类路径（反射加载）
	Use string `yaml:"use" json:"use"`

	// DisplayName 显示名称
	DisplayName string `yaml:"display_name,omitempty" json:"display_name,omitempty"`

	// SupportsThinking 是否支持思考模式
	SupportsThinking bool `yaml:"supports_thinking,omitempty" json:"supports_thinking,omitempty"`

	// SupportsVision 是否支持视觉
	SupportsVision bool `yaml:"supports_vision,omitempty" json:"supports_vision,omitempty"`

	// WhenThinkingEnabled 思考模式启用时的覆盖配置
	WhenThinkingEnabled map[string]any `yaml:"when_thinking_enabled,omitempty" json:"when_thinking_enabled,omitempty"`

	// 其他配置字段（由具体 provider 定义）
	Extra map[string]any `yaml:",inline,omitempty" json:"extra,omitempty"`
}

// ============================================
// SandboxConfig - 沙箱配置
// ============================================

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	// Use 沙箱 provider 类路径
	Use string `yaml:"use" json:"use"`

	// 其他配置字段（由具体 provider 定义）
	Extra map[string]any `yaml:",inline,omitempty" json:"extra,omitempty"`
}

// DefaultSandboxConfig 默认沙箱配置
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Use: "deerflow.sandbox.local.LocalSandboxProvider",
	}
}

// ============================================
// ToolConfig - 工具配置
// ============================================

// ToolConfig 工具配置
type ToolConfig struct {
	// Name 工具名称
	Name string `yaml:"name" json:"name"`

	// Use 工具实现变量路径
	Use string `yaml:"use" json:"use"`

	// Group 工具分组
	Group string `yaml:"group,omitempty" json:"group,omitempty"`

	// 其他配置字段
	Extra map[string]any `yaml:",inline,omitempty" json:"extra,omitempty"`
}

// ============================================
// ToolGroupConfig - 工具分组配置
// ============================================

// ToolGroupConfig 工具分组配置
type ToolGroupConfig struct {
	// Name 分组名称
	Name string `yaml:"name" json:"name"`

	// Tools 工具名称列表
	Tools []string `yaml:"tools" json:"tools"`

	// Description 描述
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// ============================================
// SkillsConfig - 技能配置
// ============================================

// SkillsConfig 技能配置
type SkillsConfig struct {
	// Path 技能目录路径（主机）
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// ContainerPath 容器路径
	ContainerPath string `yaml:"container_path,omitempty" json:"container_path,omitempty"`
}

// DefaultSkillsConfig 默认技能配置
func DefaultSkillsConfig() SkillsConfig {
	return SkillsConfig{
		ContainerPath: "/mnt/skills",
	}
}

// ============================================
// ExtensionsConfig - 扩展配置
// ============================================

// ExtensionsConfig 扩展配置（MCP + Skills 状态）
type ExtensionsConfig struct {
	// MCPServers MCP 服务器配置
	MCPServers map[string]MCPServerConfig `yaml:"mcpServers,omitempty" json:"mcpServers,omitempty"`

	// Skills 技能启用状态
	Skills map[string]SkillState `yaml:"skills,omitempty" json:"skills,omitempty"`
}

// DefaultExtensionsConfig 默认扩展配置
func DefaultExtensionsConfig() ExtensionsConfig {
	return ExtensionsConfig{
		MCPServers: make(map[string]MCPServerConfig),
		Skills:     make(map[string]SkillState),
	}
}

// MCPServerConfig MCP 服务器配置
type MCPServerConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Type 类型: stdio, sse, http
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Command 命令（stdio）
	Command string `yaml:"command,omitempty" json:"command,omitempty"`

	// Args 参数（stdio）
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`

	// Env 环境变量
	Env map[string]string `yaml:"env,omitempty" json:"env,omitempty"`

	// URL URL（sse/http）
	URL string `yaml:"url,omitempty" json:"url,omitempty"`

	// Headers 请求头（sse/http）
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// OAuth OAuth 配置
	OAuth *OAuthConfig `yaml:"oauth,omitempty" json:"oauth,omitempty"`

	// Description 描述
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// OAuthConfig OAuth 配置
// 一比一复刻 DeerFlow 的 McpOAuthConfig
type OAuthConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// ClientID 客户端 ID
	ClientID string `yaml:"client_id" json:"client_id"`

	// ClientSecret 客户端密钥
	ClientSecret string `yaml:"client_secret" json:"client_secret"`

	// TokenURL Token URL
	TokenURL string `yaml:"token_url" json:"token_url"`

	// Scopes 权限范围
	Scopes string `yaml:"scope,omitempty" json:"scope,omitempty"`

	// Audience Audience
	Audience string `yaml:"audience,omitempty" json:"audience,omitempty"`

	// GrantType 授权类型: client_credentials, refresh_token
	GrantType string `yaml:"grant_type,omitempty" json:"grant_type,omitempty"`

	// RefreshToken 刷新令牌（refresh_token 流程）
	RefreshToken string `yaml:"refresh_token,omitempty" json:"refresh_token,omitempty"`

	// TokenField Token 字段名（默认 access_token）
	TokenField string `yaml:"token_field,omitempty" json:"token_field,omitempty"`

	// TokenTypeField Token 类型字段名（默认 token_type）
	TokenTypeField string `yaml:"token_type_field,omitempty" json:"token_type_field,omitempty"`

	// ExpiresInField 过期时间字段名（默认 expires_in）
	ExpiresInField string `yaml:"expires_in_field,omitempty" json:"expires_in_field,omitempty"`

	// DefaultTokenType 默认 Token 类型（默认 Bearer）
	DefaultTokenType string `yaml:"default_token_type,omitempty" json:"default_token_type,omitempty"`

	// RefreshSkewSeconds 刷新提前量（秒）
	RefreshSkewSeconds int `yaml:"refresh_skew_seconds,omitempty" json:"refresh_skew_seconds,omitempty"`

	// ExtraTokenParams 额外的 Token 请求参数
	ExtraTokenParams map[string]string `yaml:"extra_token_params,omitempty" json:"extra_token_params,omitempty"`
}

// SkillState 技能状态
type SkillState struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// ============================================
// ToolSearchConfig - 工具搜索配置
// ============================================

// ToolSearchConfig 工具搜索配置
type ToolSearchConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// DefaultToolSearchConfig 默认工具搜索配置
func DefaultToolSearchConfig() ToolSearchConfig {
	return ToolSearchConfig{
		Enabled: false,
	}
}

// ============================================
// TitleConfig - 标题生成配置
// ============================================

// TitleConfig 标题生成配置
type TitleConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`

	// MaxWords 最大词数
	MaxWords int `yaml:"max_words,omitempty" json:"max_words,omitempty"`

	// MaxChars 最大字符数
	MaxChars int `yaml:"max_chars,omitempty" json:"max_chars,omitempty"`

	// PromptTemplate 提示词模板
	PromptTemplate string `yaml:"prompt_template,omitempty" json:"prompt_template,omitempty"`
}

// DefaultTitleConfig 默认标题生成配置
func DefaultTitleConfig() TitleConfig {
	return TitleConfig{
		Enabled:   true,
		MaxWords:  10,
		MaxChars:  100,
	}
}

// ============================================
// SummarizationConfig - 摘要配置
// ============================================

// ContextSizeType 上下文大小类型
type ContextSizeType string

const (
	// ContextSizeTypeFraction 按比例
	ContextSizeTypeFraction ContextSizeType = "fraction"
	// ContextSizeTypeTokens 按 token 数
	ContextSizeTypeTokens ContextSizeType = "tokens"
	// ContextSizeTypeMessages 按消息数
	ContextSizeTypeMessages ContextSizeType = "messages"
)

// ContextSize 上下文大小规格
// 一比一复刻 DeerFlow 的 ContextSize
type ContextSize struct {
	// Type 类型: fraction, tokens, messages
	Type ContextSizeType `yaml:"type" json:"type"`
	// Value 值
	Value any `yaml:"value" json:"value"`
}

// SummarizationConfig 摘要配置
// 一比一复刻 DeerFlow 的 SummarizationConfig
type SummarizationConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Trigger 触发条件（单个或多个）
	Trigger any `yaml:"trigger,omitempty" json:"trigger,omitempty"` // 可以是 *ContextSize 或 []*ContextSize

	// Keep 保留策略
	Keep *ContextSize `yaml:"keep,omitempty" json:"keep,omitempty"`

	// ModelName 模型名称
	ModelName string `yaml:"model_name,omitempty" json:"model_name,omitempty"`

	// TrimTokensToSummarize 要摘要的 token 数
	TrimTokensToSummarize *int `yaml:"trim_tokens_to_summarize,omitempty" json:"trim_tokens_to_summarize,omitempty"`

	// SummaryPrompt 摘要提示词
	SummaryPrompt string `yaml:"summary_prompt,omitempty" json:"summary_prompt,omitempty"`
}

// DefaultSummarizationConfig 默认摘要配置
// 一比一复刻 DeerFlow 的默认值
func DefaultSummarizationConfig() SummarizationConfig {
	defaultTrim := 4000
	return SummarizationConfig{
		Enabled: false,
		Keep: &ContextSize{
			Type:  ContextSizeTypeMessages,
			Value: 20,
		},
		TrimTokensToSummarize: &defaultTrim,
	}
}

// SummarizationTrigger 摘要触发条件（向后兼容）
type SummarizationTrigger struct {
	// Type 触发类型: tokens, messages, fraction
	Type string `yaml:"type" json:"type"`

	// Value 触发值
	Value any `yaml:"value" json:"value"`
}

// SummarizationKeep 摘要保留策略（向后兼容）
type SummarizationKeep struct {
	// Type 保留类型: keep.last_n, keep.first_n, keep.fraction
	Type string `yaml:"type" json:"type"`

	// Value 保留值
	Value any `yaml:"value" json:"value"`
}

// ============================================
// AgentsConfig - Agent 配置
// ============================================

// AgentConfig 自定义 Agent 配置
// 一比一复刻 DeerFlow 的 AgentConfig
type AgentConfig struct {
	// Name 名称
	Name string `yaml:"name" json:"name"`
	// Description 描述
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Model 模型名称
	Model string `yaml:"model,omitempty" json:"model,omitempty"`
	// ToolGroups 工具分组
	ToolGroups []string `yaml:"tool_groups,omitempty" json:"tool_groups,omitempty"`
}

// AgentsConfig Agent 配置管理（全局实例）
type AgentsConfig struct {
	// agents 按名称存储的 agent 配置
	agents map[string]*AgentConfig
}

// NewAgentsConfig 创建 Agent 配置管理
func NewAgentsConfig() *AgentsConfig {
	return &AgentsConfig{
		agents: make(map[string]*AgentConfig),
	}
}

// GetAgentConfig 获取 agent 配置
func (c *AgentsConfig) GetAgentConfig(name string) *AgentConfig {
	if c == nil || c.agents == nil {
		return nil
	}
	return c.agents[name]
}

// SetAgentConfig 设置 agent 配置
func (c *AgentsConfig) SetAgentConfig(name string, cfg *AgentConfig) {
	if c == nil {
		return
	}
	if c.agents == nil {
		c.agents = make(map[string]*AgentConfig)
	}
	c.agents[name] = cfg
}

// ListAgents 列出所有 agent 配置
func (c *AgentsConfig) ListAgents() []*AgentConfig {
	if c == nil || c.agents == nil {
		return nil
	}
	result := make([]*AgentConfig, 0, len(c.agents))
	for _, cfg := range c.agents {
		result = append(result, cfg)
	}
	return result
}

// ============================================
// SubagentsConfig - 子代理配置
// ============================================

// SubagentsConfig 子代理配置
type SubagentsConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Timeouts 超时配置（子代理名称 -> 秒数）
	Timeouts map[string]int `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`
}

// DefaultSubagentsConfig 默认子代理配置
func DefaultSubagentsConfig() SubagentsConfig {
	return SubagentsConfig{
		Enabled:  true,
		Timeouts: make(map[string]int),
	}
}

// GetTimeoutFor 获取子代理超时配置
func (c *SubagentsConfig) GetTimeoutFor(name string) int {
	if timeout, ok := c.Timeouts[name]; ok {
		return timeout
	}
	return 900 // 默认 15 分钟
}

// ============================================
// MemoryConfig - 记忆配置
// ============================================

// MemoryConfig 记忆配置
type MemoryConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`

	// InjectionEnabled 是否注入到提示词
	InjectionEnabled bool `yaml:"injection_enabled,omitempty" json:"injection_enabled,omitempty"`

	// StoragePath 存储路径
	StoragePath string `yaml:"storage_path,omitempty" json:"storage_path,omitempty"`

	// DebounceSeconds 去抖秒数
	DebounceSeconds int `yaml:"debounce_seconds,omitempty" json:"debounce_seconds,omitempty"`

	// ModelName 模型名称
	ModelName string `yaml:"model_name,omitempty" json:"model_name,omitempty"`

	// MaxFacts 最大事实数
	MaxFacts int `yaml:"max_facts,omitempty" json:"max_facts,omitempty"`

	// FactConfidenceThreshold 事实置信度阈值
	FactConfidenceThreshold float64 `yaml:"fact_confidence_threshold,omitempty" json:"fact_confidence_threshold,omitempty"`

	// MaxInjectionTokens 最大注入 token 数
	MaxInjectionTokens int `yaml:"max_injection_tokens,omitempty" json:"max_injection_tokens,omitempty"`
}

// DefaultMemoryConfig 默认记忆配置
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{
		Enabled:                false,
		InjectionEnabled:       true,
		DebounceSeconds:        30,
		MaxFacts:               100,
		FactConfidenceThreshold: 0.7,
		MaxInjectionTokens:     2000,
	}
}

// ============================================
// CheckpointerConfig - 检查点配置
// ============================================

// CheckpointerConfig 检查点配置
type CheckpointerConfig struct {
	// Use 检查点实现类路径
	Use string `yaml:"use" json:"use"`

	// 其他配置字段
	Extra map[string]any `yaml:",inline,omitempty" json:"extra,omitempty"`
}

// ============================================
// 环境变量解析
// ============================================

// ResolveEnvVariable 解析环境变量
// 支持 $VAR 或 ${VAR} 格式
func ResolveEnvVariable(value string) string {
	if !strings.HasPrefix(value, "$") {
		return value
	}

	// 移除 $ 前缀
	varName := strings.TrimPrefix(value, "$")

	// 移除 {} 包围（如果有）
	varName = strings.TrimPrefix(varName, "{")
	varName = strings.TrimSuffix(varName, "}")

	// 返回环境变量值，如果不存在则返回原值
	if envValue := os.Getenv(varName); envValue != "" {
		return envValue
	}
	return value
}

// ============================================
// 配置路径解析
// ============================================

// ResolveConfigPath 解析配置文件路径
// 优先级：
// 1. configPath 参数（如果提供）
// 2. DEER_FLOW_CONFIG_PATH 环境变量
// 3. 当前目录的 config.yaml
// 4. 父目录的 config.yaml
func ResolveConfigPath(configPath string) (string, error) {
	// 1. 参数提供的路径
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
		return "", os.ErrNotExist
	}

	// 2. 环境变量
	if envPath := os.Getenv(ConfigEnvVar); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
		return "", os.ErrNotExist
	}

	// 3. 当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	localPath := filepath.Join(currentDir, ConfigFileName)
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// 4. 父目录
	parentDir := filepath.Dir(currentDir)
	parentPath := filepath.Join(parentDir, ConfigFileName)
	if _, err := os.Stat(parentPath); err == nil {
		return parentPath, nil
	}

	return "", os.ErrNotExist
}
