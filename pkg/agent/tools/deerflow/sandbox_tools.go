package deerflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/sandbox"
)

// ============================================
// 工具上下文
// ============================================

// ToolRuntime 工具运行时
type ToolRuntime struct {
	State      *state.ThreadState
	Context    map[string]interface{}
	Config     map[string]interface{}
}

// ToolContext 工具执行上下文
type ToolContext struct {
	Runtime *ToolRuntime
	Sandbox sandbox.Sandbox
}

// ============================================
// Bash 工具
// ============================================

// BashTool bash 工具
// 一比一复刻 DeerFlow 的 bash_tool
type BashTool struct {
	*BaseDeerFlowTool
	sb sandbox.Sandbox
}

// NewBashTool 创建 bash 工具
func NewBashTool(sb sandbox.Sandbox) tool.BaseTool {
	return &BashTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"bash",
			"Execute a bash command in a Linux environment.\n\n- Use `python` to run Python code.\n- Prefer a thread-local virtual environment in `/mnt/user-data/workspace/.venv`.\n- Use `python -m pip` (inside the virtual environment) to install Python packages.",
			map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Explain why you are running this command in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.",
				},
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The bash command to execute. Always use absolute paths for files and directories.",
				},
			},
		),
		sb: sb,
	}
}

// Invoke 执行 bash 工具
func (t *BashTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	_, _ = args["description"].(string)
	command, _ := args["command"].(string)

	// 获取运行时上下文（需要从 ctx 中提取）
	// 注意：这里简化处理，实际使用时需要完整的状态管理
	var threadData *state.ThreadDataState
	var isLocalSandbox bool = true // 假设是本地沙箱

	// 验证和翻译路径
	if isLocalSandbox {
		if err := ValidateLocalBashCommandPaths(command, threadData); err != nil {
			return fmt.Sprintf("Error: %v", err), nil
		}
		command = ReplaceVirtualPathsInCommand(command, threadData)
	}

	// 执行命令
	if t.sb == nil {
		return "Error: Sandbox not initialized", nil
	}

	output, err := t.sb.ExecuteCommand(command)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	// 掩码输出路径
	if isLocalSandbox {
		output = MaskLocalPathsInOutput(output, threadData)
	}

	return output, nil
}

// ============================================
// Ls 工具
// ============================================

// LsTool ls 工具
// 一比一复刻 DeerFlow 的 ls_tool
type LsTool struct {
	*BaseDeerFlowTool
	sb sandbox.Sandbox
}

// NewLsTool 创建 ls 工具
func NewLsTool(sb sandbox.Sandbox) tool.BaseTool {
	return &LsTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"ls",
			"List the contents of a directory up to 2 levels deep in tree format.",
			map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Explain why you are listing this directory in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute path to the directory to list.",
				},
			},
		),
		sb: sb,
	}
}

// Invoke 执行 ls 工具
func (t *LsTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	_, _ = args["description"].(string)
	path, _ := args["path"].(string)

	// 验证路径
	var threadData *state.ThreadDataState
	var isLocalSandbox bool = true

	if isLocalSandbox {
		if err := ValidateLocalToolPath(path, threadData, true); err != nil {
			return fmt.Sprintf("Error: %v", err), nil
		}
		// 翻译路径
		if IsSkillsPath(path) {
			// TODO: Skills 路径解析
		} else {
			path = ReplaceVirtualPath(path, threadData)
		}
	}

	// 列出目录
	if t.sb == nil {
		return "Error: Sandbox not initialized", nil
	}

	children, err := t.sb.ListDir(path, 2)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	if len(children) == 0 {
		return "(empty)", nil
	}

	return strings.Join(children, "\n"), nil
}

// ============================================
// ReadFile 工具
// ============================================

// ReadFileTool read_file 工具
// 一比一复刻 DeerFlow 的 read_file_tool
type ReadFileTool struct {
	*BaseDeerFlowTool
	sb sandbox.Sandbox
}

// NewReadFileTool 创建 read_file 工具
func NewReadFileTool(sb sandbox.Sandbox) tool.BaseTool {
	return &ReadFileTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"read_file",
			"Read the contents of a text file. Use this to examine source code, configuration files, logs, or any text-based file.",
			map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Explain why you are reading this file in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute path to the file to read.",
				},
				"start_line": map[string]interface{}{
					"type":        "integer",
					"description": "Optional starting line number (1-indexed, inclusive). Use with end_line to read a specific range.",
				},
				"end_line": map[string]interface{}{
					"type":        "integer",
					"description": "Optional ending line number (1-indexed, inclusive). Use with start_line to read a specific range.",
				},
			},
		),
		sb: sb,
	}
}

// Invoke 执行 read_file 工具
func (t *ReadFileTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	_, _ = args["description"].(string)
	path, _ := args["path"].(string)

	// 解析可选参数
	var startLine, endLine *int
	if sl, ok := args["start_line"].(int); ok {
		startLine = &sl
	}
	if el, ok := args["end_line"].(int); ok {
		endLine = &el
	}

	// 验证路径
	var threadData *state.ThreadDataState
	var isLocalSandbox bool = true

	if isLocalSandbox {
		if err := ValidateLocalToolPath(path, threadData, true); err != nil {
			return fmt.Sprintf("Error: %v", err), nil
		}
		// 翻译路径
		if IsSkillsPath(path) {
			// TODO: Skills 路径解析
		} else {
			path = ReplaceVirtualPath(path, threadData)
		}
	}

	// 读取文件
	if t.sb == nil {
		return "Error: Sandbox not initialized", nil
	}

	content, err := t.sb.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	if content == "" {
		return "(empty)", nil
	}

	// 处理行范围
	if startLine != nil && endLine != nil {
		lines := strings.Split(content, "\n")
		start := *startLine - 1
		end := *endLine
		if start < 0 {
			start = 0
		}
		if end > len(lines) {
			end = len(lines)
		}
		if start >= end {
			return "(empty)", nil
		}
		content = strings.Join(lines[start:end], "\n")
	}

	return content, nil
}

// ============================================
// WriteFile 工具
// ============================================

// WriteFileTool write_file 工具
// 一比一复刻 DeerFlow 的 write_file_tool
type WriteFileTool struct {
	*BaseDeerFlowTool
	sb sandbox.Sandbox
}

// NewWriteFileTool 创建 write_file 工具
func NewWriteFileTool(sb sandbox.Sandbox) tool.BaseTool {
	return &WriteFileTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"write_file",
			"Write text content to a file.",
			map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Explain why you are writing to this file in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute path to the file to write to. ALWAYS PROVIDE THIS PARAMETER SECOND.",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The content to write to the file. ALWAYS PROVIDE THIS PARAMETER THIRD.",
				},
				"append": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to append to the file (optional, default false).",
				},
			},
		),
		sb: sb,
	}
}

// Invoke 执行 write_file 工具
func (t *WriteFileTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	_, _ = args["description"].(string)
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	append, _ := args["append"].(bool)

	// 验证路径
	var threadData *state.ThreadDataState
	var isLocalSandbox bool = true

	if isLocalSandbox {
		if err := ValidateLocalToolPath(path, threadData, false); err != nil {
			return fmt.Sprintf("Error: %v", err), nil
		}
		path = ReplaceVirtualPath(path, threadData)
	}

	// 写入文件
	if t.sb == nil {
		return "Error: Sandbox not initialized", nil
	}

	err := t.sb.WriteFile(path, content, append)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	return "OK", nil
}

// ============================================
// StrReplace 工具
// ============================================

// StrReplaceTool str_replace 工具
// 一比一复刻 DeerFlow 的 str_replace_tool
type StrReplaceTool struct {
	*BaseDeerFlowTool
	sb sandbox.Sandbox
}

// NewStrReplaceTool 创建 str_replace 工具
func NewStrReplaceTool(sb sandbox.Sandbox) tool.BaseTool {
	return &StrReplaceTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"str_replace",
			"Replace a substring in a file with another substring.\nIf `replace_all` is False (default), the substring to replace must appear exactly once in the file.",
			map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Explain why you are replacing the substring in short words. ALWAYS PROVIDE THIS PARAMETER FIRST.",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute path to the file to replace the substring in. ALWAYS PROVIDE THIS PARAMETER SECOND.",
				},
				"old_str": map[string]interface{}{
					"type":        "string",
					"description": "The substring to replace. ALWAYS PROVIDE THIS PARAMETER THIRD.",
				},
				"new_str": map[string]interface{}{
					"type":        "string",
					"description": "The new substring to use. ALWAYS PROVIDE THIS PARAMETER FOURTH.",
				},
				"replace_all": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to replace all occurrences of the substring. If False, only the first occurrence will be replaced. Default is False.",
				},
			},
		),
		sb: sb,
	}
}

// Invoke 执行 str_replace 工具
func (t *StrReplaceTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	_, _ = args["description"].(string)
	path, _ := args["path"].(string)
	oldStr, _ := args["old_str"].(string)
	newStr, _ := args["new_str"].(string)
	replaceAll, _ := args["replace_all"].(bool)

	// 验证路径
	var threadData *state.ThreadDataState
	var isLocalSandbox bool = true

	if isLocalSandbox {
		if err := ValidateLocalToolPath(path, threadData, false); err != nil {
			return fmt.Sprintf("Error: %v", err), nil
		}
		path = ReplaceVirtualPath(path, threadData)
	}

	// 读取文件
	if t.sb == nil {
		return "Error: Sandbox not initialized", nil
	}

	content, err := t.sb.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	if content == "" {
		return "OK", nil
	}

	// 检查 old_str 是否存在
	if !strings.Contains(content, oldStr) {
		return fmt.Sprintf("Error: String to replace not found in file: %s", path), nil
	}

	// 执行替换
	if replaceAll {
		content = strings.ReplaceAll(content, oldStr, newStr)
	} else {
		content = strings.Replace(content, oldStr, newStr, 1)
	}

	// 写回文件
	err = t.sb.WriteFile(path, content, false)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	return "OK", nil
}

// ============================================
// 基础工具版本（无沙箱）
// ============================================

// NewBasicBashTool 创建基础 bash 工具
func NewBasicBashTool() tool.BaseTool {
	return NewBashTool(nil)
}

// NewBasicLsTool 创建基础 ls 工具
func NewBasicLsTool() tool.BaseTool {
	return NewLsTool(nil)
}

// NewBasicReadFileTool 创建基础 read_file 工具
func NewBasicReadFileTool() tool.BaseTool {
	return NewReadFileTool(nil)
}

// NewBasicWriteFileTool 创建基础 write_file 工具
func NewBasicWriteFileTool() tool.BaseTool {
	return NewWriteFileTool(nil)
}

// NewBasicStrReplaceTool 创建基础 str_replace 工具
func NewBasicStrReplaceTool() tool.BaseTool {
	return NewStrReplaceTool(nil)
}
