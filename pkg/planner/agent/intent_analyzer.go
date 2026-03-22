// Package agent 提供规划系统的 Agent 实现
package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/weibaohui/nanobot-go/pkg/planner/model"
	"go.uber.org/zap"
)

// IntentAnalyzer 意图分析器
type IntentAnalyzer struct {
	logger    *zap.Logger
	chatModel interface{} // 预留字段，用于未来 LLM 集成
	useLLM    bool
}

// IntentAnalyzerConfig 配置
type IntentAnalyzerConfig struct {
	Logger    *zap.Logger
	ChatModel interface{} // 可选，预留用于未来 LLM 集成
}

// NewIntentAnalyzer 创建意图分析器
func NewIntentAnalyzer(_ context.Context, cfg *IntentAnalyzerConfig) (*IntentAnalyzer, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = zap.L()
	}

	return &IntentAnalyzer{
		logger:    logger,
		chatModel: cfg.ChatModel,
		useLLM:    false, // 暂时禁用 LLM 模式，待后续完善
	}, nil
}

// Analyze 分析用户查询
func (a *IntentAnalyzer) Analyze(ctx context.Context, query string) (*model.IntentAnalysis, error) {
	a.logger.Info("Analyzing intent",
		zap.String("query", query),
		zap.Bool("useLLM", a.useLLM))

	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// 始终使用关键词模式（LLM 模式待后续完善）
	intent := a.analyzeByKeywords(query)
	intent.RawQuery = query

	a.logger.Info("Intent analyzed",
		zap.String("taskType", string(intent.TaskType)),
		zap.String("complexity", string(intent.Complexity)),
		zap.String("scope", string(intent.Scope)))

	return intent, nil
}

// analyzeByKeywords 基于关键词分析意图
func (a *IntentAnalyzer) analyzeByKeywords(query string) *model.IntentAnalysis {
	q := strings.ToLower(query)

	intent := &model.IntentAnalysis{
		TaskType:     model.TaskTypeOther,
		Complexity:   model.ComplexityMedium,
		Scope:        model.ScopeFile,
		Technologies: []string{},
		SuccessCriteria: []string{
			"完成用户请求的任务",
			"输出符合预期格式",
		},
	}

	// 检测任务类型
	switch {
	case containsAny(q, []string{"测试", "test", "单元测试", "unittest"}):
		intent.TaskType = model.TaskTypeTest
		intent.Technologies = append(intent.Technologies, "go", "testing")
		intent.SuccessCriteria = []string{
			"覆盖主要功能逻辑",
			"包含正常和异常场景",
			"测试通过率 100%",
		}
	case containsAny(q, []string{"文档", "document", "注释", "comment", "readme"}):
		intent.TaskType = model.TaskTypeDocument
		intent.Technologies = append(intent.Technologies, "markdown")
		intent.SuccessCriteria = []string{
			"文档内容清晰完整",
			"包含使用示例",
		}
	case containsAny(q, []string{"重构", "refactor", "优化", "optimize", "清理", "cleanup"}):
		intent.TaskType = model.TaskTypeRefactor
		intent.Technologies = append(intent.Technologies, "go")
		intent.SuccessCriteria = []string{
			"代码结构更清晰",
			"不破坏现有功能",
		}
	case containsAny(q, []string{"调试", "debug", "修复", "fix", "bug", "错误"}):
		intent.TaskType = model.TaskTypeDebug
		intent.Technologies = append(intent.Technologies, "go")
		intent.SuccessCriteria = []string{
			"问题定位准确",
			"修复方案验证通过",
		}
	case containsAny(q, []string{"创建", "create", "实现", "implement", "开发", "develop", "添加", "add", "新增"}):
		intent.TaskType = model.TaskTypeCode
		intent.Technologies = append(intent.Technologies, "go")
		intent.SuccessCriteria = []string{
			"功能实现完整",
			"代码符合规范",
			"包含必要的测试",
		}
	}

	// 检测复杂度
	switch {
	case containsAny(q, []string{"简单", "simple", "quick", "快速", "小", "small"}):
		intent.Complexity = model.ComplexitySimple
	case containsAny(q, []string{"复杂", "complex", "大量", "many", "整个", "whole", "完整", "complete"}):
		intent.Complexity = model.ComplexityComplex
	}

	// 检测范围
	switch {
	case containsAny(q, []string{"项目", "project", "整个", "whole", "所有", "all"}):
		intent.Scope = model.ScopeProject
	case containsAny(q, []string{"包", "package", "目录", "directory", "文件夹"}):
		intent.Scope = model.ScopePackage
	case containsAny(q, []string{"文件", "file", "这个", "this", "该"}):
		intent.Scope = model.ScopeFile
	}

	return intent
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
