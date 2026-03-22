# 025-Eino自动规划系统-设计

| 修改人 | 修改时间 | 修改内容 |
| ------ | -------- | -------- |
| AI Assistant | 2026-03-21 | 初始版本 - Phase 1 设计 |

## 1. 架构设计

### 1.1 整体架构

```
pkg/planner/
├── api/              # API 层
│   ├── handler.go    # HTTP 处理器
│   └── routes.go     # 路由注册
├── model/            # 数据模型
│   ├── intent.go     # 意图分析模型
│   ├── task.go       # 任务模型
│   └── workflow.go   # 工作流模型
├── agent/            # Agent 实现
│   ├── intent_analyzer.go  # 意图分析 Agent
│   └── task_decomposer.go  # 任务分解 Agent
├── orchestrator/     # 编排器
│   └── workflow.go   # 工作流编排器
└── service/          # 服务层
    └── planner.go    # 规划服务
```

### 1.2 层次关系

```
HTTP API Handlers
      ↓
Planner Service
      ↓
      ├→ IntentAnalyzer Agent
      │       ↓
      │   Eino ChatModelAgent
      │
      ├→ TaskDecomposer Agent
      │       ↓
      │   Eino DeepAgent
      │
      └→ Workflow Orchestrator
```

## 2. 核心数据结构设计

### 2.1 意图分析模型

```go
// file: pkg/planner/model/intent.go

package model

// TaskType 任务类型
type TaskType string

const (
    TaskTypeCode     TaskType = "code"     // 编码任务
    TaskTypeDebug    TaskType = "debug"    // 调试任务
    TaskTypeRefactor TaskType = "refactor" // 重构任务
    TaskTypeDocument TaskType = "document" // 文档任务
    TaskTypeTest     TaskType = "test"     // 测试任务
    TaskTypeOther    TaskType = "other"    // 其他任务
)

// Complexity 复杂度
type Complexity string

const (
    ComplexitySimple  Complexity = "simple"  // 简单
    ComplexityMedium  Complexity = "medium"  // 中等
    ComplexityComplex Complexity = "complex" // 复杂
)

// Scope 范围
type Scope string

const (
    ScopeFile    Scope = "file"    // 文件级
    ScopePackage Scope = "package" // 包级
    ScopeProject Scope = "project" // 项目级
)

// IntentAnalysis 意图分析结果
type IntentAnalysis struct {
    TaskType        TaskType   `json:"taskType"`        // 任务类型
    Complexity      Complexity `json:"complexity"`      // 复杂度
    Scope           Scope      `json:"scope"`           // 范围
    Technologies    []string   `json:"technologies"`    // 涉及的技术栈
    Dependencies    []string   `json:"dependencies"`    // 依赖项
    Constraints     []string   `json:"constraints"`     // 约束条件
    SuccessCriteria []string   `json:"successCriteria"` // 成功标准
    RawQuery        string     `json:"rawQuery"`        // 原始查询
}

// IsValid 验证意图分析结果是否有效
func (i *IntentAnalysis) IsValid() bool {
    if i.TaskType == "" {
        return false
    }
    if i.Complexity == "" {
        return false
    }
    if i.Scope == "" {
        return false
    }
    return true
}
```

### 2.2 任务模型

```go
// file: pkg/planner/model/task.go

package model

// TaskType 子任务类型
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
```

### 2.3 工作流模型

```go
// file: pkg/planner/model/workflow.go

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
    Type      WorkflowType `json:"type"`                // 阶段类型
    Task      *SubTask     `json:"task,omitempty"`      // 单个任务（type=task 时）
    Tasks     []*SubTask   `json:"tasks,omitempty"`     // 多个任务（type=parallel 时）
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
```

## 3. 核心组件设计

### 3.1 意图分析 Agent

```go
// file: pkg/planner/agent/intent_analyzer.go

package agent

import (
    "context"
    "fmt"

    "github.com/weibaohui/nanobot-go/pkg/agent/eino"
    "github.com/weibaohui/nanobot-go/pkg/planner/model"
    "go.uber.org/zap"
)

// IntentAnalyzer 意图分析器
type IntentAnalyzer struct {
    agent  *eino.ChatModelAgent
    logger *zap.Logger
}

// IntentAnalyzerConfig 配置
type IntentAnalyzerConfig struct {
    Model  eino.BaseChatModel
    Logger *zap.Logger
}

// NewIntentAnalyzer 创建意图分析器
func NewIntentAnalyzer(ctx context.Context, cfg *IntentAnalyzerConfig) (*IntentAnalyzer, error) {
    if cfg.Model == nil {
        return nil, fmt.Errorf("model is required")
    }

    logger := cfg.Logger
    if logger == nil {
        logger = zap.L()
    }

    instruction := `你是一个专业的任务意图分析专家。请分析用户的查询，提取以下信息：

1. taskType: 任务类型，可选值为 code（编码）、debug（调试）、refactor（重构）、document（文档）、test（测试）、other（其他）
2. complexity: 复杂度，可选值为 simple（简单）、medium（中等）、complex（复杂）
3. scope: 范围，可选值为 file（文件级）、package（包级）、project（项目级）
4. technologies: 涉及的技术栈，如 ["go", "react", "sql"]
5. successCriteria: 成功标准，列出 2-5 条关键验收条件

请以 JSON 格式返回结果。`

    agent, err := eino.NewChatModelAgent(ctx, &eino.ChatModelAgentConfig{
        Name:           "intent-analyzer",
        Description:    "意图分析 Agent",
        Instruction:    instruction,
        Model:          cfg.Model,
        MaxIterations:  1,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create chat model agent: %w", err)
    }

    return &IntentAnalyzer{
        agent:  agent,
        logger: logger,
    }, nil
}

// Analyze 分析用户查询
func (a *IntentAnalyzer) Analyze(ctx context.Context, query string) (*model.IntentAnalysis, error) {
    a.logger.Info("Analyzing intent", zap.String("query", query))

    // 创建 Runner
    runner := eino.NewRunner(ctx, eino.RunnerConfig{
        Agent:           a.agent,
        EnableStreaming: false,
    })

    // 运行查询
    iter := runner.Query(ctx, query)

    var output *eino.AgentOutput
    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        if event.Err != nil {
            return nil, fmt.Errorf("agent error: %w", event.Err)
        }
        if event.Output != nil {
            output = event.Output
        }
    }

    if output == nil {
        return nil, fmt.Errorf("no output from agent")
    }

    // 解析输出为 IntentAnalysis
    intent, err := a.parseOutput(output.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse output: %w", err)
    }

    intent.RawQuery = query

    a.logger.Info("Intent analyzed",
        zap.String("taskType", string(intent.TaskType)),
        zap.String("complexity", string(intent.Complexity)),
        zap.String("scope", string(intent.Scope)))

    return intent, nil
}

func (a *IntentAnalyzer) parseOutput(content string) (*model.IntentAnalysis, error) {
    // 简单实现，实际需要 JSON 解析和错误处理
    // 这里先返回一个示例结构
    return &model.IntentAnalysis{
        TaskType:     model.TaskTypeTest,
        Complexity:   model.ComplexityMedium,
        Scope:        model.ScopePackage,
        Technologies: []string{"go", "testing"},
        SuccessCriteria: []string{
            "覆盖主要功能逻辑",
            "包含正常和异常场景",
        },
    }, nil
}
```

### 3.2 任务分解 Agent

```go
// file: pkg/planner/agent/task_decomposer.go

package agent

import (
    "context"
    "fmt"

    "github.com/weibaohui/nanobot-go/pkg/agent/eino"
    "github.com/weibaohui/nanobot-go/pkg/planner/model"
    "go.uber.org/zap"
)

// TaskDecomposer 任务分解器
type TaskDecomposer struct {
    agent  eino.ResumableAgent
    logger *zap.Logger
}

// TaskDecomposerConfig 配置
type TaskDecomposerConfig struct {
    ChatModel eino.ChatModel
    Logger    *zap.Logger
}

// NewTaskDecomposer 创建任务分解器
func NewTaskDecomposer(ctx context.Context, cfg *TaskDecomposerConfig) (*TaskDecomposer, error) {
    if cfg.ChatModel == nil {
        return nil, fmt.Errorf("chat model is required")
    }

    logger := cfg.Logger
    if logger == nil {
        logger = zap.L()
    }

    instruction := `你是一个专业的任务分解专家。请将用户的查询分解为可执行的子任务列表。

每个子任务应包含：
- id: 唯一标识，如 "analyze", "create-test"
- name: 简短名称
- description: 详细描述
- type: 任务类型，可选 analyze（分析）、create（创建）、modify（修改）、test（测试）、verify（验证）
- agentType: Agent 类型，通常为 "chat" 或 "tool"
- dependsOn: 依赖的任务 ID 列表
- parallel: 是否可以与其他同层级任务并行执行

请以 JSON 格式返回 tasks 数组。`

    agent, err := eino.NewDeepAgent(ctx, &eino.DeepConfig{
        Name:         "task-decomposer",
        Description:  "任务分解 Agent",
        ChatModel:    cfg.ChatModel,
        Instruction:  instruction,
        MaxIteration: 5,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create deep agent: %w", err)
    }

    return &TaskDecomposer{
        agent:  agent,
        logger: logger,
    }, nil
}

// Decompose 分解任务
func (d *TaskDecomposer) Decompose(ctx context.Context, query string, intent *model.IntentAnalysis) ([]*model.SubTask, error) {
    d.logger.Info("Decomposing task",
        zap.String("query", query),
        zap.String("taskType", string(intent.TaskType)))

    // 构建提示词
    prompt := d.buildPrompt(query, intent)

    // 创建 Runner
    runner := eino.NewRunner(ctx, eino.RunnerConfig{
        Agent:           d.agent,
        EnableStreaming: false,
    })

    // 运行
    iter := runner.Query(ctx, prompt)

    var output *eino.AgentOutput
    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        if event.Err != nil {
            return nil, fmt.Errorf("agent error: %w", event.Err)
        }
        if event.Output != nil {
            output = event.Output
        }
    }

    if output == nil {
        return nil, fmt.Errorf("no output from agent")
    }

    // 解析输出
    tasks, err := d.parseOutput(output.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse output: %w", err)
    }

    // 验证任务
    taskList := &model.TaskList{Tasks: tasks}
    if err := taskList.ValidateAll(); err != nil {
        return nil, fmt.Errorf("invalid tasks: %w", err)
    }

    d.logger.Info("Task decomposed", zap.Int("taskCount", len(tasks)))

    return tasks, nil
}

func (d *TaskDecomposer) buildPrompt(query string, intent *model.IntentAnalysis) string {
    return fmt.Sprintf(`请分解以下任务：

查询：%s

已分析的意图：
- 任务类型：%s
- 复杂度：%s
- 范围：%s
- 技术栈：%v
- 成功标准：%v

请返回子任务列表。`,
        query,
        intent.TaskType,
        intent.Complexity,
        intent.Scope,
        intent.Technologies,
        intent.SuccessCriteria)
}

func (d *TaskDecomposer) parseOutput(content string) ([]*model.SubTask, error) {
    // 示例实现，返回预设的任务列表
    // 实际需要 JSON 解析
    return []*model.SubTask{
        {
            ID:          "analyze",
            Name:        "分析代码",
            Description: "分析现有代码结构",
            Type:        model.SubTaskTypeAnalyze,
            AgentType:   model.AgentTypeChat,
            DependsOn:   []string{},
            Parallel:    false,
        },
        {
            ID:          "create-test",
            Name:        "创建测试",
            Description: "创建测试文件",
            Type:        model.SubTaskTypeCreate,
            AgentType:   model.AgentTypeChat,
            DependsOn:   []string{"analyze"},
            Parallel:    false,
        },
        {
            ID:          "test-normal",
            Name:        "正常流程测试",
            Description: "测试正常流程",
            Type:        model.SubTaskTypeTest,
            AgentType:   model.AgentTypeTool,
            DependsOn:   []string{"create-test"},
            Parallel:    true,
        },
        {
            ID:          "test-error",
            Name:        "异常流程测试",
            Description: "测试异常流程",
            Type:        model.SubTaskTypeTest,
            AgentType:   model.AgentTypeTool,
            DependsOn:   []string{"create-test"},
            Parallel:    true,
        },
        {
            ID:          "verify",
            Name:        "验证结果",
            Description: "运行测试并验证",
            Type:        model.SubTaskTypeVerify,
            AgentType:   model.AgentTypeTool,
            DependsOn:   []string{"test-normal", "test-error"},
            Parallel:    false,
        },
    }, nil
}
```

### 3.3 工作流编排器

```go
// file: pkg/planner/orchestrator/workflow.go

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

    o.logger.Info("Workflow built", zap.Int("stageCount", len(workflow.Stages)))

    return workflow, nil
}

// buildDependencyGraph 构建依赖图
func (o *WorkflowOrchestrator) buildDependencyGraph(tasks []*model.SubTask) map[string][]string {
    // graph[taskID] = list of task IDs that depend on this task
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
```

### 3.4 规划服务

```go
// file: pkg/planner/service/planner.go

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
    intentAnalyzer    *agent.IntentAnalyzer
    taskDecomposer    *agent.TaskDecomposer
    workflowOrchestrator *orchestrator.WorkflowOrchestrator
    logger            *zap.Logger
}

// PlannerServiceConfig 配置
type PlannerServiceConfig struct {
    IntentAnalyzer    *agent.IntentAnalyzer
    TaskDecomposer    *agent.TaskDecomposer
    WorkflowOrchestrator *orchestrator.WorkflowOrchestrator
    Logger            *zap.Logger
}

// NewPlannerService 创建规划服务
func NewPlannerService(cfg *PlannerServiceConfig) *PlannerService {
    logger := cfg.Logger
    if logger == nil {
        logger = zap.L()
    }

    return &PlannerService{
        intentAnalyzer:     cfg.IntentAnalyzer,
        taskDecomposer:     cfg.TaskDecomposer,
        workflowOrchestrator: cfg.WorkflowOrchestrator,
        logger:             logger,
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
```

## 4. API 设计

### 4.1 HTTP Handler

```go
// file: pkg/planner/api/handler.go

package api

import (
    "encoding/json"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/weibaohui/nanobot-go/pkg/planner/model"
    "github.com/weibaohui/nanobot-go/pkg/planner/service"
    "go.uber.org/zap"
)

// PlannerHandler 规划处理器
type PlannerHandler struct {
    service *service.PlannerService
    logger  *zap.Logger
}

// NewPlannerHandler 创建规划处理器
func NewPlannerHandler(service *service.PlannerService, logger *zap.Logger) *PlannerHandler {
    if logger == nil {
        logger = zap.L()
    }
    return &PlannerHandler{
        service: service,
        logger:  logger,
    }
}

// AnalyzeIntentRequest 分析意图请求
type AnalyzeIntentRequest struct {
    Query string `json:"query" binding:"required"`
}

// AnalyzeIntentResponse 分析意图响应
type AnalyzeIntentResponse struct {
    Intent *model.IntentAnalysis `json:"intent"`
}

// HandleAnalyzeIntent 处理意图分析
func (h *PlannerHandler) HandleAnalyzeIntent(c *gin.Context) {
    var req AnalyzeIntentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    intent, err := h.service.AnalyzeIntent(c.Request.Context(), req.Query)
    if err != nil {
        h.logger.Error("Failed to analyze intent", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, AnalyzeIntentResponse{Intent: intent})
}

// DecomposeTasksRequest 分解任务请求
type DecomposeTasksRequest struct {
    Query  string                `json:"query" binding:"required"`
    Intent *model.IntentAnalysis `json:"intent" binding:"required"`
}

// DecomposeTasksResponse 分解任务响应
type DecomposeTasksResponse struct {
    Tasks []*model.SubTask `json:"tasks"`
}

// HandleDecomposeTasks 处理任务分解
func (h *PlannerHandler) HandleDecomposeTasks(c *gin.Context) {
    var req DecomposeTasksRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    tasks, err := h.service.DecomposeTasks(c.Request.Context(), req.Query, req.Intent)
    if err != nil {
        h.logger.Error("Failed to decompose tasks", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, DecomposeTasksResponse{Tasks: tasks})
}

// BuildWorkflowRequest 构建工作流请求
type BuildWorkflowRequest struct {
    Tasks []*model.SubTask `json:"tasks" binding:"required"`
}

// BuildWorkflowResponse 构建工作流响应
type BuildWorkflowResponse struct {
    Workflow *model.Workflow `json:"workflow"`
}

// HandleBuildWorkflow 处理构建工作流
func (h *PlannerHandler) HandleBuildWorkflow(c *gin.Context) {
    var req BuildWorkflowRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    workflow, err := h.service.BuildWorkflow(c.Request.Context(), req.Tasks)
    if err != nil {
        h.logger.Error("Failed to build workflow", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, BuildWorkflowResponse{Workflow: workflow})
}

// PlanRequest 完整规划请求
type PlanRequest struct {
    Query string `json:"query" binding:"required"`
}

// PlanResponse 完整规划响应
type PlanResponse struct {
    *service.PlanResult
}

// HandlePlan 处理完整规划
func (h *PlannerHandler) HandlePlan(c *gin.Context) {
    var req PlanRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := h.service.Plan(c.Request.Context(), req.Query)
    if err != nil {
        h.logger.Error("Failed to plan", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, PlanResponse{PlanResult: result})
}

// RegisterRoutes 注册路由
func (h *PlannerHandler) RegisterRoutes(rg *gin.RouterGroup) {
    planner := rg.Group("/planner")
    {
        planner.POST("/analyze", h.HandleAnalyzeIntent)
        planner.POST("/decompose", h.HandleDecomposeTasks)
        planner.POST("/workflow", h.HandleBuildWorkflow)
        planner.POST("/plan", h.HandlePlan)
    }
}
```

## 5. 集成方案

### 5.1 依赖注入

在 `internal/wire/wire.go` 中添加：

```go
// 提供 PlannerService
func providePlannerService(
    intentAnalyzer *agent.IntentAnalyzer,
    taskDecomposer *agent.TaskDecomposer,
    workflowOrchestrator *orchestrator.WorkflowOrchestrator,
    logger *zap.Logger,
) *service.PlannerService {
    return service.NewPlannerService(&service.PlannerServiceConfig{
        IntentAnalyzer:     intentAnalyzer,
        TaskDecomposer:     taskDecomposer,
        WorkflowOrchestrator: workflowOrchestrator,
        Logger:             logger,
    })
}

// 提供 IntentAnalyzer（需要先提供 Model）
func provideIntentAnalyzer(
    ctx context.Context,
    // 需要注入一个可用的 eino.BaseChatModel
    logger *zap.Logger,
) (*agent.IntentAnalyzer, error) {
    // 注意：这里需要先配置好 Model
    // 可以通过 LLMProvider 获取
    return agent.NewIntentAnalyzer(ctx, &agent.IntentAnalyzerConfig{
        Model:  someModel, // 需要实现
        Logger: logger,
    })
}
```

### 5.2 路由注册

在 `internal/api/server.go` 中添加：

```go
// 注册 planner 路由
plannerHandler := providePlannerHandler(...)
plannerHandler.RegisterRoutes(apiV1)
```

## 6. 测试策略

### 6.1 单元测试

- 测试数据模型验证逻辑
- 测试依赖图构建
- 测试拓扑排序
- 测试工作流编排

### 6.2 集成测试

- 测试完整的规划流程
- 测试 API 端点

## 7. 目录结构

```
pkg/planner/
├── api/
│   └── handler.go
├── model/
│   ├── intent.go
│   ├── task.go
│   └── workflow.go
├── agent/
│   ├── intent_analyzer.go
│   └── task_decomposer.go
├── orchestrator/
│   └── workflow.go
├── service/
│   └── planner.go
└── planner.go  # 包入口
```
