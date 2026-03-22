package deerflow

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/google/uuid"
	"github.com/weibaohui/nanobot-go/pkg/agent/subagent"
	"go.uber.org/zap"
)

// ============================================
// Task 工具
// 一比一复刻 DeerFlow 的 task_tool
// ============================================

// TaskTool task 工具
type TaskTool struct {
	*BaseDeerFlowTool
	logger *zap.Logger
}

// NewTaskTool 创建 task 工具
func NewTaskTool(logger *zap.Logger) tool.BaseTool {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &TaskTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"task",
			`Delegate a task to a specialized subagent that runs in its own context.

Subagents help you:
- Preserve context by keeping exploration and implementation separate
- Handle complex multi-step tasks autonomously
- Execute commands or operations in isolated contexts

Available subagent types:
- **general-purpose**: A capable agent for complex, multi-step tasks that require
  both exploration and action. Use when the task requires complex reasoning,
  multiple dependent steps, or would benefit from isolated context.
- **bash**: Command execution specialist for running bash commands. Use for
  git operations, build processes, or when command output would be verbose.

When to use this tool:
- Complex tasks requiring multiple steps or tools
- Tasks that produce verbose output
- When you want to isolate context from the main conversation
- Parallel research or exploration tasks

When NOT to use this tool:
- Simple, single-step operations (use tools directly)
- Tasks requiring user interaction or clarification`,
			map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "A short (3-5 word) description of the task for logging/display. ALWAYS PROVIDE THIS PARAMETER FIRST.",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "The task description for the subagent. Be specific and clear about what needs to be done. ALWAYS PROVIDE THIS PARAMETER SECOND.",
				},
				"subagent_type": map[string]interface{}{
					"type":        "string",
					"description": "The type of subagent to use. ALWAYS PROVIDE THIS PARAMETER THIRD.",
					"enum": []string{
						"general-purpose",
						"bash",
					},
				},
				"max_turns": map[string]interface{}{
					"type":        "integer",
					"description": "Optional maximum number of agent turns. Defaults to subagent's configured max.",
				},
			},
		),
		logger: logger,
	}
}

// TaskToolResult task 工具执行结果
type TaskToolResult struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Invoke 执行 task 工具
func (t *TaskTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// 解析参数
	description, _ := args["description"].(string)
	prompt, _ := args["prompt"].(string)
	subagentType, _ := args["subagent_type"].(string)
	maxTurnsIf, _ := args["max_turns"]
	var maxTurns *int
	if maxTurnsIf != nil {
		if mt, ok := maxTurnsIf.(int); ok {
			maxTurns = &mt
		}
	}

	// 验证必需参数
	if subagentType == "" {
		return TaskToolResult{
			Success: false,
			Error:   "Error: subagent_type is required. Available: general-purpose, bash",
		}, nil
	}
	if prompt == "" {
		return TaskToolResult{
			Success: false,
			Error:   "Error: prompt is required",
		}, nil
	}
	if description == "" {
		description = "Task"
	}

	// 获取子代理配置
	config := subagent.GetSubagentConfig(subagentType)
	if config == nil {
		return TaskToolResult{
			Success: false,
			Error:   fmt.Sprintf("Error: Unknown subagent type '%s'. Available: general-purpose, bash", subagentType),
		}, nil
	}

	// 应用 max_turns 覆盖
	if maxTurns != nil && *maxTurns > 0 {
		configCopy := *config
		configCopy.MaxTurns = *maxTurns
		config = &configCopy
	}

	// 生成 trace_id
	traceID := uuid.New().String()[:8]

	// 获取事件总线
	eventBus := subagent.GetGlobalEventBus()

	// 创建工具调用 ID 作为 task_id
	taskID := uuid.New().String()[:8]

	t.logger.Info("Starting background task",
		zap.String("trace_id", traceID),
		zap.String("task_id", taskID),
		zap.String("subagent_type", subagentType),
		zap.Int("timeout_seconds", config.TimeoutSeconds))

	// 发送 Task Started 事件
	eventBus.Publish(subagent.NewTaskStartedEvent(taskID, description))

	// TODO: 从上下文中获取实际的工具、sandboxState、threadData、threadID、parentModel
	// 这里使用占位符
	var tools []string
	var sandboxState any
	var threadData any
	var threadID string
	var parentModel string

	// 创建执行器
	executor := subagent.NewSubagentExecutor(
		config,
		tools,
		parentModel,
		sandboxState,
		threadData,
		threadID,
		traceID,
		t.logger,
	)

	// 同步执行（演示用，实际应该是异步）
	// TODO: 实现真正的异步执行和轮询
	result := executor.Execute(prompt, nil)

	// 根据结果发送事件
	switch result.Status {
	case subagent.SubagentStatusCompleted:
		eventBus.Publish(subagent.NewTaskCompletedEvent(taskID, result.Result))
		t.logger.Info("Task completed",
			zap.String("trace_id", traceID),
			zap.String("task_id", taskID))
		return TaskToolResult{
			Success: true,
			Result:  fmt.Sprintf("Task Succeeded. Result: %s", result.Result),
		}, nil

	case subagent.SubagentStatusFailed:
		eventBus.Publish(subagent.NewTaskFailedEvent(taskID, result.Error))
		t.logger.Error("Task failed",
			zap.String("trace_id", traceID),
			zap.String("task_id", taskID),
			zap.String("error", result.Error))
		return TaskToolResult{
			Success: false,
			Error:   fmt.Sprintf("Task failed. Error: %s", result.Error),
		}, nil

	case subagent.SubagentStatusTimedOut:
		eventBus.Publish(subagent.NewTaskTimedOutEvent(taskID, result.Error))
		t.logger.Warn("Task timed out",
			zap.String("trace_id", traceID),
			zap.String("task_id", taskID),
			zap.String("error", result.Error))
		return TaskToolResult{
			Success: false,
			Error:   fmt.Sprintf("Task timed out. Error: %s", result.Error),
		}, nil

	default:
		return TaskToolResult{
			Success: false,
			Error:   fmt.Sprintf("Task in unexpected state: %s", result.Status),
		}, nil
	}
}

// ============================================
// 轮询辅助函数
// ============================================

// TaskPoller 任务轮询器
type TaskPoller struct {
	taskID        string
	pollInterval  time.Duration
	maxPolls      int
	logger        *zap.Logger
	eventBus      *subagent.TaskEventBus
}

// NewTaskPoller 创建任务轮询器
func NewTaskPoller(taskID string, timeoutSeconds int, logger *zap.Logger) *TaskPoller {
	if logger == nil {
		logger = zap.NewNop()
	}

	pollInterval := 5 * time.Second
	maxPolls := (timeoutSeconds + 60) / 5 // 超时+60秒缓冲，每5秒轮询一次

	return &TaskPoller{
		taskID:       taskID,
		pollInterval: pollInterval,
		maxPolls:     maxPolls,
		logger:       logger,
		eventBus:     subagent.GetGlobalEventBus(),
	}
}

// Poll 轮询任务直到完成
func (p *TaskPoller) Poll(ctx context.Context) (*subagent.SubagentResult, error) {
	pollCount := 0
	lastStatus := subagent.SubagentStatus("")
	lastMessageCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result := subagent.GetBackgroundTaskResult(p.taskID)
		if result == nil {
			p.logger.Error("Task disappeared from background tasks",
				zap.String("task_id", p.taskID))
			p.eventBus.Publish(subagent.NewTaskFailedEvent(p.taskID, "Task disappeared from background tasks"))
			subagent.CleanupBackgroundTask(p.taskID, p.logger)
			return nil, fmt.Errorf("task %s disappeared", p.taskID)
		}

		// 记录状态变化
		if result.Status != lastStatus {
			p.logger.Info("Task status update",
				zap.String("task_id", p.taskID),
				zap.String("status", string(result.Status)))
			lastStatus = result.Status
		}

		// 检查新消息并发送 task_running 事件
		currentMessageCount := len(result.AIMessages)
		if currentMessageCount > lastMessageCount {
			for i := lastMessageCount; i < currentMessageCount; i++ {
				message := result.AIMessages[i]
				p.eventBus.Publish(subagent.NewTaskRunningEvent(
					p.taskID,
					message,
					i+1,
					currentMessageCount,
				))
				p.logger.Info("Sent task running event",
					zap.String("task_id", p.taskID),
					zap.Int("message_index", i+1),
					zap.Int("total_messages", currentMessageCount))
			}
			lastMessageCount = currentMessageCount
		}

		// 检查是否完成
		switch result.Status {
		case subagent.SubagentStatusCompleted:
			p.eventBus.Publish(subagent.NewTaskCompletedEvent(p.taskID, result.Result))
			p.logger.Info("Task completed after polling",
				zap.String("task_id", p.taskID),
				zap.Int("poll_count", pollCount))
			subagent.CleanupBackgroundTask(p.taskID, p.logger)
			return result, nil

		case subagent.SubagentStatusFailed:
			p.eventBus.Publish(subagent.NewTaskFailedEvent(p.taskID, result.Error))
			p.logger.Error("Task failed after polling",
				zap.String("task_id", p.taskID),
				zap.String("error", result.Error))
			subagent.CleanupBackgroundTask(p.taskID, p.logger)
			return result, fmt.Errorf("task failed: %s", result.Error)

		case subagent.SubagentStatusTimedOut:
			p.eventBus.Publish(subagent.NewTaskTimedOutEvent(p.taskID, result.Error))
			p.logger.Warn("Task timed out after polling",
				zap.String("task_id", p.taskID),
				zap.String("error", result.Error))
			subagent.CleanupBackgroundTask(p.taskID, p.logger)
			return result, fmt.Errorf("task timed out: %s", result.Error)
		}

		// 等待下一次轮询
		time.Sleep(p.pollInterval)
		pollCount++

		// 轮询超时安全网
		if pollCount > p.maxPolls {
			timeoutMinutes := p.maxPolls * 5 / 60
			p.logger.Error("Task polling timed out",
				zap.String("task_id", p.taskID),
				zap.Int("poll_count", pollCount),
				zap.String("status", string(result.Status)))
			p.eventBus.Publish(subagent.NewTaskTimedOutEvent(p.taskID, ""))
			return result, fmt.Errorf("task polling timed out after %d minutes", timeoutMinutes)
		}
	}
}
