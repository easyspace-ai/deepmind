package model

// WorkflowType 工作流类型
type WorkflowType string

const (
	WorkflowTypeSequential WorkflowType = "sequential" // 顺序工作流
	WorkflowTypeParallel   WorkflowType = "parallel"   // 并行工作流
	WorkflowTypeTask       WorkflowType = "task"       // 单个任务
)

// WorkflowStage 工作流阶段
type WorkflowStage struct {
	Type      WorkflowType     `json:"type"`                // 阶段类型
	Task      *SubTask         `json:"task,omitempty"`      // 单个任务（type=task 时）
	Tasks     []*SubTask       `json:"tasks,omitempty"`     // 多个任务（type=parallel 时）
	SubStages []*WorkflowStage `json:"subStages,omitempty"` // 子阶段（嵌套）
}

// Workflow 工作流
type Workflow struct {
	Type   WorkflowType     `json:"type"`   // 工作流类型
	Stages []*WorkflowStage `json:"stages"` // 阶段列表
}

// NewSequentialWorkflow 创建顺序工作流
func NewSequentialWorkflow(stages []*WorkflowStage) *Workflow {
	return &Workflow{
		Type:   WorkflowTypeSequential,
		Stages: stages,
	}
}

// NewTaskStage 创建任务阶段
func NewTaskStage(task *SubTask) *WorkflowStage {
	return &WorkflowStage{
		Type: WorkflowTypeTask,
		Task: task,
	}
}

// NewParallelStage 创建并行阶段
func NewParallelStage(tasks []*SubTask) *WorkflowStage {
	return &WorkflowStage{
		Type:  WorkflowTypeParallel,
		Tasks: tasks,
	}
}

// TaskCount 返回工作流中的任务总数
func (w *Workflow) TaskCount() int {
	count := 0
	for _, stage := range w.Stages {
		count += stage.TaskCount()
	}
	return count
}

// TaskCount 返回阶段中的任务总数
func (s *WorkflowStage) TaskCount() int {
	switch s.Type {
	case WorkflowTypeTask:
		if s.Task != nil {
			return 1
		}
		return 0
	case WorkflowTypeParallel:
		return len(s.Tasks)
	case WorkflowTypeSequential:
		count := 0
		for _, sub := range s.SubStages {
			count += sub.TaskCount()
		}
		return count
	default:
		return 0
	}
}
