package sandbox

import (
	"path/filepath"
	"regexp"
	"strings"
)

// PathTranslator 路径翻译器
// 一比一复刻 DeerFlow 的虚拟路径系统
type PathTranslator struct {
	threadData *ThreadData
}

// NewPathTranslator 创建路径翻译器
func NewPathTranslator(threadData *ThreadData) *PathTranslator {
	return &PathTranslator{
		threadData: threadData,
	}
}

// ToPhysical 将虚拟路径转换为物理路径
func (t *PathTranslator) ToPhysical(virtualPath string) (string, error) {
	if t.threadData == nil {
		return "", ErrSandboxNotAcquired
	}

	virtualPath = filepath.Clean(virtualPath)

	// 防止路径遍历
	if err := t.ValidatePath(virtualPath, false); err != nil {
		return "", err
	}

	var physicalPath string

	switch {
	case strings.HasPrefix(virtualPath, VirtualWorkspacePath):
		relPath := strings.TrimPrefix(virtualPath, VirtualWorkspacePath)
		physicalPath = filepath.Join(t.threadData.WorkspacePath, relPath)
	case strings.HasPrefix(virtualPath, VirtualUploadsPath):
		relPath := strings.TrimPrefix(virtualPath, VirtualUploadsPath)
		physicalPath = filepath.Join(t.threadData.UploadsPath, relPath)
	case strings.HasPrefix(virtualPath, VirtualOutputsPath):
		relPath := strings.TrimPrefix(virtualPath, VirtualOutputsPath)
		physicalPath = filepath.Join(t.threadData.OutputsPath, relPath)
	default:
		return "", ErrPathNotInSandbox
	}

	return filepath.Clean(physicalPath), nil
}

// ToVirtual 将物理路径转换为虚拟路径
func (t *PathTranslator) ToVirtual(physicalPath string) string {
	if t.threadData == nil {
		return physicalPath
	}

	result := physicalPath

	// 替换工作区路径
	if t.threadData.WorkspacePath != "" && strings.Contains(result, t.threadData.WorkspacePath) {
		result = strings.ReplaceAll(result, t.threadData.WorkspacePath, VirtualWorkspacePath)
	}

	// 替换上传路径
	if t.threadData.UploadsPath != "" && strings.Contains(result, t.threadData.UploadsPath) {
		result = strings.ReplaceAll(result, t.threadData.UploadsPath, VirtualUploadsPath)
	}

	// 替换输出路径
	if t.threadData.OutputsPath != "" && strings.Contains(result, t.threadData.OutputsPath) {
		result = strings.ReplaceAll(result, t.threadData.OutputsPath, VirtualOutputsPath)
	}

	// 统一使用正斜杠
	result = filepath.ToSlash(result)

	return result
}

// ValidatePath 验证路径是否安全
func (t *PathTranslator) ValidatePath(virtualPath string, allowOutside bool) error {
	virtualPath = filepath.Clean(virtualPath)

	// 检查是否包含路径遍历
	if strings.Contains(virtualPath, "..") {
		return ErrPathTraversal
	}

	// 如果不允许在沙箱外，检查路径前缀
	if !allowOutside {
		hasValidPrefix := strings.HasPrefix(virtualPath, VirtualWorkspacePath) ||
			strings.HasPrefix(virtualPath, VirtualUploadsPath) ||
			strings.HasPrefix(virtualPath, VirtualOutputsPath) ||
			strings.HasPrefix(virtualPath, VirtualSkillsPath)

		if !hasValidPrefix {
			return ErrPathNotInSandbox
		}
	}

	return nil
}

// MaskPathsInOutput 在输出中掩码物理路径
func (t *PathTranslator) MaskPathsInOutput(output string) string {
	return t.ToVirtual(output)
}

// IsVirtualPath 检查是否是虚拟路径
func IsVirtualPath(path string) bool {
	path = filepath.Clean(path)
	return strings.HasPrefix(path, VirtualPathPrefix) ||
		strings.HasPrefix(path, VirtualSkillsPath)
}

// 路径匹配模式
var (
	// 匹配可能的路径格式
	pathPatterns = []*regexp.Regexp{
		// 匹配绝对路径: /path/to/something
		regexp.MustCompile(`(?m)(/[\w\-./]+)`),
		// 匹配 Windows 路径: C:\path\to\something
		regexp.MustCompile(`(?m)([A-Za-z]:\\[\w\-\\.]+)`),
	}
)

// TranslatePathsInCommand 翻译命令中的虚拟路径
func TranslatePathsInCommand(command string, threadData *ThreadData, skillsPath string) string {
	if threadData == nil {
		return command
	}

	result := command

	// 替换工作区路径
	if threadData.WorkspacePath != "" {
		result = strings.ReplaceAll(result, VirtualWorkspacePath, threadData.WorkspacePath)
	}

	// 替换上传路径
	if threadData.UploadsPath != "" {
		result = strings.ReplaceAll(result, VirtualUploadsPath, threadData.UploadsPath)
	}

	// 替换输出路径
	if threadData.OutputsPath != "" {
		result = strings.ReplaceAll(result, VirtualOutputsPath, threadData.OutputsPath)
	}

	// 替换技能路径
	if skillsPath != "" {
		result = strings.ReplaceAll(result, VirtualSkillsPath, skillsPath)
	}

	return result
}

// ExtractVirtualPaths 从文本中提取虚拟路径
func ExtractVirtualPaths(text string) []string {
	var paths []string
	seen := make(map[string]bool)

	for _, pattern := range pathPatterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			cleanPath := filepath.Clean(match)
			if IsVirtualPath(cleanPath) && !seen[cleanPath] {
				seen[cleanPath] = true
				paths = append(paths, cleanPath)
			}
		}
	}

	return paths
}

// EnsureTrailingSlash 确保路径以斜杠结尾
func EnsureTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") && !strings.HasSuffix(path, "\\") {
		return path + "/"
	}
	return path
}
