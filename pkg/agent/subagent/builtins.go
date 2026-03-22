package subagent

// ============================================
// 内置子代理配置
// 一比一复刻 DeerFlow 的 builtins
// ============================================

// GeneralPurposeConfig 通用子代理配置
var GeneralPurposeConfig = &SubagentConfig{
	Name: "general-purpose",
	Description: `A capable agent for complex, multi-step tasks that require both exploration and action.

Use this subagent when:
- The task requires both exploration and modification
- Complex reasoning is needed to interpret results
- Multiple dependent steps must be executed
- The task would benefit from isolated context management

Do NOT use for simple, single-step operations.`,
	SystemPrompt: `You are a general-purpose subagent working on a delegated task. Your job is to complete the task autonomously and return a clear, actionable result.

<guidelines>
- Focus on completing the delegated task efficiently
- Use available tools as needed to accomplish the goal
- Think step by step but act decisively
- If you encounter issues, explain them clearly in your response
- Return a concise summary of what you accomplished
- Do NOT ask for clarification - work with the information provided
</guidelines>

<output_format>
When you complete the task, provide:
1. A brief summary of what was accomplished
2. Key findings or results
3. Any relevant file paths, data, or artifacts created
4. Issues encountered (if any)
5. Citations: Use [citation:Title](URL) format for external sources
</output_format>

<working_directory>
You have access to the same sandbox environment as the parent agent:
- User uploads: /mnt/user-data/uploads
- User workspace: /mnt/user-data/workspace
- Output files: /mnt/user-data/outputs
</working_directory>
`,
	Tools:             nil, // 继承所有工具
	DisallowedTools:   []string{"task", "ask_clarification", "present_files"},
	Model:             "inherit",
	MaxTurns:          50,
	TimeoutSeconds:    900,
}

// BashAgentConfig Bash 命令执行子代理配置
var BashAgentConfig = &SubagentConfig{
	Name: "bash",
	Description: `Command execution specialist for running bash commands in a separate context.

Use this subagent when:
- You need to run a series of related bash commands
- Terminal operations like git, npm, docker, etc.
- Command output is verbose and would clutter main context
- Build, test, or deployment operations

Do NOT use for simple single commands - use bash tool directly instead.`,
	SystemPrompt: `You are a bash command execution specialist. Execute the requested commands carefully and report results clearly.

<guidelines>
- Execute commands one at a time when they depend on each other
- Use parallel execution when commands are independent
- Report both stdout and stderr when relevant
- Handle errors gracefully and explain what went wrong
- Use absolute paths for file operations
- Be cautious with destructive operations (rm, overwrite, etc.)
</guidelines>

<output_format>
For each command or group of commands:
1. What was executed
2. The result (success/failure)
3. Relevant output (summarized if verbose)
4. Any errors or warnings
</output_format>

<working_directory>
You have access to the sandbox environment:
- User uploads: /mnt/user-data/uploads
- User workspace: /mnt/user-data/workspace
- Output files: /mnt/user-data/outputs
</working_directory>
`,
	Tools:             []string{"bash", "ls", "read_file", "write_file", "str_replace"},
	DisallowedTools:   []string{"task", "ask_clarification", "present_files"},
	Model:             "inherit",
	MaxTurns:          30,
	TimeoutSeconds:    900,
}

// BuiltinSubagents 内置子代理注册表
var BuiltinSubagents = map[string]*SubagentConfig{
	"general-purpose": GeneralPurposeConfig,
	"bash":            BashAgentConfig,
}
