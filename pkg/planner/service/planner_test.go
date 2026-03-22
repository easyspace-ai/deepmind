package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weibaohui/nanobot-go/pkg/planner/agent"
	"github.com/weibaohui/nanobot-go/pkg/planner/model"
	"github.com/weibaohui/nanobot-go/pkg/planner/orchestrator"
	"github.com/weibaohui/nanobot-go/pkg/planner/service"
	"go.uber.org/zap"
)

func setupTestService(t *testing.T) *service.PlannerService {
	logger := zap.NewNop()
	ctx := context.Background()

	intentAnalyzer, err := agent.NewIntentAnalyzer(ctx, &agent.IntentAnalyzerConfig{
		Logger: logger,
	})
	assert.NoError(t, err)

	taskDecomposer, err := agent.NewTaskDecomposer(ctx, &agent.TaskDecomposerConfig{
		Logger: logger,
	})
	assert.NoError(t, err)

	workflowOrchestrator := orchestrator.NewWorkflowOrchestrator(logger)

	return service.NewPlannerService(&service.PlannerServiceConfig{
		IntentAnalyzer:       intentAnalyzer,
		TaskDecomposer:       taskDecomposer,
		WorkflowOrchestrator: workflowOrchestrator,
		Logger:               logger,
	})
}

func TestPlannerService_AnalyzeIntent(t *testing.T) {
	svc := setupTestService(t)

	tests := []struct {
		name     string
		query    string
		wantType model.TaskType
	}{
		{
			name:     "test task",
			query:    "给登录功能添加单元测试",
			wantType: model.TaskTypeTest,
		},
		{
			name:     "code task",
			query:    "创建一个新的 API 接口",
			wantType: model.TaskTypeCode,
		},
		{
			name:     "document task",
			query:    "写一个 README 文档",
			wantType: model.TaskTypeDocument,
		},
		{
			name:     "refactor task",
			query:    "重构这部分代码",
			wantType: model.TaskTypeRefactor,
		},
		{
			name:     "debug task",
			query:    "修复这个 bug",
			wantType: model.TaskTypeDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent, err := svc.AnalyzeIntent(context.Background(), tt.query)
			assert.NoError(t, err)
			assert.NotNil(t, intent)
			assert.Equal(t, tt.wantType, intent.TaskType)
			assert.True(t, intent.IsValid())
		})
	}
}

func TestPlannerService_AnalyzeIntent_EmptyQuery(t *testing.T) {
	svc := setupTestService(t)

	intent, err := svc.AnalyzeIntent(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, intent)
}

func TestPlannerService_DecomposeTasks(t *testing.T) {
	svc := setupTestService(t)

	query := "给登录功能添加单元测试"
	intent := &model.IntentAnalysis{
		TaskType:     model.TaskTypeTest,
		Complexity:   model.ComplexityMedium,
		Scope:        model.ScopePackage,
		Technologies: []string{"go", "testing"},
	}

	tasks, err := svc.DecomposeTasks(context.Background(), query, intent)
	assert.NoError(t, err)
	assert.NotEmpty(t, tasks)
	assert.Greater(t, len(tasks), 1)
}

func TestPlannerService_DecomposeTasks_InvalidIntent(t *testing.T) {
	svc := setupTestService(t)

	query := "test"
	intent := &model.IntentAnalysis{} // invalid - missing required fields

	tasks, err := svc.DecomposeTasks(context.Background(), query, intent)
	assert.Error(t, err)
	assert.Nil(t, tasks)
}

func TestPlannerService_BuildWorkflow(t *testing.T) {
	svc := setupTestService(t)

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
	}

	workflow, err := svc.BuildWorkflow(context.Background(), tasks)
	assert.NoError(t, err)
	assert.NotNil(t, workflow)
	assert.Equal(t, 2, workflow.TaskCount())
}

func TestPlannerService_Plan(t *testing.T) {
	svc := setupTestService(t)

	result, err := svc.Plan(context.Background(), "给登录功能添加单元测试")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Intent)
	assert.NotEmpty(t, result.Tasks)
	assert.NotNil(t, result.Workflow)

	assert.True(t, result.Intent.IsValid())
	assert.Greater(t, len(result.Tasks), 0)
	assert.Greater(t, result.Workflow.TaskCount(), 0)
}

func TestPlannerService_Plan_EmptyQuery(t *testing.T) {
	svc := setupTestService(t)

	result, err := svc.Plan(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, result)
}
