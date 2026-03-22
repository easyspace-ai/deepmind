package orchestrator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weibaohui/nanobot-go/pkg/planner/model"
	"github.com/weibaohui/nanobot-go/pkg/planner/orchestrator"
	"go.uber.org/zap"
)

func TestWorkflowOrchestrator_BuildWorkflow_Sequential(t *testing.T) {
	logger := zap.NewNop()
	o := orchestrator.NewWorkflowOrchestrator(logger)

	tasks := []*model.SubTask{
		{
			ID:        "task1",
			Name:      "Task 1",
			Type:      model.SubTaskTypeAnalyze,
			DependsOn: []string{},
			Parallel:  false,
		},
		{
			ID:        "task2",
			Name:      "Task 2",
			Type:      model.SubTaskTypeCreate,
			DependsOn: []string{"task1"},
			Parallel:  false,
		},
		{
			ID:        "task3",
			Name:      "Task 3",
			Type:      model.SubTaskTypeTest,
			DependsOn: []string{"task2"},
			Parallel:  false,
		},
	}

	workflow, err := o.BuildWorkflow(context.Background(), tasks)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	assert.Equal(t, model.WorkflowTypeSequential, workflow.Type)
	assert.Len(t, workflow.Stages, 3)
	assert.Equal(t, 3, workflow.TaskCount())

	// 验证阶段顺序
	assert.Equal(t, model.WorkflowTypeTask, workflow.Stages[0].Type)
	assert.Equal(t, "task1", workflow.Stages[0].Task.ID)
	assert.Equal(t, model.WorkflowTypeTask, workflow.Stages[1].Type)
	assert.Equal(t, "task2", workflow.Stages[1].Task.ID)
	assert.Equal(t, model.WorkflowTypeTask, workflow.Stages[2].Type)
	assert.Equal(t, "task3", workflow.Stages[2].Task.ID)
}

func TestWorkflowOrchestrator_BuildWorkflow_Parallel(t *testing.T) {
	logger := zap.NewNop()
	o := orchestrator.NewWorkflowOrchestrator(logger)

	tasks := []*model.SubTask{
		{
			ID:        "task1",
			Name:      "Task 1",
			Type:      model.SubTaskTypeAnalyze,
			DependsOn: []string{},
			Parallel:  false,
		},
		{
			ID:        "task2a",
			Name:      "Task 2a",
			Type:      model.SubTaskTypeTest,
			DependsOn: []string{"task1"},
			Parallel:  true,
		},
		{
			ID:        "task2b",
			Name:      "Task 2b",
			Type:      model.SubTaskTypeTest,
			DependsOn: []string{"task1"},
			Parallel:  true,
		},
		{
			ID:        "task3",
			Name:      "Task 3",
			Type:      model.SubTaskTypeVerify,
			DependsOn: []string{"task2a", "task2b"},
			Parallel:  false,
		},
	}

	workflow, err := o.BuildWorkflow(context.Background(), tasks)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	assert.Equal(t, model.WorkflowTypeSequential, workflow.Type)
	assert.Len(t, workflow.Stages, 3)
	assert.Equal(t, 4, workflow.TaskCount())

	// 验证中间是并行阶段
	assert.Equal(t, model.WorkflowTypeParallel, workflow.Stages[1].Type)
	assert.Len(t, workflow.Stages[1].Tasks, 2)
}

func TestWorkflowOrchestrator_BuildWorkflow_MixedParallel(t *testing.T) {
	logger := zap.NewNop()
	o := orchestrator.NewWorkflowOrchestrator(logger)

	tasks := []*model.SubTask{
		{
			ID:        "task1",
			Name:      "Task 1",
			Type:      model.SubTaskTypeAnalyze,
			DependsOn: []string{},
			Parallel:  false,
		},
		{
			ID:        "task2a",
			Name:      "Task 2a",
			Type:      model.SubTaskTypeTest,
			DependsOn: []string{"task1"},
			Parallel:  true, // 可并行
		},
		{
			ID:        "task2b",
			Name:      "Task 2b",
			Type:      model.SubTaskTypeTest,
			DependsOn: []string{"task1"},
			Parallel:  false, // 不可并行
		},
		{
			ID:        "task3",
			Name:      "Task 3",
			Type:      model.SubTaskTypeVerify,
			DependsOn: []string{"task2a", "task2b"},
			Parallel:  false,
		},
	}

	workflow, err := o.BuildWorkflow(context.Background(), tasks)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)

	// 因为 task2b 的 Parallel=false，它们应该被拆分为顺序阶段
	assert.GreaterOrEqual(t, len(workflow.Stages), 3)
}

func TestWorkflowOrchestrator_BuildWorkflow_InvalidTasks(t *testing.T) {
	logger := zap.NewNop()
	o := orchestrator.NewWorkflowOrchestrator(logger)

	// 循环依赖的任务
	tasks := []*model.SubTask{
		{
			ID:        "task1",
			Name:      "Task 1",
			Type:      model.SubTaskTypeAnalyze,
			DependsOn: []string{"task2"},
		},
		{
			ID:        "task2",
			Name:      "Task 2",
			Type:      model.SubTaskTypeCreate,
			DependsOn: []string{"task1"},
		},
	}

	workflow, err := o.BuildWorkflow(context.Background(), tasks)
	assert.Error(t, err)
	assert.Nil(t, workflow)
}

func TestWorkflowOrchestrator_BuildWorkflow_EmptyTasks(t *testing.T) {
	logger := zap.NewNop()
	o := orchestrator.NewWorkflowOrchestrator(logger)

	workflow, err := o.BuildWorkflow(context.Background(), []*model.SubTask{})
	assert.Error(t, err)
	assert.Nil(t, workflow)
}
