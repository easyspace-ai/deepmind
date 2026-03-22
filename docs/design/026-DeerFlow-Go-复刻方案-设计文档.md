# DeerFlow Go 复刻方案 - 设计文档

| 项目 | 内容 |
|------|------|
| 版本 | v1.0 |
| 作者 | AI Assistant |
| 创建日期 | 2026-03-22 |
| 状态 | 设计中 |

---

## 1. 背景与目标

### 1.1 背景
- DeerFlow 是字节跳动开源的 Python 超级 Agent 系统
- 我们的项目 nanobot-go 已有 Eino 框架集成
- 需要用 Go 复刻 DeerFlow 的核心能力

### 1.2 目标
- 复刻 DeerFlow 的核心架构和功能
- 复用 Eino 框架作为底层编排引擎
- 解决提示词管理的复杂性问题
- 提供与 DeerFlow 兼容的 API 接口

### 1.3 非目标
- 不 1:1 复刻每一行代码
- 不完全兼容 DeerFlow 的 Python API
- 不实现 LangGraph（已用 Eino 替代）

---

## 2. 架构映射

### 2.1 核心组件映射表

| DeerFlow (Python) | nanobot-go (Go) | 说明 |
|-------------------|------------------|------|
| `make_lead_agent()` | `deerflow.NewLeadAgent()` | 基于 Eino DeepAgent |
| `LangGraph` | `eino/compose.Graph` | 图编排 |
| `Middleware Chain` | `eino.AgentMiddleware` | 中间件系统 |
| `SubagentExecutor` | `deerflow.SubagentExecutor` | 子代理执行器 |
| `Sandbox` | `deerflow.Sandbox` | Sandbox 抽象接口 |
| `Memory System` | `eino.CheckPointStore` + Session | 记忆与状态 |
| `Skills System` | `deerflow.Skills` | Skills 加载 |

### 2.2 总体架构

```
┌─────────────────────────────────────────────────────────────┐
│                   nanobot-go API Layer                      │
└────────────────────────────┬────────────────────────────────┘
                             │
┌─────────────────────────────────────────────────────────────┐
│                    pkg/deerflow/                            │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Lead Agent (基于 Eino DeepAgent)                     │ │
│  │  - 意图分析与任务拆解                                  │ │
│  │  - 子代理委托与监督                                    │ │
│  │  - 结果汇总与生成                                      │ │
│  └───────────────────────────────────────────────────────┘ │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  Middleware  │  │   Sandbox    │  │   Tools      │   │
│  │    Chain     │  │   System     │  │   System     │   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Prompts Management (模块化提示词)                    │ │
│  └───────────────────────────────────────────────────────┘ │
└────────────────────────────┬────────────────────────────────┘
                             │
┌─────────────────────────────────────────────────────────────┐
│                   Eino Framework Layer                      │
│  ┌──────────────────┐  ┌──────────────────────────────┐  │
│  │  DeepAgent       │  │  compose.Graph               │  │
│  │  ChatModelAgent  │  │  Agent Middlewares          │  │
│  │  PlanExecute     │  │  CheckPoint Store           │  │
│  └──────────────────┘  └──────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. 提示词管理方案（关键！）

### 3.1 问题分析

DeerFlow 的提示词问题：
- ❌ 单一大文件，难以维护
- ❌ Python 格式的字符串模板，Go 中不好处理
- ❌ 动态变量嵌入（`{variable}`）与 XML 标签混杂
- ❌ 难以测试和版本控制

### 3.2 解决方案：模块化提示词系统

```
pkg/deerflow/prompts/
├── prompts.go              # 提示词管理器
├── types.go                # 提示词类型定义
├── loader.go               # 提示词加载器
│
├── sections/               # 提示词分段（可独立维护）
│   ├── role.go             # <role> 部分
│   ├── thinking_style.go   # <thinking_style> 部分
│   ├── clarification.go     # <clarification_system> 部分
│   ├── subagent.go         # <subagent_system> 部分
│   ├── working_dir.go      # <working_directory> 部分
│   ├── citations.go        # <citations> 部分
│   └── critical_reminders.go
│
└── templates/              # 完整提示词模板
    ├── lead_agent.go       # Lead Agent 完整提示词
    ├── general_purpose.go  # 通用子 Agent 提示词
    └── bash.go             # Bash 专家提示词
```

### 3.3 提示词构建器模式

```go
// pkg/deerflow/prompts/builder.go

type PromptBuilder struct {
    sections map[string]string
    variables map[string]interface{}
}

func NewPromptBuilder() *PromptBuilder {
    return &PromptBuilder{
        sections: make(map[string]string),
        variables: make(map[string]interface{}),
    }
}

func (b *PromptBuilder) WithRole(name string) *PromptBuilder
func (b *PromptBuilder) WithThinkingStyle() *PromptBuilder
func (b *PromptBuilder) WithClarification() *PromptBuilder
func (b *PromptBuilder) WithSubagent(maxConcurrent int) *PromptBuilder
func (b *PromptBuilder) WithSkills(skills []Skill) *PromptBuilder
func (b *PromptBuilder) WithMemory(memory MemoryData) *PromptBuilder

func (b *PromptBuilder) Build() string
```

### 3.4 提示词分段设计

每个提示词分段都是独立的 Go 常量或函数：

```go
// pkg/deerflow/prompts/sections/role.go

const RoleSection = `
<role>
You are {{.AgentName}}, an open-source super agent.
</role>
`

func RenderRoleSection(agentName string) string {
    return renderTemplate(RoleSection, map[string]interface{}{
        "AgentName": agentName,
    })
}
```

---

## 4. 目录结构设计

```
pkg/deerflow/
├── README.md                   # 使用文档
├── go.mod
│
├── agent/                      # Agent 相关
│   ├── lead_agent.go          # Lead Agent 实现
│   ├── lead_agent_prompt.go   # Lead Agent 提示词构建
│   ├── subagent/              # 子 Agent
│   │   ├── general_purpose.go
│   │   └── bash.go
│   └── registry.go             # Agent 注册表
│
├── sandbox/                    # Sandbox 系统
│   ├── interface.go            # Sandbox 抽象接口
│   ├── local/                  # 本地实现
│   │   ├── provider.go
│   │   └── sandbox.go
│   ├── docker/                 # Docker 实现（后续）
│   └── tools/                 # Sandbox 工具
│       ├── bash.go
│       ├── read_file.go
│       ├── write_file.go
│       └── ls.go
│
├── middleware/                 # 中间件系统
│   ├── interface.go            # Middleware 接口
│   ├── chain.go                # 中间件链
│   ├── thread_data.go          # 线程数据中间件
│   ├── sandbox.go              # Sandbox 中间件
│   ├── title.go                # 标题生成中间件
│   ├── memory.go               # 记忆中间件
│   └── uploads.go              # 上传文件中间件
│
├── subagents/                  # 子代理执行器
│   ├── executor.go             # 执行器核心
│   ├── registry.go             # 子 Agent 注册表
│   ├── task_tool.go            # task() 工具
│   └── types.go                # 状态类型
│
├── prompts/                    # 提示词系统（见上一节）
│   ├── builder.go
│   ├── loader.go
│   ├── sections/
│   └── templates/
│
├── skills/                     # Skills 系统
│   ├── loader.go               # Skills 加载
│   ├── parser.go               # SKILL.md 解析
│   └── types.go
│
├── memory/                     # 记忆系统
│   ├── interface.go
│   ├── store.go                # JSON 文件存储
│   ├── updater.go              # LLM 记忆更新
│   └── types.go
│
├── config/                     # 配置系统
│   ├── config.go               # 主配置
│   ├── model.go                # 模型配置
│   ├── sandbox.go              # Sandbox 配置
│   └── extensions.go           # 扩展配置（MCP, Skills）
│
├── state/                      # 状态管理
│   ├── thread_state.go         # 线程状态
│   └── reducers.go             # Reducer 函数
│
└── tools/                      # 工具系统
    ├── registry.go             # 工具注册表
    ├── builtins/               # 内置工具
    │   ├── present_files.go
    │   ├── ask_clarification.go
    │   └── view_image.go
    └── wrapper.go              # Eino 工具包装
```

---

## 5. 核心模块详细设计

### 5.1 Lead Agent

基于 Eino DeepAgent 封装：

```go
// pkg/deerflow/agent/lead_agent.go

type LeadAgentConfig struct {
    Name              string
    ChatModel         model.BaseChatModel
    Instruction       string
    SubAgents         []eino.Agent
    Tools             []tool.BaseTool
    MaxIteration      int
    SandboxBackend    sandbox.Backend
    Shell             sandbox.Shell
    SkillsEnabled     bool
    SubagentEnabled   bool
    MaxConcurrentSubagents int
}

func NewLeadAgent(ctx context.Context, cfg *LeadAgentConfig) (eino.ResumableAgent, error) {
    // 1. 构建提示词
    prompt := prompts.NewPromptBuilder().
        WithRole(cfg.Name).
        WithThinkingStyle().
        WithClarification().
        WithSubagent(cfg.MaxConcurrentSubagents).
        WithWorkingDirectory().
        Build()

    // 2. 准备中间件
    middlewares := buildMiddlewareChain(cfg)

    // 3. 创建 Eino DeepAgent
    return eino.NewDeepAgent(ctx, &eino.DeepConfig{
        Name:          cfg.Name,
        ChatModel:     cfg.ChatModel,
        Instruction:   prompt,
        SubAgents:     cfg.SubAgents,
        ToolsConfig:   eino.ToolsConfig{...},
        MaxIteration:  cfg.MaxIteration,
        Middlewares:   middlewares,
    })
}
```

### 5.2 Sandbox 系统

抽象接口设计（参考 DeerFlow）：

```go
// pkg/deerflow/sandbox/interface.go

type Sandbox interface {
    ID() string
    ExecuteCommand(command string) (string, error)
    ReadFile(path string) (string, error)
    WriteFile(path string, content string, append bool) error
    ListDir(path string, maxDepth int) ([]string, error)
    UpdateFile(path string, content []byte) error
}

type SandboxProvider interface {
    Acquire(threadID string) (string, error)
    Get(sandboxID string) (Sandbox, error)
    Release(sandboxID string) error
}
```

### 5.3 子代理执行器

```go
// pkg/deerflow/subagents/executor.go

type SubagentStatus string

const (
    SubagentStatusPending   SubagentStatus = "pending"
    SubagentStatusRunning   SubagentStatus = "running"
    SubagentStatusCompleted SubagentStatus = "completed"
    SubagentStatusFailed    SubagentStatus = "failed"
    SubagentStatusTimedOut  SubagentStatus = "timed_out"
)

type SubagentResult struct {
    TaskID      string
    TraceID     string
    Status      SubagentStatus
    Result      string
    Error       string
    StartedAt   time.Time
    CompletedAt time.Time
    AIMessages  []map[string]interface{}
}

type SubagentExecutor struct {
    config      SubagentConfig
    tools       []tool.BaseTool
    parentModel string
    // ...
}

func (e *SubagentExecutor) ExecuteAsync(task string, taskID string) (string, error)
func (e *SubagentExecutor) GetResult(taskID string) (*SubagentResult, error)
```

---

## 6. 实施计划

### Phase 1: 基础设施（优先级：最高）
- [ ] 目录结构搭建
- [ ] 提示词管理系统（模块化）
- [ ] 配置系统
- [ ] 状态管理（ThreadState）

### Phase 2: 核心 Agent（优先级：高）
- [ ] Lead Agent 实现（基于 Eino DeepAgent）
- [ ] 提示词构建器集成
- [ ] 基础工具集成

### Phase 3: Sandbox 系统（优先级：高）
- [ ] Sandbox 抽象接口
- [ ] LocalSandbox 实现
- [ ] 路径翻译系统
- [ ] Sandbox 工具（bash, read_file, write_file, ls）

### Phase 4: 子代理系统（优先级：中）
- [ ] 子代理执行器
- [ ] task() 工具
- [ ] 并发控制
- [ ] 状态轮询与 SSE 事件

### Phase 5: 中间件链（优先级：中）
- [ ] 中间件接口
- [ ] ThreadDataMiddleware
- [ ] SandboxMiddleware
- [ ] TitleMiddleware
- [ ] MemoryMiddleware

### Phase 6: 其他功能（优先级：低）
- [ ] Skills 系统
- [ ] Memory 系统
- [ ] MCP 集成

---

## 7. 关键技术决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| **编排引擎** | Eino | 已集成，Go 原生，字节跳动出品 |
| **提示词方案** | 模块化 Builder 模式 | 解决 DeerFlow 提示词难维护问题 |
| **Sandbox 抽象** | 参考 DeerFlow 接口 | 保持概念一致，便于后续扩展 Docker |
| **配置格式** | YAML | 与现有项目一致 |
| **状态持久化** | Eino CheckPoint | 复用现有能力 |

---

## 8. 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| Eino DeepAgent 能力不足 | 高 | 中 | 备用方案：直接用 ChatModelAgent + 自定义逻辑 |
| 提示词模板复杂度过高 | 中 | 高 | 分阶段实施，先做最小可用版本 |
| 与现有系统集成问题 | 中 | 低 | 充分调研现有 pkg/agent/ 架构 |

---

## 9. 验收标准

- [ ] Lead Agent 能正常对话
- [ ] 能调用 Sandbox 工具（bash, read_file, write_file）
- [ ] 能委托任务给子 Agent 并监督执行
- [ ] 提示词模块化可独立维护和测试
- [ ] 完整的单元测试覆盖核心模块

---

## 附录

### A. 参考资料
- [DeerFlow GitHub](https://github.com/bytedance/deer-flow)
- [Eino GitHub](https://github.com/cloudwego/eino)
- [DeerFlow 后端架构](../backend/README.md)

### B. 相关文档
- [Eino 自动规划系统 - 需求文档](../requirements/025-Eino自动规划系统-需求.md)
- [Eino 自动规划系统 - 设计文档](../design/025-Eino自动规划系统-设计文档.md)
