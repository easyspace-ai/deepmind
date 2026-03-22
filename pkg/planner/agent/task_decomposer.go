// Package agent 提供规划系统的 Agent 实现
package agent

import (
	"context"
	"fmt"

	"github.com/weibaohui/nanobot-go/pkg/planner/model"
	"go.uber.org/zap"
)

// TaskDecomposer 任务分解器
type TaskDecomposer struct {
	logger    *zap.Logger
	chatModel interface{} // 预留字段，用于未来 LLM 集成
	useLLM    bool
}

// TaskDecomposerConfig 配置
type TaskDecomposerConfig struct {
	Logger    *zap.Logger
	ChatModel interface{} // 可选，预留用于未来 LLM 集成
}

// NewTaskDecomposer 创建任务分解器
func NewTaskDecomposer(_ context.Context, cfg *TaskDecomposerConfig) (*TaskDecomposer, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = zap.L()
	}

	return &TaskDecomposer{
		logger:    logger,
		chatModel: cfg.ChatModel,
		useLLM:    false, // 暂时禁用 LLM 模式，待后续完善
	}, nil
}

// Decompose 分解任务
func (d *TaskDecomposer) Decompose(ctx context.Context, query string, intent *model.IntentAnalysis) ([]*model.SubTask, error) {
	d.logger.Info("Decomposing task",
		zap.String("query", query),
		zap.String("taskType", string(intent.TaskType)))

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if intent == nil {
		return nil, fmt.Errorf("intent is required")
	}
	if !intent.IsValid() {
		return nil, fmt.Errorf("intent is invalid")
	}

	tasks := d.generateTasksByType(query, intent)

	// 验证任务
	taskList := &model.TaskList{Tasks: tasks}
	if err := taskList.ValidateAll(); err != nil {
		return nil, fmt.Errorf("invalid tasks: %w", err)
	}

	d.logger.Info("Task decomposed", zap.Int("taskCount", len(tasks)))

	return tasks, nil
}

// generateTasksByType 根据任务类型生成子任务
func (d *TaskDecomposer) generateTasksByType(query string, intent *model.IntentAnalysis) []*model.SubTask {
	switch intent.TaskType {
	case model.TaskTypeTest:
		return d.generateTestTasks(query)
	case model.TaskTypeCode:
		return d.generateCodeTasks(query)
	case model.TaskTypeDocument:
		return d.generateDocumentTasks(query)
	case model.TaskTypeRefactor:
		return d.generateRefactorTasks(query)
	case model.TaskTypeDebug:
		return d.generateDebugTasks(query)
	default:
		return d.generateDefaultTasks(query)
	}
}

func (d *TaskDecomposer) generateTestTasks(query string) []*model.SubTask {
	return []*model.SubTask{
		{
			ID:          "analyze",
			Name:        "分析代码",
			Description: "分析现有代码结构和功能",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{},
			Parallel:    false,
		},
		{
			ID:          "create-test",
			Name:        "创建测试文件",
			Description: "创建测试文件并设置基本结构",
			Type:        model.SubTaskTypeCreate,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"analyze"},
			Parallel:    false,
		},
		{
			ID:          "test-normal",
			Name:        "正常流程测试",
			Description: "编写正常流程的测试用例",
			Type:        model.SubTaskTypeTest,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{"create-test"},
			Parallel:    true,
		},
		{
			ID:          "test-error",
			Name:        "异常流程测试",
			Description: "编写异常和边界场景的测试用例",
			Type:        model.SubTaskTypeTest,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{"create-test"},
			Parallel:    true,
		},
		{
			ID:          "verify",
			Name:        "验证测试",
			Description: "运行测试并验证结果",
			Type:        model.SubTaskTypeVerify,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{"test-normal", "test-error"},
			Parallel:    false,
		},
	}
}

func (d *TaskDecomposer) generateCodeTasks(query string) []*model.SubTask {
	return []*model.SubTask{
		{
			ID:          "analyze-req",
			Name:        "需求分析",
			Description: "分析需求并确定实现方案",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{},
			Parallel:    false,
		},
		{
			ID:          "design",
			Name:        "设计",
			Description: "设计接口和数据结构",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"analyze-req"},
			Parallel:    false,
		},
		{
			ID:          "implement",
			Name:        "实现",
			Description: "编写核心功能代码",
			Type:        model.SubTaskTypeCreate,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"design"},
			Parallel:    false,
		},
		{
			ID:          "test",
			Name:        "测试",
			Description: "编写测试用例",
			Type:        model.SubTaskTypeTest,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{"implement"},
			Parallel:    false,
		},
		{
			ID:          "review",
			Name:        "代码审查",
			Description: "检查代码质量",
			Type:        model.SubTaskTypeVerify,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"test"},
			Parallel:    false,
		},
	}
}

func (d *TaskDecomposer) generateDocumentTasks(query string) []*model.SubTask {
	return []*model.SubTask{
		{
			ID:          "analyze-code",
			Name:        "分析代码",
			Description: "分析需要文档化的代码",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{},
			Parallel:    false,
		},
		{
			ID:          "outline",
			Name:        "编写大纲",
			Description: "编写文档大纲",
			Type:        model.SubTaskTypeCreate,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"analyze-code"},
			Parallel:    false,
		},
		{
			ID:          "write-doc",
			Name:        "编写文档",
			Description: "编写完整文档内容",
			Type:        model.SubTaskTypeCreate,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"outline"},
			Parallel:    false,
		},
		{
			ID:          "add-examples",
			Name:        "添加示例",
			Description: "添加使用示例",
			Type:        model.SubTaskTypeModify,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"write-doc"},
			Parallel:    false,
		},
	}
}

func (d *TaskDecomposer) generateRefactorTasks(query string) []*model.SubTask {
	return []*model.SubTask{
		{
			ID:          "analyze-current",
			Name:        "分析现状",
			Description: "分析当前代码结构和问题",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{},
			Parallel:    false,
		},
		{
			ID:          "plan-refactor",
			Name:        "规划重构",
			Description: "制定重构计划",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"analyze-current"},
			Parallel:    false,
		},
		{
			ID:          "backup",
			Name:        "备份",
			Description: "备份当前代码",
			Type:        model.SubTaskTypeModify,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{"plan-refactor"},
			Parallel:    false,
		},
		{
			ID:          "apply-refactor",
			Name:        "执行重构",
			Description: "应用重构变更",
			Type:        model.SubTaskTypeModify,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"backup"},
			Parallel:    false,
		},
		{
			ID:          "verify-refactor",
			Name:        "验证",
			Description: "运行测试验证重构",
			Type:        model.SubTaskTypeVerify,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{"apply-refactor"},
			Parallel:    false,
		},
	}
}

func (d *TaskDecomposer) generateDebugTasks(query string) []*model.SubTask {
	return []*model.SubTask{
		{
			ID:          "reproduce",
			Name:        "复现问题",
			Description: "复现并确认问题",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{},
			Parallel:    false,
		},
		{
			ID:          "diagnose",
			Name:        "诊断问题",
			Description: "定位问题根源",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"reproduce"},
			Parallel:    false,
		},
		{
			ID:          "fix",
			Name:        "修复问题",
			Description: "实现修复方案",
			Type:        model.SubTaskTypeModify,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"diagnose"},
			Parallel:    false,
		},
		{
			ID:          "test-fix",
			Name:        "测试修复",
			Description: "验证修复是否有效",
			Type:        model.SubTaskTypeTest,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{"fix"},
			Parallel:    false,
		},
		{
			ID:          "add-regression",
			Name:        "添加回归测试",
			Description: "添加防止回归的测试用例",
			Type:        model.SubTaskTypeCreate,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"test-fix"},
			Parallel:    false,
		},
	}
}

func (d *TaskDecomposer) generateDefaultTasks(query string) []*model.SubTask {
	return []*model.SubTask{
		{
			ID:          "understand",
			Name:        "理解需求",
			Description: "理解用户需求",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{},
			Parallel:    false,
		},
		{
			ID:          "plan",
			Name:        "制定计划",
			Description: "制定执行计划",
			Type:        model.SubTaskTypeAnalyze,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"understand"},
			Parallel:    false,
		},
		{
			ID:          "execute",
			Name:        "执行",
			Description: "执行计划",
			Type:        model.SubTaskTypeModify,
			AgentType:   model.AgentTypeChat,
			DependsOn:   []string{"plan"},
			Parallel:    false,
		},
		{
			ID:          "verify",
			Name:        "验证",
			Description: "验证结果",
			Type:        model.SubTaskTypeVerify,
			AgentType:   model.AgentTypeTool,
			DependsOn:   []string{"execute"},
			Parallel:    false,
		},
	}
}
