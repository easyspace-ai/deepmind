package deerflow

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/tool"
)

// ============================================
// DeferredTool Registry（一比一复刻 DeerFlow）
// ============================================

const (
	// MaxResults 单次搜索最大返回工具数
	MaxResults = 5
)

// DeferredToolEntry 延迟工具条目（轻量级元数据）
type DeferredToolEntry struct {
	Name        string
	Description string
	Tool        tool.BaseTool
}

// DeferredToolRegistry 延迟工具注册表
type DeferredToolRegistry struct {
	mu      sync.RWMutex
	entries []*DeferredToolEntry
}

var (
	globalRegistry     *DeferredToolRegistry
	globalRegistryOnce sync.Once
	globalRegistryMu   sync.Mutex
)

// NewDeferredToolRegistry 创建延迟工具注册表
func NewDeferredToolRegistry() *DeferredToolRegistry {
	return &DeferredToolRegistry{
		entries: make([]*DeferredToolEntry, 0),
	}
}

// GetDeferredRegistry 获取全局延迟工具注册表（单例）
func GetDeferredRegistry() *DeferredToolRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewDeferredToolRegistry()
	})
	return globalRegistry
}

// SetDeferredRegistry 设置全局延迟工具注册表
func SetDeferredRegistry(registry *DeferredToolRegistry) {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()
	globalRegistry = registry
}

// ResetDeferredRegistry 重置全局延迟工具注册表（用于测试）
func ResetDeferredRegistry() {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()
	globalRegistry = nil
	globalRegistryOnce = sync.Once{}
}

// Register 注册延迟工具
func (r *DeferredToolRegistry) Register(t tool.BaseTool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, _ := t.Info(context.Background())
	name := ""
	desc := ""
	if info != nil {
		name = info.Name
		desc = info.Desc
	}

	r.entries = append(r.entries, &DeferredToolEntry{
		Name:        name,
		Description: desc,
		Tool:        t,
	})
}

// Search 搜索延迟工具（支持三种查询形式）
//   - "select:name1,name2" - 精确名称匹配
//   - "+keyword rest" - 名称必须包含 keyword，其余部分排序
//   - "keyword query" - 正则匹配 name + description
func (r *DeferredToolRegistry) Search(query string) []tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if strings.HasPrefix(query, "select:") {
		// 精确选择模式
		namesStr := query[7:]
		nameParts := strings.Split(namesStr, ",")
		nameSet := make(map[string]bool)
		for _, n := range nameParts {
			nameSet[strings.TrimSpace(n)] = true
		}

		result := make([]tool.BaseTool, 0, len(nameSet))
		for _, e := range r.entries {
			if nameSet[e.Name] {
				result = append(result, e.Tool)
			}
			if len(result) >= MaxResults {
				break
			}
		}
		return result
	}

	if strings.HasPrefix(query, "+") {
		// 必需关键词模式
		parts := strings.SplitN(query[1:], " ", 2)
		required := strings.ToLower(parts[0])

		candidates := make([]*DeferredToolEntry, 0)
		for _, e := range r.entries {
			if strings.Contains(strings.ToLower(e.Name), required) {
				candidates = append(candidates, e)
			}
		}

		if len(parts) > 1 {
			// 按其余部分排序
			pattern := parts[1]
			sortByRegexScore(candidates, pattern)
		}

		result := make([]tool.BaseTool, 0, len(candidates))
		for _, e := range candidates {
			result = append(result, e.Tool)
			if len(result) >= MaxResults {
				break
			}
		}
		return result
	}

	// 通用正则搜索
	regex, err := regexp.Compile(`(?i)` + query)
	if err != nil {
		regex, _ = regexp.Compile(`(?i)` + regexp.QuoteMeta(query))
	}

	type scoredEntry struct {
		score int
		entry *DeferredToolEntry
	}
	scored := make([]scoredEntry, 0)

	for _, entry := range r.entries {
		searchable := entry.Name + " " + entry.Description
		if regex.MatchString(searchable) {
			score := 1
			if regex.MatchString(entry.Name) {
				score = 2
			}
			scored = append(scored, scoredEntry{score: score, entry: entry})
		}
	}

	// 按分数降序排序
	for i := range scored {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	result := make([]tool.BaseTool, 0, len(scored))
	for _, se := range scored {
		result = append(result, se.entry.Tool)
		if len(result) >= MaxResults {
			break
		}
	}
	return result
}

// Entries 获取所有条目（只读副本）
func (r *DeferredToolRegistry) Entries() []*DeferredToolEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*DeferredToolEntry, len(r.entries))
	copy(result, r.entries)
	return result
}

// Len 获取条目数量
func (r *DeferredToolRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// ============================================
// 辅助函数
// ============================================

func regexScore(pattern string, entry *DeferredToolEntry) int {
	regex, err := regexp.Compile(`(?i)` + pattern)
	if err != nil {
		regex, _ = regexp.Compile(`(?i)` + regexp.QuoteMeta(pattern))
	}
	searchable := entry.Name + " " + entry.Description
	return len(regex.FindAllString(searchable, -1))
}

func sortByRegexScore(entries []*DeferredToolEntry, pattern string) {
	for i := range entries {
		for j := i + 1; j < len(entries); j++ {
			scoreI := regexScore(pattern, entries[i])
			scoreJ := regexScore(pattern, entries[j])
			if scoreI < scoreJ {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}
