// Package service 提供规划系统的服务层
package service

import (
	"context"
	"fmt"

	"github.com/weibaohui/nanobot-go/pkg/planner/agent"
	"github.com/weibaohui/nanobot-go/pkg/planner/model"
	"github.com/weibaohui/nanobot-go/pkg/planner/orchestrator"
	"go.uber.org/zap"
)

// PlannerService 规划服务
type PlannerService struct {
	intentAnalyzer       *agent.IntentAnalyzer
	taskDecomposer       *agent.TaskDecomposer
	workflowOrchestrator *orchestrator.WorkflowOrchestrator
	logger               *zap.Logger
}

// PlannerServiceConfig 配置
type PlannerServiceConfig struct {
	IntentAnalyzer       *agent.IntentAnalyzer
	TaskDecomposer       *agent.TaskDecomposer
	WorkflowOrchestrator *orchestrator.WorkflowOrchestrator
	Logger               *zap.Logger
}

// NewPlannerService 创建规划服务
func NewPlannerService(cfg *PlannerServiceConfig) *PlannerService {
	logger := cfg.Logger
	if logger == nil {
		logger = zap.L()
	}

	return &PlannerService{
		intentAnalyzer:       cfg.IntentAnalyzer,
		taskDecomposer:       cfg.TaskDecomposer,
		workflowOrchestrator: cfg.WorkflowOrchestrator,
		logger:               logger,
	}
}

// AnalyzeIntent 分析意图
func (s *PlannerService) AnalyzeIntent(ctx context.Context, query string) (*model.IntentAnalysis, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	return s.intentAnalyzer.Analyze(ctx, query)
}

// DecomposeTasks 分解任务
func (s *PlannerService) DecomposeTasks(ctx context.Context, query string, intent *model.IntentAnalysis) ([]*model.SubTask, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if intent == nil {
		return nil, fmt.Errorf("intent is required")
	}
	if !intent.IsValid() {
		return nil, fmt.Errorf("intent is invalid")
	}

	return s.taskDecomposer.Decompose(ctx, query, intent)
}

// BuildWorkflow 构建工作流
func (s *PlannerService) BuildWorkflow(ctx context.Context, tasks []*model.SubTask) (*model.Workflow, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("tasks are required")
	}

	return s.workflowOrchestrator.BuildWorkflow(ctx, tasks)
}

// Plan 完整规划流程
func (s *PlannerService) Plan(ctx context.Context, query string) (*PlanResult, error) {
	s.logger.Info("Starting planning", zap.String("query", query))

	// 1. 分析意图
	intent, err := s.AnalyzeIntent(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze intent: %w", err)
	}

	// 2. 分解任务
	tasks, err := s.DecomposeTasks(ctx, query, intent)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose tasks: %w", err)
	}

	// 3. 构建工作流
	workflow, err := s.BuildWorkflow(ctx, tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow: %w", err)
	}

	result := &PlanResult{
		Intent:   intent,
		Tasks:    tasks,
		Workflow: workflow,
	}

	s.logger.Info("Planning completed",
		zap.String("taskType", string(intent.TaskType)),
		zap.Int("taskCount", len(tasks)),
		zap.Int("workflowStages", len(workflow.Stages)))

	return result, nil
}

// PlanResult 规划结果
type PlanResult struct {
	Intent   *model.IntentAnalysis `json:"intent"`
	Tasks    []*model.SubTask      `json:"tasks"`
	Workflow *model.Workflow        `json:"workflow"`
}
