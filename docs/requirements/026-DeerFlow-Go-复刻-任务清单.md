# DeerFlow Go 复刻 - 任务清单

| 项目 | 内容 |
|------|------|
| 创建日期 | 2026-03-22 |
| 状态 | 待确认 |

---

## 任务总览

### 核心原则
1. **复用 Eino** - 尽可能用 Eino 现有能力，不重复造轮子
2. **提示词模块化** - 解决 DeerFlow 提示词难维护的痛点
3. **分阶段实施** - 每个阶段都有可交付物，可独立验证
4. **保持 API 兼容** - 概念与 DeerFlow 对齐，便于理解

---

## Phase 0: 准备工作（已完成 ✅）

- [x] 分析 DeerFlow 源码与架构
- [x] 分析 Eino 框架能力
- [x] 完成架构映射设计
- [x] 创建设计文档
- [x] 创建任务清单（本文件）

**交付物**：
- [docs/design/026-DeerFlow-Go-复刻方案-设计文档.md](../design/026-DeerFlow-Go-复刻方案-设计文档.md)
- 本任务清单

---

## Phase 1: 基础设施（优先级：P0）

### 1.1 创建目录结构

**任务**：搭建 `pkg/deerflow/` 目录骨架

```
pkg/deerflow/
├── README.md
├── go.mod
├── agent/
├── sandbox/
├── middleware/
├── subagents/
├── prompts/
├── skills/
├── memory/
├── config/
├── state/
└── tools/
```

**验收**：目录结构创建完成，空文件占位

---

### 1.2 提示词管理系统（核心！）

**目标**：解决 DeerFlow 提示词难处理的问题

**任务清单**：

- [ ] **1.2.1 提示词类型定义** (`prompts/types.go`)
  - `PromptSection` 接口
  - `PromptBuilder` 结构体
  - 模板变量占位符定义

- [ ] **1.2.2 提示词分段模块** (`prompts/sections/`)
  - `role.go` - `<role>` 部分
  - `thinking_style.go` - `<thinking_style>` 部分
  - `clarification.go` - `<clarification_system>` 部分
  - `subagent.go` - `<subagent_system>` 部分（支持 max_concurrent 变量）
  - `working_dir.go` - `<working_directory>` 部分
  - `citations.go` - `<citations>` 部分
  - `critical_reminders.go` - `<critical_reminders>` 部分

- [ ] **1.2.3 提示词构建器** (`prompts/builder.go`)
  ```go
  type PromptBuilder struct { ... }

  func NewPromptBuilder() *PromptBuilder
  func (b *PromptBuilder) WithRole(agentName string) *PromptBuilder
  func (b *PromptBuilder) WithThinkingStyle() *PromptBuilder
  func (b *PromptBuilder) WithClarification() *PromptBuilder
  func (b *PromptBuilder) WithSubagent(maxConcurrent int) *PromptBuilder
  func (b *PromptBuilder) WithWorkingDir() *PromptBuilder
  func (b *PromptBuilder) WithSkills(skills []Skill) *PromptBuilder
  func (b *PromptBuilder) WithMemory(memory MemoryData) *PromptBuilder
  func (b *PromptBuilder) Build() string
  ```

- [ ] **1.2.4 完整提示词模板** (`prompts/templates/`)
  - `lead_agent.go` - Lead Agent 完整提示词
  - `general_purpose.go` - 通用子 Agent 提示词
  - `bash.go` - Bash 专家提示词

- [ ] **1.2.5 单元测试**
  - 每个分段的渲染测试
  - PromptBuilder 集成测试
  - 变量替换测试

**验收标准**：
- [ ] 可以用 Builder 模式按需组合提示词
- [ ] 提示词分段可独立测试
- [ ] 支持动态变量（如 `max_concurrent`）
- [ ] 生成的提示词与 DeerFlow 效果一致

**预计工作量**：2-3 天

---

### 1.3 配置系统

**任务**：`pkg/deerflow/config/`

- [ ] `config.go` - 主配置结构
- [ ] `model.go` - 模型配置
- [ ] `sandbox.go` - Sandbox 配置
- [ ] `extensions.go` - 扩展配置（MCP, Skills）
- [ ] YAML 解析与环境变量替换

**验收**：可以从 YAML 加载配置，支持 `$OPENAI_API_KEY` 格式的环境变量

---

### 1.4 状态管理

**任务**：`pkg/deerflow/state/`

- [ ] `thread_state.go` - ThreadState 定义（参考 DeerFlow）
  - sandbox
  - thread_data (workspace_path, uploads_path, outputs_path)
  - title
  - artifacts
  - todos
  - uploaded_files
  - viewed_images
- [ ] `reducers.go` - Reducer 函数（用于合并状态）

**验收**：状态结构定义完整，reducer 函数可正常工作

---

## Phase 1 交付物检查清单

- [ ] pkg/deerflow/ 目录结构完整
- [ ] 提示词 Builder 可正常工作
- [ ] 配置系统可加载 YAML
- [ ] ThreadState 定义完成
- [ ] 所有模块有单元测试

---

## Phase 2: Sandbox 系统（优先级：P0）

### 2.1 Sandbox 抽象接口

**文件**：`pkg/deerflow/sandbox/interface.go`

```go
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

**任务**：
- [ ] 定义 Sandbox 接口
- [ ] 定义 SandboxProvider 接口
- [ ] 定义错误类型（SandboxError, SandboxNotFoundError 等）

---

### 2.2 路径翻译系统

**文件**：`pkg/deerflow/sandbox/path.go`

核心功能：
```
虚拟路径 → 物理路径
/mnt/user-data/workspace/* → .deer-flow/threads/{id}/user-data/workspace/*
/mnt/user-data/uploads/*   → .deer-flow/threads/{id}/user-data/uploads/*
/mnt/user-data/outputs/*   → .deer-flow/threads/{id}/user-data/outputs/*
/mnt/skills/*              → skills/*
```

**任务**：
- [ ] `ReplaceVirtualPath(path, threadData)` - 虚拟→物理
- [ ] `MaskLocalPathsInOutput(output, threadData)` - 物理→虚拟（用于输出）
- [ ] `ValidateLocalToolPath(path, threadData, readOnly)` - 安全校验
- [ ] `ReplaceVirtualPathsInCommand(command, threadData)` - 命令中的路径替换

---

### 2.3 LocalSandbox 实现

**文件**：`pkg/deerflow/sandbox/local/`

- [ ] `local_sandbox.go` - LocalSandbox 实现
- [ ] `provider.go` - LocalSandboxProvider 实现
- [ ] `list_dir.go` - 目录树生成（maxDepth=2）

---

### 2.4 Sandbox 工具

**文件**：`pkg/deerflow/sandbox/tools/`

包装成 Eino 工具：

- [ ] `bash.go` - `bash(description, command)` 工具
- [ ] `read_file.go` - `read_file(description, path, start_line?, end_line?)`
- [ ] `write_file.go` - `write_file(description, path, content, append?)`
- [ ] `ls.go` - `ls(description, path)`
- [ ] `str_replace.go` - `str_replace(description, path, old_str, new_str, replace_all?)`

**关键实现细节**：
- 路径翻译（虚拟 ↔ 物理）
- 安全校验（防止路径遍历）
- 本地路径掩码（输出中隐藏物理路径）

---

## Phase 2 交付物检查清单

- [ ] Sandbox 接口定义完整
- [ ] 路径翻译系统正常工作
- [ ] LocalSandbox 实现完成
- [ ] 5 个 Sandbox 工具可正常调用
- [ ] 安全机制到位（路径遍历检测）

---

## Phase 3: Lead Agent 核心（优先级：P0）

### 3.1 Lead Agent 封装

**文件**：`pkg/deerflow/agent/lead_agent.go`

基于 Eino DeepAgent 封装：

```go
type LeadAgentConfig struct {
    Name                  string
    ChatModel             model.BaseChatModel
    Instruction           string        // 可选，不填用默认
    SubAgents             []eino.Agent
    Tools                 []tool.BaseTool
    MaxIteration          int
    SandboxBackend        sandbox.Backend
    Shell                 sandbox.Shell
    SkillsEnabled         bool
    SubagentEnabled       bool
    MaxConcurrentSubagents int
}

func NewLeadAgent(ctx context.Context, cfg *LeadAgentConfig) (eino.ResumableAgent, error)
```

**任务**：
- [ ] 定义 LeadAgentConfig
- [ ] 用 PromptBuilder 构建系统提示词
- [ ] 集成 Sandbox 工具
- [ ] 集成中间件链
- [ ] 创建 Eino DeepAgent

---

### 3.2 子 Agent 定义

**文件**：`pkg/deerflow/agent/subagent/`

- [ ] `general_purpose.go` - 通用子 Agent（不含 task 工具，防止递归）
- [ ] `bash.go` - Bash 专家 Agent
- [ ] `registry.go` - 子 Agent 注册表

---

## Phase 3 交付物检查清单

- [ ] LeadAgent 可正常创建
- [ ] 提示词正确构建
- [ ] Sandbox 工具可被调用
- [ ] 子 Agent 可被注册和使用

---

## Phase 4: 子代理执行器（优先级：P1）

### 4.1 子代理执行器核心

**文件**：`pkg/deerflow/subagents/executor.go`

参考 DeerFlow 实现：
- 双线程池设计（scheduler + execution）
- 状态管理（PENDING → RUNNING → COMPLETED/FAILED/TIMED_OUT）
- 并发控制（MAX_CONCURRENT_SUBAGENTS = 3）
- 超时机制（15 分钟）

**任务**：
- [ ] 定义 SubagentStatus 枚举
- [ ] 定义 SubagentResult 结构体
- [ ] 实现 SubagentExecutor
- [ ] 实现后台任务存储与锁
- [ ] 实现 execute_async 方法
- [ ] 实现 get_result 方法
- [ ] 实现 cleanup 方法

---

### 4.2 task() 工具

**文件**：`pkg/deerflow/subagents/task_tool.go`

```go
// task(description, prompt, subagent_type, max_turns?)
```

**任务**：
- [ ] 定义 task 工具
- [ ] 调用 SubagentExecutor
- [ ] 轮询状态（每 5 秒）
- [ ] 发送 SSE 事件（task_started, task_running, task_completed 等）
- [ ] 返回最终结果

---

## Phase 4 交付物检查清单

- [ ] 子代理可被异步执行
- [ ] 状态正确流转
- [ ] 并发控制生效（最多 3 个）
- [ ] 超时机制工作
- [ ] task() 工具可被 Lead Agent 调用

---

## Phase 5: 中间件链（优先级：P1）

### 5.1 中间件接口与链

**文件**：`pkg/deerflow/middleware/`

- [ ] `interface.go` - Middleware 接口定义
- [ ] `chain.go` - MiddlewareChain 实现

---

### 5.2 具体中间件实现

| 中间件 | 优先级 | 说明 |
|--------|--------|------|
| ThreadDataMiddleware | P0 | 创建线程目录 |
| SandboxMiddleware | P0 | 获取 Sandbox |
| UploadsMiddleware | P1 | 注入上传文件 |
| TitleMiddleware | P1 | 自动生成标题 |
| MemoryMiddleware | P2 | 记忆提取队列 |
| ViewImageMiddleware | P2 | 图片注入（视觉模型） |
| ClarificationMiddleware | P2 | 澄清请求拦截 |

**任务**（按优先级）：
- [ ] ThreadDataMiddleware - 创建 workspace/uploads/outputs 目录
- [ ] SandboxMiddleware - 从 provider 获取 sandbox
- [ ] UploadsMiddleware（可选）
- [ ] TitleMiddleware（可选）
- [ ] 其他（后续）

---

## Phase 5 交付物检查清单

- [ ] 中间件接口定义
- [ ] MiddlewareChain 可正常工作
- [ ] ThreadDataMiddleware 实现完成
- [ ] SandboxMiddleware 实现完成

---

## Phase 6+: 其他功能（优先级：P2）

- [ ] Skills 系统
- [ ] Memory 系统
- [ ] MCP 集成
- [ ] IM 渠道（Feishu/Slack/Telegram）
- [ ] Gateway API（如果需要）

---

## 实施建议顺序

### 第一周（MVP）
1. Phase 1 - 基础设施（提示词系统是重点）
2. Phase 2 - Sandbox 系统
3. Phase 3 - Lead Agent 核心

**此时可以跑通：Lead Agent → Sandbox 工具**

### 第二周
4. Phase 4 - 子代理执行器
5. Phase 5 - 中间件链（核心部分）

**此时可以跑通：Lead Agent → task() → 子 Agent**

### 第三周及以后
6. 优化和完善
7. 补充测试
8. 文档

---

## 风险提示

### 高风险项
1. **提示词效果** - 模块化后需要验证效果是否与原 DeerFlow 一致
2. **Eino DeepAgent 能力** - 如果 DeepAgent 不够灵活，可能需要改用 ChatModelAgent 自己实现

### 应对措施
- 每个阶段都做验证测试
- 保留切换到备用方案的可能性

---

## 验收标准（最终）

- [ ] Lead Agent 可以进行正常对话
- [ ] 可以调用 Sandbox 工具（bash, read_file, write_file, ls）
- [ ] 可以委托任务给子 Agent 并监督执行
- [ ] 提示词模块化，可独立维护和测试
- [ ] 核心模块单元测试覆盖率 ≥ 70%
- [ ] 完整的使用文档和示例

---

## 相关文档

- [设计文档](../design/026-DeerFlow-Go-复刻方案-设计文档.md)
- [DeerFlow 后端架构](../../deer-flow/backend/README.md)
- [Eino 文档](https://github.com/cloudwego/eino)
