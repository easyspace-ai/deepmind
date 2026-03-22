package sandbox

import (
	"errors"
)

// 常用错误定义
var (
	ErrInvalidPath       = errors.New("invalid path")
	ErrPathTraversal     = errors.New("path traversal detected")
	ErrPathNotInSandbox  = errors.New("path not in sandbox")
	ErrSandboxNotAcquired = errors.New("sandbox not acquired")
	ErrSandboxRuntime    = errors.New("sandbox runtime error")
	ErrPermissionDenied  = errors.New("permission denied")
)

// SandboxState 沙箱状态
type SandboxState struct {
	SandboxID string `json:"sandbox_id,omitempty"`
}

// ThreadData 线程数据（路径映射）
type ThreadData struct {
	WorkspacePath string `json:"workspace_path,omitempty"`
	UploadsPath   string `json:"uploads_path,omitempty"`
	OutputsPath   string `json:"outputs_path,omitempty"`
}

// Sandbox 沙箱接口
// 一比一复刻 DeerFlow 的 Sandbox 抽象
type Sandbox interface {
	// ID 返回沙箱 ID
	ID() string

	// ExecuteCommand 执行 bash 命令
	ExecuteCommand(command string) (string, error)

	// ReadFile 读取文件内容
	ReadFile(path string) (string, error)

	// WriteFile 写入文件内容
	WriteFile(path, content string, append bool) error

	// ListDir 列出目录内容（树状格式）
	ListDir(path string, maxDepth int) ([]string, error)

	// UpdateFile 更新二进制文件内容
	UpdateFile(path string, content []byte) error
}

// SandboxProvider 沙箱提供者接口
type SandboxProvider interface {
	// Acquire 获取沙箱
	Acquire(threadID string) (Sandbox, error)

	// Get 获取已存在的沙箱
	Get(threadID string) (Sandbox, bool)

	// Release 释放沙箱
	Release(threadID string) error
}

// 虚拟路径常量
const (
	// VirtualPathPrefix 虚拟路径前缀
	VirtualPathPrefix = "/mnt/user-data"
	// VirtualWorkspacePath 虚拟工作区路径
	VirtualWorkspacePath = "/mnt/user-data/workspace"
	// VirtualUploadsPath 虚拟上传路径
	VirtualUploadsPath = "/mnt/user-data/uploads"
	// VirtualOutputsPath 虚拟输出路径
	VirtualOutputsPath = "/mnt/user-data/outputs"
	// VirtualSkillsPath 虚拟技能路径
	VirtualSkillsPath = "/mnt/skills"
)
