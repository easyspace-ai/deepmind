// Package planner 提供基于 Eino 框架的自动规划系统
//
// 该包实现了类似 Claude Code 的自动规划能力，包括：
//   - 意图分析：自动识别用户请求的任务类型、复杂度、范围
//   - 任务分解：将复杂任务拆分为可执行的子任务
//   - 工作流编排：自动构建顺序或并行的工作流
//
// 基本用法：
//
//	import "github.com/weibaohui/nanobot-go/pkg/planner"
//
//	// 创建组件
//	intentAnalyzer, _ := planner.NewIntentAnalyzer(ctx, &agent.IntentAnalyzerConfig{})
//	taskDecomposer, _ := planner.NewTaskDecomposer(ctx, &agent.TaskDecomposerConfig{})
//	workflowOrchestrator := planner.NewWorkflowOrchestrator(logger)
//
//	// 创建服务
//	plannerService := planner.NewPlannerService(&service.PlannerServiceConfig{
//	    IntentAnalyzer:       intentAnalyzer,
//	    TaskDecomposer:       taskDecomposer,
//	    WorkflowOrchestrator: workflowOrchestrator,
//	})
//
//	// 执行规划
//	result, _ := plannerService.Plan(ctx, "给登录功能添加单元测试")
package planner

import (
	"github.com/weibaohui/nanobot-go/pkg/planner/agent"
	"github.com/weibaohui/nanobot-go/pkg/planner/api"
	"github.com/weibaohui/nanobot-go/pkg/planner/model"
	"github.com/weibaohui/nanobot-go/pkg/planner/orchestrator"
	"github.com/weibaohui/nanobot-go/pkg/planner/service"
)

// 重新导出常用类型，方便使用
type (
	// IntentAnalysis 意图分析结果
	IntentAnalysis = model.IntentAnalysis
	// TaskType 任务类型
	TaskType = model.TaskType
	// Complexity 复杂度
	Complexity = model.Complexity
	// Scope 范围
	Scope = model.Scope
	// SubTask 子任务
	SubTask = model.SubTask
	// SubTaskType 子任务类型
	SubTaskType = model.SubTaskType
	// Workflow 工作流
	Workflow = model.Workflow
	// WorkflowStage 工作流阶段
	WorkflowStage = model.WorkflowStage
	// WorkflowType 工作流类型
	WorkflowType = model.WorkflowType
	// PlanResult 规划结果
	PlanResult = service.PlanResult
)

// 重新导出构造函数
var (
	// NewIntentAnalyzer 创建意图分析器
	NewIntentAnalyzer = agent.NewIntentAnalyzer
	// NewTaskDecomposer 创建任务分解器
	NewTaskDecomposer = agent.NewTaskDecomposer
	// NewWorkflowOrchestrator 创建工作流编排器
	NewWorkflowOrchestrator = orchestrator.NewWorkflowOrchestrator
	// NewPlannerService 创建规划服务
	NewPlannerService = service.NewPlannerService
	// NewPlannerHandler 创建 API 处理器
	NewPlannerHandler = api.NewPlannerHandler
)

// 重新导出常量
const (
	// TaskTypeCode 编码任务
	TaskTypeCode = model.TaskTypeCode
	// TaskTypeDebug 调试任务
	TaskTypeDebug = model.TaskTypeDebug
	// TaskTypeRefactor 重构任务
	TaskTypeRefactor = model.TaskTypeRefactor
	// TaskTypeDocument 文档任务
	TaskTypeDocument = model.TaskTypeDocument
	// TaskTypeTest 测试任务
	TaskTypeTest = model.TaskTypeTest
	// TaskTypeOther 其他任务
	TaskTypeOther = model.TaskTypeOther

	// ComplexitySimple 简单
	ComplexitySimple = model.ComplexitySimple
	// ComplexityMedium 中等
	ComplexityMedium = model.ComplexityMedium
	// ComplexityComplex 复杂
	ComplexityComplex = model.ComplexityComplex

	// ScopeFile 文件级
	ScopeFile = model.ScopeFile
	// ScopePackage 包级
	ScopePackage = model.ScopePackage
	// ScopeProject 项目级
	ScopeProject = model.ScopeProject

	// SubTaskTypeAnalyze 分析
	SubTaskTypeAnalyze = model.SubTaskTypeAnalyze
	// SubTaskTypeCreate 创建
	SubTaskTypeCreate = model.SubTaskTypeCreate
	// SubTaskTypeModify 修改
	SubTaskTypeModify = model.SubTaskTypeModify
	// SubTaskTypeTest 测试
	SubTaskTypeTest = model.SubTaskTypeTest
	// SubTaskTypeVerify 验证
	SubTaskTypeVerify = model.SubTaskTypeVerify

	// WorkflowTypeSequential 顺序工作流
	WorkflowTypeSequential = model.WorkflowTypeSequential
	// WorkflowTypeParallel 并行工作流
	WorkflowTypeParallel = model.WorkflowTypeParallel
	// WorkflowTypeTask 任务工作流
	WorkflowTypeTask = model.WorkflowTypeTask
)
