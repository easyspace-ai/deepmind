// Package orchestrator 提供工作流编排功能
package orchestrator

import (
	"context"
	"fmt"

	"github.com/weibaohui/nanobot-go/pkg/planner/model"
	"go.uber.org/zap"
)

// WorkflowOrchestrator 工作流编排器
type WorkflowOrchestrator struct {
	logger *zap.Logger
}

// NewWorkflowOrchestrator 创建工作流编排器
func NewWorkflowOrchestrator(logger *zap.Logger) *WorkflowOrchestrator {
	if logger == nil {
		logger = zap.L()
	}
	return &WorkflowOrchestrator{logger: logger}
}

// BuildWorkflow 构建工作流
func (o *WorkflowOrchestrator) BuildWorkflow(ctx context.Context, tasks []*model.SubTask) (*model.Workflow, error) {
	o.logger.Info("Building workflow", zap.Int("taskCount", len(tasks)))

	// 检查空任务
	if len(tasks) == 0 {
		return nil, fmt.Errorf("tasks are required")
	}

	// 创建任务列表并验证
	taskList := &model.TaskList{Tasks: tasks}
	if err := taskList.ValidateAll(); err != nil {
		return nil, fmt.Errorf("invalid tasks: %w", err)
	}

	// 构建依赖图
	graph := o.buildDependencyGraph(tasks)

	// 拓扑排序
	stages, err := o.topologicalSort(tasks, graph)
	if err != nil {
		return nil, fmt.Errorf("failed to sort tasks: %w", err)
	}

	// 构建工作流阶段
	workflowStages := o.buildStages(stages)

	workflow := model.NewSequentialWorkflow(workflowStages)

	o.logger.Info("Workflow built",
		zap.Int("stageCount", len(workflow.Stages)),
		zap.Int("totalTasks", workflow.TaskCount()))

	return workflow, nil
}

// buildDependencyGraph 构建依赖图
// 返回: graph[taskID] = 依赖该任务的任务 ID 列表
func (o *WorkflowOrchestrator) buildDependencyGraph(tasks []*model.SubTask) map[string][]string {
	graph := make(map[string][]string)

	// 初始化
	for _, t := range tasks {
		graph[t.ID] = []string{}
	}

	// 填充依赖关系
	for _, t := range tasks {
		for _, depID := range t.DependsOn {
			graph[depID] = append(graph[depID], t.ID)
		}
	}

	return graph
}

// topologicalSort 拓扑排序，返回阶段列表
// 每个阶段包含可以并行执行的任务
func (o *WorkflowOrchestrator) topologicalSort(
	tasks []*model.SubTask,
	graph map[string][]string,
) ([][]*model.SubTask, error) {

	// 计算入度
	inDegree := make(map[string]int)
	taskMap := make(map[string]*model.SubTask)

	for _, t := range tasks {
		inDegree[t.ID] = len(t.DependsOn)
		taskMap[t.ID] = t
	}

	var stages [][]*model.SubTask

	remaining := len(tasks)
	for remaining > 0 {
		// 找出当前入度为 0 的任务
		var currentStage []*model.SubTask
		for _, t := range tasks {
			if inDegree[t.ID] == 0 {
				currentStage = append(currentStage, t)
			}
		}

		if len(currentStage) == 0 {
			return nil, fmt.Errorf("cyclic dependency detected")
		}

		stages = append(stages, currentStage)
		remaining -= len(currentStage)

		// 减少依赖这些任务的其他任务的入度
		for _, t := range currentStage {
			inDegree[t.ID] = -1 // 标记为已处理
			for _, dependentID := range graph[t.ID] {
				inDegree[dependentID]--
			}
		}
	}

	return stages, nil
}

// buildStages 构建工作流阶段
func (o *WorkflowOrchestrator) buildStages(sortedStages [][]*model.SubTask) []*model.WorkflowStage {
	var workflowStages []*model.WorkflowStage

	for _, stageTasks := range sortedStages {
		if len(stageTasks) == 1 {
			// 单个任务
			workflowStages = append(workflowStages, model.NewTaskStage(stageTasks[0]))
		} else {
			// 检查是否都标记为 parallel
			allParallel := true
			for _, t := range stageTasks {
				if !t.Parallel {
					allParallel = false
					break
				}
			}

			if allParallel {
				// 并行阶段
				workflowStages = append(workflowStages, model.NewParallelStage(stageTasks))
			} else {
				// 不能并行，拆分为多个顺序阶段
				for _, t := range stageTasks {
					workflowStages = append(workflowStages, model.NewTaskStage(t))
				}
			}
		}
	}

	return workflowStages
}
