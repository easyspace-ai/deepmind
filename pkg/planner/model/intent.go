// Package model 定义规划系统的数据模型
package model

import "fmt"

// TaskType 任务类型
type TaskType string

const (
	TaskTypeCode     TaskType = "code"     // 编码任务
	TaskTypeDebug    TaskType = "debug"    // 调试任务
	TaskTypeRefactor TaskType = "refactor" // 重构任务
	TaskTypeDocument TaskType = "document" // 文档任务
	TaskTypeTest     TaskType = "test"     // 测试任务
	TaskTypeOther    TaskType = "other"    // 其他任务
)

// Complexity 复杂度
type Complexity string

const (
	ComplexitySimple  Complexity = "simple"  // 简单
	ComplexityMedium  Complexity = "medium"  // 中等
	ComplexityComplex Complexity = "complex" // 复杂
)

// Scope 范围
type Scope string

const (
	ScopeFile    Scope = "file"    // 文件级
	ScopePackage Scope = "package" // 包级
	ScopeProject Scope = "project" // 项目级
)

// IntentAnalysis 意图分析结果
type IntentAnalysis struct {
	TaskType        TaskType   `json:"taskType"`        // 任务类型
	Complexity      Complexity `json:"complexity"`      // 复杂度
	Scope           Scope      `json:"scope"`           // 范围
	Technologies    []string   `json:"technologies"`    // 涉及的技术栈
	Dependencies    []string   `json:"dependencies"`    // 依赖项
	Constraints     []string   `json:"constraints"`     // 约束条件
	SuccessCriteria []string   `json:"successCriteria"` // 成功标准
	RawQuery        string     `json:"rawQuery"`        // 原始查询
}

// IsValid 验证意图分析结果是否有效
func (i *IntentAnalysis) IsValid() bool {
	if i.TaskType == "" {
		return false
	}
	if i.Complexity == "" {
		return false
	}
	if i.Scope == "" {
		return false
	}
	return true
}

// String 返回字符串表示
func (i *IntentAnalysis) String() string {
	return fmt.Sprintf("IntentAnalysis{Type=%s, Complexity=%s, Scope=%s}",
		i.TaskType, i.Complexity, i.Scope)
}
