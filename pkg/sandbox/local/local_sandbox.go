package local

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/weibaohui/nanobot-go/pkg/sandbox"
)

// LocalSandbox 本地文件系统沙箱实现
// 一比一复刻 DeerFlow 的 LocalSandboxProvider
type LocalSandbox struct {
	id         string
	threadData *sandbox.ThreadData
	translator *sandbox.PathTranslator
}

// NewLocalSandbox 创建本地沙箱
func NewLocalSandbox(id string, threadData *sandbox.ThreadData) *LocalSandbox {
	return &LocalSandbox{
		id:         id,
		threadData: threadData,
		translator: sandbox.NewPathTranslator(threadData),
	}
}

// ID 实现 Sandbox 接口
func (s *LocalSandbox) ID() string {
	return s.id
}

// ExecuteCommand 实现 Sandbox 接口
func (s *LocalSandbox) ExecuteCommand(command string) (string, error) {
	// 翻译命令中的虚拟路径
	translatedCommand := sandbox.TranslatePathsInCommand(command, s.threadData, "")

	// 执行命令
	cmd := exec.Command("bash", "-c", translatedCommand)

	// 设置工作目录
	if s.threadData != nil && s.threadData.WorkspacePath != "" {
		cmd.Dir = s.threadData.WorkspacePath
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// 组合输出
	output := stdout.String()
	if stderr.String() != "" {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	// 掩码输出中的物理路径
	output = s.translator.MaskPathsInOutput(output)

	// 如果命令失败，仍然返回输出（但带错误）
	if err != nil {
		return output, fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}

// ReadFile 实现 Sandbox 接口
func (s *LocalSandbox) ReadFile(path string) (string, error) {
	// 翻译虚拟路径到物理路径
	physPath, err := s.translator.ToPhysical(path)
	if err != nil {
		return "", err
	}

	// 读取文件
	content, err := os.ReadFile(physPath)
	if err != nil {
		return "", fmt.Errorf("read file failed: %w", err)
	}

	return string(content), nil
}

// WriteFile 实现 Sandbox 接口
func (s *LocalSandbox) WriteFile(path, content string, append bool) error {
	// 翻译虚拟路径到物理路径
	physPath, err := s.translator.ToPhysical(path)
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(physPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}

	// 写入文件
	flags := os.O_WRONLY | os.O_CREATE
	if append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(physPath, flags, 0644)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

// ListDir 实现 Sandbox 接口
func (s *LocalSandbox) ListDir(path string, maxDepth int) ([]string, error) {
	// 翻译虚拟路径到物理路径
	physPath, err := s.translator.ToPhysical(path)
	if err != nil {
		return nil, err
	}

	// 验证路径存在
	info, err := os.Stat(physPath)
	if err != nil {
		return nil, fmt.Errorf("stat path failed: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}

	// 列出目录（树状格式）
	var result []string
	err = s.listDirRecursive(physPath, "", 0, maxDepth, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// listDirRecursive 递归列出目录
func (s *LocalSandbox) listDirRecursive(physPath, virtualPath string, depth, maxDepth int, result *[]string) error {
	if depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(physPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(physPath, entry.Name())
		entryVirtualPath := filepath.Join(virtualPath, entry.Name())

		// 转换为虚拟路径格式
		displayPath := s.translator.ToVirtual(entryPath)

		// 添加前缀表示层级
		prefix := strings.Repeat("  ", depth)
		symbol := "├── "
		if entry.IsDir() {
			symbol = "├── "
			displayPath += "/"
		}

		*result = append(*result, prefix+symbol+filepath.Base(displayPath))

		// 递归处理子目录
		if entry.IsDir() && depth < maxDepth {
			err = s.listDirRecursive(entryPath, entryVirtualPath, depth+1, maxDepth, result)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// UpdateFile 实现 Sandbox 接口
func (s *LocalSandbox) UpdateFile(path string, content []byte) error {
	// 翻译虚拟路径到物理路径
	physPath, err := s.translator.ToPhysical(path)
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(physPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(physPath, content, 0644); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

// ============================================
// LocalSandboxProvider 本地沙箱提供者
// ============================================

// LocalSandboxProvider 本地沙箱提供者（单例）
type LocalSandboxProvider struct {
	sandboxes map[string]*LocalSandbox
	mu        sync.RWMutex
	baseDir   string
}

var (
	localProviderInstance *LocalSandboxProvider
	localProviderOnce     sync.Once
)

// GetLocalSandboxProvider 获取本地沙箱提供者单例
func GetLocalSandboxProvider(baseDir string) *LocalSandboxProvider {
	localProviderOnce.Do(func() {
		if baseDir == "" {
			baseDir = ".deer-flow/threads"
		}
		localProviderInstance = &LocalSandboxProvider{
			sandboxes: make(map[string]*LocalSandbox),
			baseDir:   baseDir,
		}
	})
	return localProviderInstance
}

// Acquire 实现 SandboxProvider 接口
func (p *LocalSandboxProvider) Acquire(threadID string) (sandbox.Sandbox, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查是否已存在
	if sb, exists := p.sandboxes[threadID]; exists {
		return sb, nil
	}

	// 创建线程目录
	threadData, err := p.createThreadDirs(threadID)
	if err != nil {
		return nil, err
	}

	// 创建沙箱
	sb := NewLocalSandbox("local", threadData)
	p.sandboxes[threadID] = sb

	return sb, nil
}

// Get 实现 SandboxProvider 接口
func (p *LocalSandboxProvider) Get(threadID string) (sandbox.Sandbox, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sb, exists := p.sandboxes[threadID]
	return sb, exists
}

// Release 实现 SandboxProvider 接口
func (p *LocalSandboxProvider) Release(threadID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.sandboxes, threadID)
	return nil
}

// createThreadDirs 创建线程目录
func (p *LocalSandboxProvider) createThreadDirs(threadID string) (*sandbox.ThreadData, error) {
	threadDir := filepath.Join(p.baseDir, threadID, "user-data")

	workspacePath := filepath.Join(threadDir, "workspace")
	uploadsPath := filepath.Join(threadDir, "uploads")
	outputsPath := filepath.Join(threadDir, "outputs")

	// 创建目录
	for _, dir := range []string{workspacePath, uploadsPath, outputsPath} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s failed: %w", dir, err)
		}
	}

	return &sandbox.ThreadData{
		WorkspacePath: workspacePath,
		UploadsPath:   uploadsPath,
		OutputsPath:   outputsPath,
	}, nil
}

// Reset 重置提供者（用于测试）
func (p *LocalSandboxProvider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sandboxes = make(map[string]*LocalSandbox)
}
