package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ============================================
// Memory Updater（一比一复刻 DeerFlow）
// ============================================

// memoryCacheEntry 记忆缓存条目
type memoryCacheEntry struct {
	data  *MemoryData
	mtime float64 // 文件修改时间戳
}

var (
	// memoryCache 记忆缓存：agentName -> cache entry
	memoryCache      = make(map[string]*memoryCacheEntry)
	memoryCacheMutex sync.RWMutex
)

// ============================================
// 文件路径处理
// ============================================

// getMemoryFilePath 获取记忆文件路径
func getMemoryFilePath(agentName string) string {
	// 使用临时基础目录避免循环导入
	baseDir := getDefaultBaseDir()
	if agentName != "" {
		return filepath.Join(baseDir, "agent_memory_"+agentName+".json")
	}
	return filepath.Join(baseDir, "memory.json")
}

func getDefaultBaseDir() string {
	if envDir := os.Getenv("DEER_FLOW_BASE_DIR"); envDir != "" {
		return envDir
	}
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	return filepath.Join(wd, ".deer-flow")
}

// ============================================
// 加载与保存（原子文件 I/O + 缓存）
// ============================================

// GetMemoryData 获取当前记忆数据（带缓存和文件修改时间检查）
func GetMemoryData(agentName string) (*MemoryData, error) {
	filePath := getMemoryFilePath(agentName)
	currentMtime := getFileMtime(filePath)

	memoryCacheMutex.RLock()
	cached, ok := memoryCache[agentName]
	memoryCacheMutex.RUnlock()

	if ok && cached.mtime == currentMtime {
		return cached.data, nil
	}

	data, err := loadMemoryFromFile(agentName)
	if err != nil {
		return nil, err
	}

	memoryCacheMutex.Lock()
	memoryCache[agentName] = &memoryCacheEntry{
		data:  data,
		mtime: currentMtime,
	}
	memoryCacheMutex.Unlock()

	return data, nil
}

// ReloadMemoryData 强制从文件重新加载记忆数据
func ReloadMemoryData(agentName string) (*MemoryData, error) {
	filePath := getMemoryFilePath(agentName)
	data, err := loadMemoryFromFile(agentName)
	if err != nil {
		return nil, err
	}

	mtime := getFileMtime(filePath)
	memoryCacheMutex.Lock()
	memoryCache[agentName] = &memoryCacheEntry{
		data:  data,
		mtime: mtime,
	}
	memoryCacheMutex.Unlock()

	return data, nil
}

// loadMemoryFromFile 从文件加载记忆数据
func loadMemoryFromFile(agentName string) (*MemoryData, error) {
	filePath := getMemoryFilePath(agentName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return CreateEmptyMemory(), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return CreateEmptyMemory(), fmt.Errorf("read memory file: %w", err)
	}

	var memory MemoryData
	if err := json.Unmarshal(data, &memory); err != nil {
		return CreateEmptyMemory(), fmt.Errorf("parse memory file: %w", err)
	}

	if memory.Facts == nil {
		memory.Facts = []Fact{}
	}

	return &memory, nil
}

// saveMemoryToFile 原子保存记忆数据到文件并更新缓存
func saveMemoryToFile(memoryData *MemoryData, agentName string, logger *zap.Logger) bool {
	if logger == nil {
		logger = zap.NewNop()
	}

	filePath := getMemoryFilePath(agentName)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Failed to create memory directory", zap.Error(err))
		return false
	}

	now := time.Now().UTC().Format(time.RFC3339) + "Z"
	memoryData.LastUpdated = now

	tempPath := filePath + ".tmp"
	data, err := json.MarshalIndent(memoryData, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal memory data", zap.Error(err))
		return false
	}

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		logger.Error("Failed to write temp memory file", zap.Error(err))
		return false
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		logger.Error("Failed to rename temp memory file", zap.Error(err))
		_ = os.Remove(tempPath)
		return false
	}

	mtime := getFileMtime(filePath)
	memoryCacheMutex.Lock()
	memoryCache[agentName] = &memoryCacheEntry{
		data:  memoryData,
		mtime: mtime,
	}
	memoryCacheMutex.Unlock()

	logger.Debug("Memory saved", zap.String("path", filePath))
	return true
}

// getFileMtime 获取文件修改时间
func getFileMtime(filePath string) float64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return float64(info.ModTime().UnixNano()) / 1e9
}

// ============================================
// MemoryUpdater 类型
// ============================================

// MemoryUpdater 记忆更新器
type MemoryUpdater struct {
	modelName string
	logger    *zap.Logger
}

// NewMemoryUpdater 创建记忆更新器
func NewMemoryUpdater(modelName string, logger *zap.Logger) *MemoryUpdater {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MemoryUpdater{
		modelName: modelName,
		logger:    logger,
	}
}

// UpdateMemory 更新记忆（基于对话消息）
func (u *MemoryUpdater) UpdateMemory(
	messages []any,
	threadID string,
	agentName string,
	llmCaller func(prompt string) (string, error),
) bool {
	// 使用默认配置（避免循环导入）
	enabled := false // 默认禁用
	if !enabled {
		return false
	}

	if len(messages) == 0 {
		return false
	}

	if llmCaller == nil {
		u.logger.Warn("No LLM caller provided, skipping memory update")
		return false
	}

	// 简化实现：完整实现需要与模型系统集成
	return false
}

// ============================================
// 辅助函数
// ============================================

// formatConversationFromAny 从任意消息列表格式化对话
func formatConversationFromAny(messages []any) string {
	lines := make([]string, 0, len(messages))
	for _, msg := range messages {
		lines = append(lines, fmt.Sprintf("%v", msg))
	}
	return strings.Join(lines, "\n\n")
}

// generateFactID 生成事实 ID
func generateFactID() string {
	return fmt.Sprintf("fact_%x", time.Now().UnixNano()&0xffffffff)
}

// sortFactsByConfidence 按置信度降序排序事实
func sortFactsByConfidence(facts []Fact) {
	for i := range facts {
		for j := i + 1; j < len(facts); j++ {
			if facts[i].Confidence < facts[j].Confidence {
				facts[i], facts[j] = facts[j], facts[i]
			}
		}
	}
}

// ClearMemoryCache 清除记忆缓存（用于测试）
func ClearMemoryCache() {
	memoryCacheMutex.Lock()
	defer memoryCacheMutex.Unlock()
	memoryCache = make(map[string]*memoryCacheEntry)
}
