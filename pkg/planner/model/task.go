package model

import "fmt"

// SubTaskType 子任务类型
type SubTaskType string

const (
	SubTaskTypeAnalyze SubTaskType = "analyze" // 分析
	SubTaskTypeCreate  SubTaskType = "create"  // 创建
	SubTaskTypeModify  SubTaskType = "modify"  // 修改
	SubTaskTypeDelete  SubTaskType = "delete"  // 删除
	SubTaskTypeTest    SubTaskType = "test"    // 测试
	SubTaskTypeVerify  SubTaskType = "verify"  // 验证
)

// AgentType Agent 类型
type AgentType string

const (
	AgentTypeChat   AgentType = "chat"   // 聊天 Agent
	AgentTypeTool   AgentType = "tool"   // 工具 Agent
	AgentTypeCustom AgentType = "custom" // 自定义 Agent
)

// SubTask 子任务
type SubTask struct {
	ID          string      `json:"id"`          // 任务唯一标识
	Name        string      `json:"name"`        // 任务名称
	Description string      `json:"description"` // 任务描述
	Type        SubTaskType `json:"type"`        // 任务类型
	AgentType   AgentType   `json:"agentType"`   // Agent 类型
	Tools       []string    `json:"tools"`       // 需要的工具
	DependsOn   []string    `json:"dependsOn"`   // 依赖的任务 ID
	Parallel    bool        `json:"parallel"`    // 是否可并行
}

// Validate 验证子任务
func (t *SubTask) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task ID is required")
	}
	if t.Name == "" {
		return fmt.Errorf("task name is required")
	}
	if t.Type == "" {
		return fmt.Errorf("task type is required")
	}
	return nil
}

// HasDependency 检查是否依赖某个任务
func (t *SubTask) HasDependency(taskID string) bool {
	for _, dep := range t.DependsOn {
		if dep == taskID {
			return true
		}
	}
	return false
}

// TaskList 任务列表
type TaskList struct {
	Tasks []*SubTask `json:"tasks"`
}

// GetTask 通过 ID 获取任务
func (tl *TaskList) GetTask(id string) (*SubTask, bool) {
	for _, t := range tl.Tasks {
		if t.ID == id {
			return t, true
		}
	}
	return nil, false
}

// ValidateAll 验证所有任务
func (tl *TaskList) ValidateAll() error {
	// 验证每个任务
	taskIDs := make(map[string]bool)
	for _, t := range tl.Tasks {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("task %s: %w", t.ID, err)
		}
		if taskIDs[t.ID] {
			return fmt.Errorf("duplicate task ID: %s", t.ID)
		}
		taskIDs[t.ID] = true
	}

	// 验证依赖关系
	for _, t := range tl.Tasks {
		for _, depID := range t.DependsOn {
			if !taskIDs[depID] {
				return fmt.Errorf("task %s depends on non-existent task %s", t.ID, depID)
			}
		}
	}

	// 检测循环依赖
	return tl.detectCyclicDependency()
}

func (tl *TaskList) detectCyclicDependency() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(id string) bool
	dfs = func(id string) bool {
		visited[id] = true
		recStack[id] = true

		task, _ := tl.GetTask(id)
		for _, depID := range task.DependsOn {
			if !visited[depID] {
				if dfs(depID) {
					return true
				}
			} else if recStack[depID] {
				return true
			}
		}

		recStack[id] = false
		return false
	}

	for _, t := range tl.Tasks {
		if !visited[t.ID] {
			if dfs(t.ID) {
				return fmt.Errorf("cyclic dependency detected involving task %s", t.ID)
			}
		}
	}

	return nil
}
