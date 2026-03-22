package subagent

import (
	"sync"
	"time"

	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// ============================================
// SubagentStatus - 子代理执行状态
// ============================================

// SubagentStatus 子代理执行状态
type SubagentStatus string

const (
	SubagentStatusPending   SubagentStatus = "pending"
	SubagentStatusRunning   SubagentStatus = "running"
	SubagentStatusCompleted SubagentStatus = "completed"
	SubagentStatusFailed    SubagentStatus = "failed"
	SubagentStatusTimedOut  SubagentStatus = "timed_out"
)

// IsTerminal 判断状态是否为终态
func (s SubagentStatus) IsTerminal() bool {
	return s == SubagentStatusCompleted || s == SubagentStatusFailed || s == SubagentStatusTimedOut
}

// ============================================
// SubagentResult - 子代理执行结果
// ============================================

// SubagentResult 子代理执行结果
// 一比一复刻 DeerFlow 的 SubagentResult
type SubagentResult struct {
	TaskID      string                 `json:"task_id"`
	TraceID     string                 `json:"trace_id"`
	Status      SubagentStatus         `json:"status"`
	Result      string                 `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	AIMessages  []map[string]any       `json:"ai_messages,omitempty"`
}

// NewSubagentResult 创建子代理执行结果
func NewSubagentResult(taskID, traceID string) *SubagentResult {
	return &SubagentResult{
		TaskID:     taskID,
		TraceID:    traceID,
		Status:     SubagentStatusPending,
		AIMessages: make([]map[string]any, 0),
	}
}

// ============================================
// SubagentConfig - 子代理配置
// ============================================

// SubagentConfig 子代理配置
// 一比一复刻 DeerFlow 的 SubagentConfig
type SubagentConfig struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	SystemPrompt      string   `json:"system_prompt"`
	Tools             []string `json:"tools,omitempty"`
	DisallowedTools   []string `json:"disallowed_tools,omitempty"`
	Model             string   `json:"model"` // "inherit" 或具体模型名
	MaxTurns          int      `json:"max_turns"`
	TimeoutSeconds    int      `json:"timeout_seconds"`
}

// DefaultSubagentConfig 默认子代理配置
func DefaultSubagentConfig() *SubagentConfig {
	return &SubagentConfig{
		Model:           "inherit",
		MaxTurns:        50,
		TimeoutSeconds:  900, // 15分钟
		DisallowedTools: []string{"task"},
	}
}

// ============================================
// 全局任务存储
// ============================================

var (
	backgroundTasks     = make(map[string]*SubagentResult)
	backgroundTasksLock sync.RWMutex
)

// GetBackgroundTaskResult 获取后台任务结果
func GetBackgroundTaskResult(taskID string) *SubagentResult {
	backgroundTasksLock.RLock()
	defer backgroundTasksLock.RUnlock()
	return backgroundTasks[taskID]
}

// ListBackgroundTasks 列出所有后台任务
func ListBackgroundTasks() []*SubagentResult {
	backgroundTasksLock.RLock()
	defer backgroundTasksLock.RUnlock()

	results := make([]*SubagentResult, 0, len(backgroundTasks))
	for _, result := range backgroundTasks {
		results = append(results, result)
	}
	return results
}

// CleanupBackgroundTask 清理已完成的后台任务
// 只能清理终态任务，避免与后台执行器的竞态条件
func CleanupBackgroundTask(taskID string, logger *zap.Logger) {
	if logger == nil {
		logger = zap.NewNop()
	}

	backgroundTasksLock.Lock()
	defer backgroundTasksLock.Unlock()

	result := backgroundTasks[taskID]
	if result == nil {
		logger.Debug("Requested cleanup for unknown background task", zap.String("task_id", taskID))
		return
	}

	// 只清理终态任务
	if result.Status.IsTerminal() || result.CompletedAt != nil {
		delete(backgroundTasks, taskID)
		logger.Debug("Cleaned up background task", zap.String("task_id", taskID))
	} else {
		logger.Debug("Skipping cleanup for non-terminal background task",
			zap.String("task_id", taskID),
			zap.String("status", string(result.Status)))
	}
}

// ============================================
// 工具过滤
// ============================================

// FilterTools 根据子代理配置过滤工具
func FilterTools(allTools []string, allowed, disallowed []string) []string {
	filtered := make([]string, 0, len(allTools))

	// 创建允许集合
	allowedSet := make(map[string]bool)
	if allowed != nil {
		for _, t := range allowed {
			allowedSet[t] = true
		}
	}

	// 创建禁止集合
	disallowedSet := make(map[string]bool)
	if disallowed != nil {
		for _, t := range disallowed {
			disallowedSet[t] = true
		}
	}

	// 应用过滤
	for _, tool := range allTools {
		// 如果有允许列表，只允许在列表中的
		if allowed != nil && !allowedSet[tool] {
			continue
		}
		// 如果在禁止列表中，跳过
		if disallowedSet[tool] {
			continue
		}
		filtered = append(filtered, tool)
	}

	return filtered
}

// ============================================
// 模型名称解析
// ============================================

// ResolveModelName 解析子代理使用的模型名称
func ResolveModelName(config *SubagentConfig, parentModel string) string {
	if config.Model == "inherit" {
		return parentModel
	}
	return config.Model
}

// ============================================
// 子代理执行上下文
// ============================================

// SubagentExecutorContext 子代理执行上下文
type SubagentExecutorContext struct {
	Config        *SubagentConfig
	ParentModel   string
	SandboxState  *state.SandboxState
	ThreadData    *state.ThreadDataState
	ThreadID      string
	TraceID       string
	AvailableTools []string
}

// NewSubagentExecutorContext 创建子代理执行上下文
func NewSubagentExecutorContext(
	config *SubagentConfig,
	parentModel string,
	sandboxState *state.SandboxState,
	threadData *state.ThreadDataState,
	threadID string,
	traceID string,
	availableTools []string,
) *SubagentExecutorContext {
	return &SubagentExecutorContext{
		Config:        config,
		ParentModel:   parentModel,
		SandboxState:  sandboxState,
		ThreadData:    threadData,
		ThreadID:      threadID,
		TraceID:       traceID,
		AvailableTools: availableTools,
	}
}
