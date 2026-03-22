package deerflow

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/config"
	"github.com/weibaohui/nanobot-go/pkg/sandbox"
)

// 本地 bash 系统路径前缀白名单
var localBashSystemPathPrefixes = []string{
	"/bin/",
	"/usr/bin/",
	"/usr/sbin/",
	"/sbin/",
	"/opt/homebrew/bin/",
	"/dev/",
}

// 绝对路径正则表达式（RE2 无 lookbehind，用宽松匹配；排除空白与 shell 元字符）
var absolutePathPattern = regexp.MustCompile("/(?:[^\\s\"'" + "`" + ";\\&|<>()]+)")

// ============================================
// 路径验证和翻译
// ============================================

// ValidateLocalToolPath 验证本地工具路径
// 一比一复刻 DeerFlow 的 validate_local_tool_path
func ValidateLocalToolPath(path string, threadData *state.ThreadDataState, readOnly bool) error {
	if threadData == nil {
		return sandbox.ErrSandboxRuntime
	}

	// 拒绝路径遍历
	if err := RejectPathTraversal(path); err != nil {
		return err
	}

	// Skills 路径：仅只读访问
	if IsSkillsPath(path) {
		if !readOnly {
			return sandbox.ErrPermissionDenied
		}
		return nil
	}

	// 用户数据路径
	if strings.HasPrefix(path, config.VirtualPathPrefix+"/") {
		return nil
	}

	return sandbox.ErrPermissionDenied
}

// RejectPathTraversal 拒绝路径遍历
// 一比一复刻 DeerFlow 的 _reject_path_traversal
func RejectPathTraversal(path string) error {
	// 标准化为正斜杠，然后检查 '..' 段
	normalized := strings.ReplaceAll(path, "\\", "/")
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return sandbox.ErrPermissionDenied
		}
	}
	return nil
}

// IsSkillsPath 检查是否是技能路径
func IsSkillsPath(path string) bool {
	skillsPrefix := config.VirtualSkillsPath
	return path == skillsPrefix || strings.HasPrefix(path, skillsPrefix+"/")
}

// IsVirtualPath 检查是否是虚拟路径
func IsVirtualPath(path string) bool {
	return strings.HasPrefix(path, config.VirtualPathPrefix+"/") || IsSkillsPath(path)
}

// ReplaceVirtualPath 替换虚拟路径为实际路径
// 一比一复刻 DeerFlow 的 replace_virtual_path
func ReplaceVirtualPath(path string, threadData *state.ThreadDataState) string {
	if threadData == nil {
		return path
	}

	mappings := getThreadVirtualToActualMappings(threadData)
	if len(mappings) == 0 {
		return path
	}

	// 最长前缀优先替换
	// 按长度降序排序
	type mapping struct {
		virtual string
		actual  string
	}
	var sortedMappings []mapping
	for v, a := range mappings {
		sortedMappings = append(sortedMappings, mapping{v, a})
	}
	// 冒泡排序按长度降序
	for i := 0; i < len(sortedMappings); i++ {
		for j := i + 1; j < len(sortedMappings); j++ {
			if len(sortedMappings[i].virtual) < len(sortedMappings[j].virtual) {
				sortedMappings[i], sortedMappings[j] = sortedMappings[j], sortedMappings[i]
			}
		}
	}

	for _, m := range sortedMappings {
		if path == m.virtual {
			return m.actual
		}
		if strings.HasPrefix(path, m.virtual+"/") {
			rest := strings.TrimPrefix(path[len(m.virtual):], "/")
			if rest == "" {
				return m.actual
			}
			return filepath.Join(m.actual, rest)
		}
	}

	return path
}

// getThreadVirtualToActualMappings 获取线程虚拟到实际路径映射
func getThreadVirtualToActualMappings(threadData *state.ThreadDataState) map[string]string {
	mappings := make(map[string]string)

	if threadData.WorkspacePath != "" {
		mappings[config.VirtualWorkspacePath] = threadData.WorkspacePath
	}
	if threadData.UploadsPath != "" {
		mappings[config.VirtualUploadsPath] = threadData.UploadsPath
	}
	if threadData.OutputsPath != "" {
		mappings[config.VirtualOutputsPath] = threadData.OutputsPath
	}

	// 当所有已知目录共享同一个父目录时，也映射虚拟根
	var actualDirs []string
	if threadData.WorkspacePath != "" {
		actualDirs = append(actualDirs, threadData.WorkspacePath)
	}
	if threadData.UploadsPath != "" {
		actualDirs = append(actualDirs, threadData.UploadsPath)
	}
	if threadData.OutputsPath != "" {
		actualDirs = append(actualDirs, threadData.OutputsPath)
	}

	if len(actualDirs) > 0 {
		commonParent := filepath.Dir(actualDirs[0])
		allSameParent := true
		for _, dir := range actualDirs {
			if filepath.Dir(dir) != commonParent {
				allSameParent = false
				break
			}
		}
		if allSameParent {
			mappings[config.VirtualPathPrefix] = commonParent
		}
	}

	return mappings
}

// getThreadActualToVirtualMappings 获取线程实际到虚拟路径映射
func getThreadActualToVirtualMappings(threadData *state.ThreadDataState) map[string]string {
	virtualToActual := getThreadVirtualToActualMappings(threadData)
	actualToVirtual := make(map[string]string)
	for v, a := range virtualToActual {
		actualToVirtual[a] = v
	}
	return actualToVirtual
}

// ============================================
// 命令路径验证
// ============================================

// ValidateLocalBashCommandPaths 验证本地 bash 命令中的路径
// 一比一复刻 DeerFlow 的 validate_local_bash_command_paths
func ValidateLocalBashCommandPaths(command string, threadData *state.ThreadDataState) error {
	if threadData == nil {
		return sandbox.ErrSandboxRuntime
	}

	var unsafePaths []string

	for _, absolutePath := range absolutePathPattern.FindAllString(command, -1) {
		if absolutePath == config.VirtualPathPrefix || strings.HasPrefix(absolutePath, config.VirtualPathPrefix+"/") {
			if err := RejectPathTraversal(absolutePath); err != nil {
				return err
			}
			continue
		}

		// 允许技能容器路径
		if IsSkillsPath(absolutePath) {
			if err := RejectPathTraversal(absolutePath); err != nil {
				return err
			}
			continue
		}

		// 检查系统路径白名单
		isAllowed := false
		for _, prefix := range localBashSystemPathPrefixes {
			if absolutePath == strings.TrimSuffix(prefix, "/") || strings.HasPrefix(absolutePath, prefix) {
				isAllowed = true
				break
			}
		}
		if isAllowed {
			continue
		}

		unsafePaths = append(unsafePaths, absolutePath)
	}

	if len(unsafePaths) > 0 {
		// 去重并排序
		seen := make(map[string]bool)
		var unique []string
		for _, p := range unsafePaths {
			if !seen[p] {
				seen[p] = true
				unique = append(unique, p)
			}
		}
		return sandbox.ErrPermissionDenied
	}

	return nil
}

// ============================================
// 路径掩码（输出隐藏物理路径）
// ============================================

// MaskLocalPathsInOutput 掩码输出中的本地路径
// 一比一复刻 DeerFlow 的 mask_local_paths_in_output
func MaskLocalPathsInOutput(output string, threadData *state.ThreadDataState) string {
	result := output

	// TODO: Skills 路径掩码（低优先级）

	// 用户数据路径掩码
	if threadData == nil {
		return result
	}

	mappings := getThreadActualToVirtualMappings(threadData)
	if len(mappings) == 0 {
		return result
	}

	// 最长前缀优先替换
	type mapping struct {
		actual  string
		virtual string
	}
	var sortedMappings []mapping
	for a, v := range mappings {
		sortedMappings = append(sortedMappings, mapping{a, v})
	}
	// 冒泡排序按长度降序
	for i := 0; i < len(sortedMappings); i++ {
		for j := i + 1; j < len(sortedMappings); j++ {
			if len(sortedMappings[i].actual) < len(sortedMappings[j].actual) {
				sortedMappings[i], sortedMappings[j] = sortedMappings[j], sortedMappings[i]
			}
		}
	}

	for _, m := range sortedMappings {
		// 替换实际路径为虚拟路径
		// 处理正斜杠和反斜杠变体
		for _, base := range pathVariants(m.actual) {
			// 使用正则替换
			escaped := regexp.QuoteMeta(base)
			// 替换反斜杠为 [/\\] 以匹配两种分隔符
			escaped = strings.ReplaceAll(escaped, "\\", "[/\\\\]")
			pathTail := "(?:[/\\\\][^\\s\"'" + "`" + ";\\&|<>()]*)?"
			pattern := regexp.MustCompile(escaped + pathTail)
			result = pattern.ReplaceAllStringFunc(result, func(matched string) string {
				if matched == base {
					return m.virtual
				}
				relative := strings.TrimPrefix(matched[len(base):], "/")
				relative = strings.TrimPrefix(relative, "\\")
				if relative == "" {
					return m.virtual
				}
				return m.virtual + "/" + relative
			})
		}
	}

	return result
}

// pathVariants 获取路径变体（正斜杠和反斜杠）
func pathVariants(path string) []string {
	return []string{
		path,
		strings.ReplaceAll(path, "\\", "/"),
		strings.ReplaceAll(path, "/", "\\"),
	}
}

// ============================================
// 命令路径翻译
// ============================================

// ReplaceVirtualPathsInCommand 替换命令中的虚拟路径
// 一比一复刻 DeerFlow 的 replace_virtual_paths_in_command
func ReplaceVirtualPathsInCommand(command string, threadData *state.ThreadDataState) string {
	result := command

	// TODO: Skills 路径替换（低优先级）

	// 用户数据路径替换
	if threadData != nil && strings.Contains(command, config.VirtualPathPrefix) {
		mappings := getThreadVirtualToActualMappings(threadData)
		if len(mappings) > 0 {
			// 最长前缀优先替换
			type mapping struct {
				virtual string
				actual  string
			}
			var sortedMappings []mapping
			for v, a := range mappings {
				sortedMappings = append(sortedMappings, mapping{v, a})
			}
			// 冒泡排序按长度降序
			for i := 0; i < len(sortedMappings); i++ {
				for j := i + 1; j < len(sortedMappings); j++ {
					if len(sortedMappings[i].virtual) < len(sortedMappings[j].virtual) {
						sortedMappings[i], sortedMappings[j] = sortedMappings[j], sortedMappings[i]
					}
				}
			}

			for _, m := range sortedMappings {
				virtTail := "(/[^\\s\"'" + "`" + ";\\&|<>()]*)?"
				pattern := regexp.MustCompile(regexp.QuoteMeta(m.virtual) + virtTail)
				result = pattern.ReplaceAllStringFunc(result, func(matched string) string {
					return ReplaceVirtualPath(matched, threadData)
				})
			}
		}
	}

	return result
}
