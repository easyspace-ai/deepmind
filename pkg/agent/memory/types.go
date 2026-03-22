package memory

import (
	"time"
)

// ============================================
// Memory 数据结构（一比一复刻 DeerFlow）
// ============================================

// MemoryData 记忆数据结构
type MemoryData struct {
	Version     string         `json:"version"`
	LastUpdated string         `json:"lastUpdated"`
	User        UserContext    `json:"user"`
	History     HistoryContext `json:"history"`
	Facts       []Fact         `json:"facts"`
}

// UserContext 用户上下文
type UserContext struct {
	WorkContext     MemorySection `json:"workContext"`
	PersonalContext MemorySection `json:"personalContext"`
	TopOfMind       MemorySection `json:"topOfMind"`
}

// HistoryContext 历史上下文
type HistoryContext struct {
	RecentMonths      MemorySection `json:"recentMonths"`
	EarlierContext    MemorySection `json:"earlierContext"`
	LongTermBackground MemorySection `json:"longTermBackground"`
}

// MemorySection 记忆节
type MemorySection struct {
	Summary   string `json:"summary"`
	UpdatedAt string `json:"updatedAt"`
}

// Fact 事实条目
type Fact struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	CreatedAt  string  `json:"createdAt"`
	Source     string  `json:"source"`
}

// FactCategory 事实分类
const (
	FactCategoryPreference = "preference"
	FactCategoryKnowledge  = "knowledge"
	FactCategoryContext    = "context"
	FactCategoryBehavior   = "behavior"
	FactCategoryGoal       = "goal"
)

// ============================================
// Memory Update 请求/响应结构
// ============================================

// MemoryUpdateRequest 记忆更新请求（LLM 输出格式）
type MemoryUpdateRequest struct {
	User           UserUpdateRequest    `json:"user"`
	History        HistoryUpdateRequest `json:"history"`
	NewFacts       []FactUpdate         `json:"newFacts"`
	FactsToRemove  []string             `json:"factsToRemove"`
}

// UserUpdateRequest 用户上下文更新
type UserUpdateRequest struct {
	WorkContext     SectionUpdate `json:"workContext"`
	PersonalContext SectionUpdate `json:"personalContext"`
	TopOfMind       SectionUpdate `json:"topOfMind"`
}

// HistoryUpdateRequest 历史上下文更新
type HistoryUpdateRequest struct {
	RecentMonths       SectionUpdate `json:"recentMonths"`
	EarlierContext     SectionUpdate `json:"earlierContext"`
	LongTermBackground SectionUpdate `json:"longTermBackground"`
}

// SectionUpdate 节更新
type SectionUpdate struct {
	Summary     string `json:"summary"`
	ShouldUpdate bool   `json:"shouldUpdate"`
}

// FactUpdate 事实更新
type FactUpdate struct {
	Content    string  `json:"content"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
}

// ============================================
// 辅助函数
// ============================================

// CreateEmptyMemory 创建空记忆结构
func CreateEmptyMemory() *MemoryData {
	now := time.Now().UTC().Format(time.RFC3339) + "Z"
	return &MemoryData{
		Version:     "1.0",
		LastUpdated: now,
		User: UserContext{
			WorkContext:     MemorySection{Summary: "", UpdatedAt: ""},
			PersonalContext: MemorySection{Summary: "", UpdatedAt: ""},
			TopOfMind:       MemorySection{Summary: "", UpdatedAt: ""},
		},
		History: HistoryContext{
			RecentMonths:      MemorySection{Summary: "", UpdatedAt: ""},
			EarlierContext:    MemorySection{Summary: "", UpdatedAt: ""},
			LongTermBackground: MemorySection{Summary: "", UpdatedAt: ""},
		},
		Facts: []Fact{},
	}
}

// factContentKey 获取事实内容的键（用于去重）
func factContentKey(content string) string {
	if content == "" {
		return ""
	}
	return content
}
