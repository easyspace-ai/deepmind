package prompts

// Builder 提示词构建器
// 一比一复刻 DeerFlow 的 Builder 模式
type Builder struct {
	prompt *Prompt
}

// NewBuilder 创建新的构建器
func NewBuilder() *Builder {
	return &Builder{
		prompt: NewPrompt(),
	}
}

// Add 添加任意分段
func (b *Builder) Add(section PromptSection) *Builder {
	b.prompt.AddSection(section)
	return b
}

// WithRole 添加角色分段
func (b *Builder) WithRole(agentName string) *Builder {
	b.prompt.AddSection(NewRoleSection(agentName))
	return b
}

// WithSoul 添加 SOUL 分段
func (b *Builder) WithSoul(soulContent string) *Builder {
	if soulContent != "" {
		b.prompt.AddSection(NewSoulSection(soulContent))
	}
	return b
}

// WithMemory 添加记忆分段
func (b *Builder) WithMemory(memoryContent string) *Builder {
	if memoryContent != "" {
		b.prompt.AddSection(NewMemorySection(memoryContent))
	}
	return b
}

// WithThinkingStyle 添加思考方式分段
func (b *Builder) WithThinkingStyle() *Builder {
	b.prompt.AddSection(NewThinkingStyleSection(false, 3))
	return b
}

// WithThinkingStyleWithSubagent 添加带子代理的思考方式分段
func (b *Builder) WithThinkingStyleWithSubagent(subagentLimit int) *Builder {
	b.prompt.AddSection(NewThinkingStyleSection(true, subagentLimit))
	return b
}

// WithClarification 添加澄清询问分段
func (b *Builder) WithClarification() *Builder {
	b.prompt.AddSection(NewClarificationSection())
	return b
}

// WithSubagent 添加子代理分段
func (b *Builder) WithSubagent(maxConcurrent int) *Builder {
	b.prompt.AddSection(NewSubagentSection(maxConcurrent))
	return b
}

// WithSkills 添加技能列表分段
func (b *Builder) WithSkills(skillsContent string) *Builder {
	if skillsContent != "" {
		b.prompt.AddSection(NewSkillsSection(skillsContent))
	}
	return b
}

// WithDeferredTools 添加延迟工具列表分段
func (b *Builder) WithDeferredTools(toolsContent string) *Builder {
	if toolsContent != "" {
		b.prompt.AddSection(NewDeferredToolsSection(toolsContent))
	}
	return b
}

// WithWorkingDir 添加工作目录分段
func (b *Builder) WithWorkingDir() *Builder {
	b.prompt.AddSection(NewWorkingDirSection())
	return b
}

// WithResponseStyle 添加响应风格分段
func (b *Builder) WithResponseStyle() *Builder {
	b.prompt.AddSection(NewResponseStyleSection())
	return b
}

// WithCitations 添加引用格式分段
func (b *Builder) WithCitations() *Builder {
	b.prompt.AddSection(NewCitationsSection())
	return b
}

// WithCriticalReminders 添加关键提醒分段
func (b *Builder) WithCriticalReminders() *Builder {
	b.prompt.AddSection(NewCriticalRemindersSection(false, 3))
	return b
}

// WithCriticalRemindersWithSubagent 添加带子代理的关键提醒分段
func (b *Builder) WithCriticalRemindersWithSubagent(subagentLimit int) *Builder {
	b.prompt.AddSection(NewCriticalRemindersSection(true, subagentLimit))
	return b
}

// WithCurrentDate 添加当前日期分段
func (b *Builder) WithCurrentDate() *Builder {
	b.prompt.AddSection(NewCurrentDateSection())
	return b
}

// WithEnvironment 添加环境信息分段
func (b *Builder) WithEnvironment() *Builder {
	b.prompt.AddSection(NewEnvironmentSection())
	return b
}

// WithEnvInfo 添加环境信息分段（向后兼容）
func (b *Builder) WithEnvInfo() *Builder {
	return b.WithEnvironment()
}

// WithCustom 添加自定义分段
func (b *Builder) WithCustom(name, tagName, content string) *Builder {
	if content != "" {
		b.prompt.AddSection(NewCustomSection(name, tagName, content))
	}
	return b
}

// WithSection 添加任意分段
func (b *Builder) WithSection(section PromptSection) *Builder {
	b.prompt.AddSection(section)
	return b
}

// Build 构建完整提示词
func (b *Builder) Build() *Prompt {
	return b.prompt
}

// BuildString 构建并渲染为字符串
func (b *Builder) BuildString() string {
	return b.prompt.Render()
}

// ============================================
// 预设提示词构建函数
// ============================================

// BuildLeadAgentPrompt 构建 Lead Agent 提示词
// 一比一复刻 DeerFlow 的 apply_prompt_template
func BuildLeadAgentPrompt(config *LeadAgentConfig) *Prompt {
	if config == nil {
		config = &LeadAgentConfig{}
	}

	builder := NewBuilder()

	// 1. Role
	agentName := config.AgentName
	if agentName == "" {
		agentName = "DeerFlow 2.0"
	}
	builder.WithRole(agentName)

	// 2. Soul
	builder.WithSoul(config.SoulContent)

	// 3. Memory
	builder.WithMemory(config.MemoryContent)

	// 4. Thinking Style
	if config.SubagentEnabled {
		builder.WithThinkingStyleWithSubagent(config.MaxConcurrentSubagents)
	} else {
		builder.WithThinkingStyle()
	}

	// 5. Clarification
	builder.WithClarification()

	// 6. Skills
	builder.WithSkills(config.SkillsContent)

	// 7. Deferred Tools
	builder.WithDeferredTools(config.DeferredToolsContent)

	// 8. Subagent
	if config.SubagentEnabled {
		builder.WithSubagent(config.MaxConcurrentSubagents)
	}

	// 9. Working Dir
	builder.WithWorkingDir()

	// 10. Response Style
	builder.WithResponseStyle()

	// 11. Citations
	builder.WithCitations()

	// 12. Critical Reminders
	if config.SubagentEnabled {
		builder.WithCriticalRemindersWithSubagent(config.MaxConcurrentSubagents)
	} else {
		builder.WithCriticalReminders()
	}

	// 13. Current Date
	builder.WithCurrentDate()

	return builder.Build()
}

// BuildLeadAgentPromptString 构建 Lead Agent 提示词字符串
func BuildLeadAgentPromptString(config *LeadAgentConfig) string {
	return BuildLeadAgentPrompt(config).Render()
}

// BuildGeneralPurposeSubagentPrompt 构建通用子代理提示词
func BuildGeneralPurposeSubagentPrompt() *Prompt {
	builder := NewBuilder()
	builder.WithRole("General-Purpose Subagent")
	builder.WithThinkingStyle()
	builder.WithClarification()
	builder.WithWorkingDir()
	builder.WithResponseStyle()
	builder.WithCurrentDate()
	return builder.Build()
}

// BuildGeneralPurposeSubagentPromptString 构建通用子代理提示词字符串
func BuildGeneralPurposeSubagentPromptString() string {
	return BuildGeneralPurposeSubagentPrompt().Render()
}

// BuildBashSubagentPrompt 构建 Bash 子代理提示词
func BuildBashSubagentPrompt() *Prompt {
	builder := NewBuilder()
	builder.WithRole("Bash Specialist Subagent")
	builder.WithThinkingStyle()
	builder.WithClarification()
	builder.WithWorkingDir()
	builder.WithResponseStyle()
	builder.WithCurrentDate()

	// 添加 Bash 专家特定内容
	bashSpecialist := `You are a bash command execution specialist. Your primary tools are bash, ls, read_file, write_file, and str_replace.

**Focus on:**
- Executing shell commands efficiently
- File system operations
- Git operations
- Build and test commands
- Deployment operations

**Always explain what you're doing and why.**`
	builder.WithCustom("bash_specialist", "", bashSpecialist)

	return builder.Build()
}

// BuildBashSubagentPromptString 构建 Bash 子代理提示词字符串
func BuildBashSubagentPromptString() string {
	return BuildBashSubagentPrompt().Render()
}

// LeadAgentConfig Lead Agent 配置
type LeadAgentConfig struct {
	// AgentName 代理名称（默认：DeerFlow 2.0）
	AgentName string
	// SoulContent SOUL.md 内容
	SoulContent string
	// MemoryContent 记忆内容
	MemoryContent string
	// SkillsContent 技能内容
	SkillsContent string
	// DeferredToolsContent 延迟工具内容
	DeferredToolsContent string
	// SubagentEnabled 是否启用子代理
	SubagentEnabled bool
	// MaxConcurrentSubagents 最大并发子代理数（默认：3）
	MaxConcurrentSubagents int
}

// DefaultLeadAgentConfig 默认 Lead Agent 配置
func DefaultLeadAgentConfig() *LeadAgentConfig {
	return &LeadAgentConfig{
		AgentName:              "DeerFlow 2.0",
		SubagentEnabled:        false,
		MaxConcurrentSubagents: 3,
	}
}
