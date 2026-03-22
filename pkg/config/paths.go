package config

import (
	"os"
	"path/filepath"
	"sync"
)

// 虚拟路径常量（一比一复刻 DeerFlow）
const (
	VirtualPathPrefix     = "/mnt/user-data"
	VirtualWorkspacePath  = VirtualPathPrefix + "/workspace"
	VirtualUploadsPath    = VirtualPathPrefix + "/uploads"
	VirtualOutputsPath    = VirtualPathPrefix + "/outputs"
	VirtualSkillsPath     = "/mnt/skills"
)

// Paths 路径管理器
// 一比一复刻 DeerFlow 的 Paths
type Paths struct {
	baseDir string
}

var (
	globalPaths *Paths
	pathsOnce   sync.Once
)

// NewPaths 创建路径管理器
func NewPaths(baseDir string) *Paths {
	if baseDir == "" {
		baseDir = getDefaultBaseDir()
	}
	return &Paths{
		baseDir: baseDir,
	}
}

// GetPaths 获取全局路径管理器（单例）
func GetPaths() *Paths {
	pathsOnce.Do(func() {
		globalPaths = NewPaths("")
	})
	return globalPaths
}

// getDefaultBaseDir 获取默认基础目录
func getDefaultBaseDir() string {
	// 优先使用环境变量
	if envDir := os.Getenv("DEER_FLOW_BASE_DIR"); envDir != "" {
		return envDir
	}
	// 默认：当前工作目录下的 .deer-flow
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	return filepath.Join(wd, ".deer-flow")
}

// BaseDir 获取基础目录
func (p *Paths) BaseDir() string {
	return p.baseDir
}

// ThreadsDir 获取线程目录
func (p *Paths) ThreadsDir() string {
	return filepath.Join(p.baseDir, "threads")
}

// ThreadDir 获取指定线程的目录
func (p *Paths) ThreadDir(threadID string) string {
	return filepath.Join(p.ThreadsDir(), threadID)
}

// SandboxWorkDir 获取沙箱工作区目录
func (p *Paths) SandboxWorkDir(threadID string) string {
	return filepath.Join(p.ThreadDir(threadID), "user-data", "workspace")
}

// SandboxUploadsDir 获取沙箱上传目录
func (p *Paths) SandboxUploadsDir(threadID string) string {
	return filepath.Join(p.ThreadDir(threadID), "user-data", "uploads")
}

// SandboxOutputsDir 获取沙箱输出目录
func (p *Paths) SandboxOutputsDir(threadID string) string {
	return filepath.Join(p.ThreadDir(threadID), "user-data", "outputs")
}

// EnsureThreadDirs 确保线程目录存在
func (p *Paths) EnsureThreadDirs(threadID string) error {
	dirs := []string{
		p.SandboxWorkDir(threadID),
		p.SandboxUploadsDir(threadID),
		p.SandboxOutputsDir(threadID),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// MemoryFile 获取记忆文件路径
func (p *Paths) MemoryFile() string {
	return filepath.Join(p.baseDir, "memory.json")
}

// AgentMemoryFile 获取代理专属记忆文件路径
func (p *Paths) AgentMemoryFile(agentName string) string {
	return filepath.Join(p.baseDir, "agent_memory_"+agentName+".json")
}

// ResolveVirtualPath 解析虚拟路径到物理路径
func (p *Paths) ResolveVirtualPath(threadID, virtualPath string) string {
	workspaceDir := p.SandboxWorkDir(threadID)
	uploadsDir := p.SandboxUploadsDir(threadID)
	outputsDir := p.SandboxOutputsDir(threadID)

	// 最长前缀匹配
	switch {
	case virtualPath == VirtualWorkspacePath:
		return workspaceDir
	case virtualPath == VirtualUploadsPath:
		return uploadsDir
	case virtualPath == VirtualOutputsPath:
		return outputsDir
	case len(virtualPath) > len(VirtualWorkspacePath) && virtualPath[:len(VirtualWorkspacePath)+1] == VirtualWorkspacePath+"/":
		return filepath.Join(workspaceDir, virtualPath[len(VirtualWorkspacePath)+1:])
	case len(virtualPath) > len(VirtualUploadsPath) && virtualPath[:len(VirtualUploadsPath)+1] == VirtualUploadsPath+"/":
		return filepath.Join(uploadsDir, virtualPath[len(VirtualUploadsPath)+1:])
	case len(virtualPath) > len(VirtualOutputsPath) && virtualPath[:len(VirtualOutputsPath)+1] == VirtualOutputsPath+"/":
		return filepath.Join(outputsDir, virtualPath[len(VirtualOutputsPath)+1:])
	default:
		return virtualPath
	}
}
