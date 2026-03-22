# DeerFlow 深度集成 - 差距分析文档

## 概述

本文档对比原版 DeerFlow 系统与现有 nanobot-go 系统，识别需要集成的核心特性与差距。

## 一、完整数据流转路径对比

### 1.1 DeerFlow 完整数据流转（LangGraph）

```
用户输入
    ↓
[中间件链执行] (严格顺序！)
    ├─ ThreadDataMiddleware        - 创建线程目录
    ├─ UploadsMiddleware            - 处理上传文件
    ├─ DanglingToolCallMiddleware   - 处理挂起工具调用
    ├─ SummarizationMiddleware      - 上下文摘要（可选）
    ├─ TodoListMiddleware           - 任务追踪（可选）
    ├─ TitleMiddleware              - 自动生成标题
    ├─ MemoryMiddleware             - 记忆提取
    ├─ ViewImageMiddleware          - 图像查看
    ├─ SubagentLimitMiddleware      - 限制并发子代理（可选）
    ├─ LoopDetectionMiddleware      - 循环检测
    └─ ClarificationMiddleware      - 询问澄清（必须最后！）
    ↓
[Lead Agent 执行]
    ├─ 构建系统提示词（模块化）
    ├─ 加载可用工具
    │   ├─ Sandbox 工具（bash, ls, read_file, write_file, str_replace）
    │   ├─ Built-in 工具（present_files, ask_clarification, view_image）
    │   ├─ MCP 工具
    │   ├─ Community 工具（tavily, jina_ai, firecrawl, image_search）
    │   └─ Subagent 工具（task）
    └─ Agent 执行循环
        ├─ 思考 → 调用工具 → 获取结果
        └─ 重复直到完成
    ↓
[工具调用路径]
    └─ Sandbox 工具
        ├─ 虚拟路径验证（/mnt/user-data/*）
        ├─ 路径翻译（虚拟 → 物理）
        ├─ Sandbox 执行
        ├─ 结果路径掩码（物理 → 虚拟）
        └─ 返回结果
    ↓
[子代理 task() 工具路径]
    ├─ 获取子代理配置
    ├─ 继承父 Agent 的 sandbox_state 和 thread_data
    ├─ 过滤工具（移除 task 工具防止嵌套）
    ├─ SubagentExecutor.execute_async()
    │   ├─ 创建后台任务
    │   ├─ 双线程池：_scheduler_pool (3) + _execution_pool (3)
    │   ├─ 发送 task_started 事件
    │   └─ 返回 task_id
    ├─ 轮询（每 5 秒）
    │   ├─ 检查状态
    │   ├─ 发送 task_running 事件（新 AI 消息）
    │   └─ 持续直到完成/失败/超时
    └─ 发送 task_completed/task_failed/task_timed_out 事件
```

### 1.2 nanobot-go 现有数据流转（Eino + ADK）

```
用户输入
    ↓
[MasterAgent.Process]
    ├─ ContextBuilder.BuildSystemPrompt()
    │   ├─ getIdentity() - 核心身份
    │   ├─ loadBootstrapFiles() - AGENTS.md, SOUL.md, USER.md, TOOLS.md, IDENTITY.md
    │   ├─ GetAlwaysSkills() - 活动技能
    │   ├─ BuildSkillsSummary() - 技能摘要
    │   └─ BuildMCPServersSection() - MCP Servers
    ├─ BuildMessageList() - 构建消息列表
    └─ interruptible.Process()
        ↓
    [Hook 系统]
        ├─ BeforeMessageProcess
        ├─ BeforeLLMGenerate
        ├─ AfterLLMGenerate
        ├─ BeforeToolCall
        ├─ AfterToolCall
        └─ AfterMessageProcess
        ↓
    [Eino ChatModelAgent 执行]
        ├─ 工具调用
        │   ├─ exec - 执行命令
        │   ├─ read_file - 读取文件
        │   ├─ write_file - 写入文件
        │   ├─ editfile - 编辑文件
        │   ├─ listdir - 列出目录
        │   ├─ websearch - 网络搜索
        │   ├─ webfetch - 获取网页
        │   ├─ message - 发送消息
        │   ├─ askuser - 询问用户
        │   ├─ use_skill - 使用技能
        │   ├─ use_mcp - 使用 MCP
        │   ├─ start_task/get_task/stop_task/list_task - 后台任务
        │   └─ config 工具
        └─ 循环直到完成
```

## 二、核心差距分析

### 2.1 提示词系统

| 特性 | DeerFlow | nanobot-go | 差距 |
|------|----------|-------------|------|
| **架构** | 模块化 Builder 模式 | 文件加载模式 | ⚠️ 需要重构 |
| **分段管理** | 独立分段（Role, Thinking, Clarification, Subagent, WorkingDir, Skills） | 整体文件 | ✅ nanobot 有类似概念但不够细粒度 |
| **动态组合** | Builder 模式动态组合 | 固定文件加载 | ❌ 缺失 |
| **Skills 注入** | 从配置动态加载 | SkillsLoader 加载 | ⚠️ 实现方式不同 |

**行动项**：
- 重构 ContextBuilder，引入 Builder 模式
- 将现有引导文件转换为模块化分段
- 保持向后兼容（支持现有文件格式）

---

### 2.2 Sandbox 系统

| 特性 | DeerFlow | nanobot-go | 差距 |
|------|----------|-------------|------|
| **虚拟路径** | `/mnt/user-data/{workspace,uploads,outputs}` | 直接物理路径 | ❌ 缺失 |
| **路径翻译** | 虚拟 ↔ 物理双向翻译 | 无 | ❌ 缺失 |
| **路径验证** | 严格验证防止遍历 | 基本验证 | ⚠️ 需要增强 |
| **路径掩码** | 错误信息中隐藏物理路径 | 无 | ❌ 缺失 |
| **Skills 路径** | `/mnt/skills` 只读 | 无 | ❌ 缺失 |
| **Sandbox 抽象** | Sandbox 接口 + Provider 模式 | 直接工具实现 | ⚠️ 需要抽象 |

**行动项**：
- 创建 Sandbox 接口和 Provider 模式
- 实现虚拟路径系统（`/mnt/user-data/*`）
- 实现路径翻译和验证
- 集成到现有文件/命令工具

---

### 2.3 中间件链

| 特性 | DeerFlow | nanobot-go | 差距 |
|------|----------|-------------|------|
| **架构** | LangGraph Middleware | Eino AgentMiddleware + Hooks | ⚠️ 概念类似但实现不同 |
| **执行顺序** | 严格固定顺序 | Hooks 事件驱动 | ❌ 需要顺序控制 |
| **ThreadData** | 创建线程目录 | Session 管理 | ⚠️ 需要集成 |
| **Uploads** | 上传文件追踪 | 无 | ❌ 缺失 |
| **DanglingToolCall** | 处理挂起调用 | 无 | ❌ 缺失 |
| **Summarization** | 上下文摘要 | 无 | ❌ 缺失 |
| **TodoList** | 任务追踪 | Task 工具 | ⚠️ 实现方式不同 |
| **Title** | 自动生成标题 | 无 | ❌ 缺失 |
| **Memory** | 记忆提取 | 无 | ❌ 缺失 |
| **ViewImage** | 图像处理 | 有基本支持 | ✅ 已有 |
| **SubagentLimit** | 限制并发 | 无 | ❌ 缺失 |
| **LoopDetection** | 循环检测 | 无 | ❌ 缺失 |
| **Clarification** | 询问澄清 | askuser 工具 | ⚠️ 需要集成到中间件 |

**行动项**：
- 设计 Eino-compatible 中间件系统
- 按 DeerFlow 顺序实现各个中间件
- 集成到现有 MasterAgent 流程

---

### 2.4 子代理系统

| 特性 | DeerFlow | nanobot-go | 差距 |
|------|----------|-------------|------|
| **task() 工具** | 完整子代理委托 | start_task 后台任务 | ⚠️ 概念不同 |
| **子代理类型** | general-purpose, bash | 无类型 | ❌ 缺失 |
| **继承状态** | sandbox_state, thread_data | 无 | ❌ 缺失 |
| **双线程池** | scheduler + execution | 单线程池 | ❌ 缺失 |
| **并发限制** | MAX_CONCURRENT_SUBAGENTS=3 | 无 | ❌ 缺失 |
| **超时机制** | 配置超时 + 双重保护 | 无 | ❌ 缺失 |
| **实时事件** | task_started/running/completed/failed/timed_out | 无 | ❌ 缺失 |
| **AI 消息流** | 实时捕获子代理 AI 消息 | 无 | ❌ 缺失 |
| **工具过滤** | 移除 task() 防止嵌套 | 无 | ❌ 缺失 |

**行动项**：
- 重写 task 工具为完整子代理委托
- 实现 SubagentExecutor
- 实现双线程池
- 实现事件系统
- 集成到现有 Task 系统

---

### 2.5 工具系统

| 特性 | DeerFlow | nanobot-go | 差距 |
|------|----------|-------------|------|
| **bash** | 带路径翻译的 bash 工具 | exec 工具 | ⚠️ 需要增强 |
| **ls** | 树状格式，最多 2 层 | listdir 工具 | ⚠️ 需要增强 |
| **read_file** | 支持行范围 | read_file 工具 | ✅ 已有 |
| **write_file** | 支持 append | write_file 工具 | ✅ 已有 |
| **str_replace** | 单处或全部替换 | editfile 工具 | ⚠️ 不同实现 |
| **present_files** | 展示输出文件 | 无 | ❌ 缺失 |
| **ask_clarification** | 询问澄清 | askuser 工具 | ✅ 已有 |
| **view_image** | 查看图像 | 无 | ❌ 缺失 |
| **task** | 子代理委托 | start_task | ⚠️ 需要重写 |

**行动项**：
- 增强现有工具以支持 Sandbox 路径翻译
- 添加缺失的工具（present_files, view_image）
- 保持工具签名一致性

---

### 2.6 状态管理

| 特性 | DeerFlow | nanobot-go | 差距 |
|------|----------|-------------|------|
| **ThreadState** | 完整状态 schema | interruptible 状态 | ⚠️ 需要扩展 |
| **sandbox** | sandbox_state | 无 | ❌ 缺失 |
| **thread_data** | thread_data（路径等） | Session | ⚠️ 需要集成 |
| **title** | 标题 | 无 | ❌ 缺失 |
| **artifacts** | 产物 | 无 | ❌ 缺失 |
| **todos** | 待办列表 | 无 | ❌ 缺失 |
| **uploaded_files** | 上传文件 | 无 | ❌ 缺失 |
| **viewed_images** | 已查看图像 | 无 | ❌ 缺失 |
| **Checkpoint** | LangGraph Checkpoint | Eino CheckPointStore | ✅ 已有 |

**行动项**：
- 扩展现有状态管理
- 添加缺失的状态字段
- 集成到 Session 系统

---

## 三、分阶段集成计划

### Phase 1: 提示词系统模块化

目标：将 DeerFlow 的模块化提示词系统集成到 nanobot-go

- [ ] 创建 `pkg/agent/prompts/` 包
  - [ ] 定义 PromptSection 接口
  - [ ] 实现所有 DeerFlow 分段（Role, Thinking, Clarification, Subagent, WorkingDir, Skills）
  - [ ] 实现 PromptBuilder
- [ ] 重构 ContextBuilder
  - [ ] 保持向后兼容（支持现有文件）
  - [ ] 内部使用新的 PromptBuilder
- [ ] 单元测试

---

### Phase 2: Sandbox 虚拟路径系统

目标：实现完整的 Sandbox 抽象和虚拟路径系统

- [ ] 创建 `pkg/sandbox/` 包
  - [ ] Sandbox 接口
  - [ ] SandboxProvider 接口
  - [ ] LocalSandbox 实现
  - [ ] 路径翻译工具
  - [ ] 路径验证工具
- [ ] 集成到现有工具
  - [ ] exec 工具 → bash 工具（带路径翻译）
  - [ ] read_file 工具（带路径验证）
  - [ ] write_file 工具（带路径验证）
  - [ ] listdir 工具 → ls 工具（树状格式）
  - [ ] 添加 str_replace 工具
- [ ] 单元测试

---

### Phase 3: 中间件链系统

目标：实现 DeerFlow 完整中间件链

- [ ] 创建 `pkg/agent/middleware/` 包
  - [ ] Middleware 接口（Eino-compatible）
  - [ ] ThreadDataMiddleware
  - [ ] UploadsMiddleware
  - [ ] DanglingToolCallMiddleware
  - [ ] SummarizationMiddleware（可选）
  - [ ] TodoListMiddleware（可选）
  - [ ] TitleMiddleware
  - [ ] MemoryMiddleware（可选）
  - [ ] ViewImageMiddleware
  - [ ] SubagentLimitMiddleware
  - [ ] LoopDetectionMiddleware
  - [ ] ClarificationMiddleware（必须最后！）
- [ ] 集成到 MasterAgent
- [ ] 单元测试

---

### Phase 4: 子代理系统

目标：实现完整的子代理委托系统

- [ ] 创建 `pkg/agent/subagent/` 包
  - [ ] SubagentConfig 定义
  - [ ] SubagentExecutor（双线程池）
  - [ ] 实时事件系统
  - [ ] 子代理注册
- [ ] 重写 task 工具
  - [ ] 支持 general-purpose 和 bash 类型
  - [ ] 继承父状态
  - [ ] 工具过滤（防止嵌套）
- [ ] 集成到现有 Task 系统
- [ ] 单元测试

---

### Phase 5: 状态管理扩展

目标：扩展状态管理以支持 DeerFlow 完整状态

- [ ] 扩展 Session 管理
  - [ ] 添加 sandbox 状态
  - [ ] 添加 thread_data
  - [ ] 添加 title
  - [ ] 添加 artifacts
  - [ ] 添加 todos
  - [ ] 添加 uploaded_files
  - [ ] 添加 viewed_images
- [ ] 集成 CheckpointStore
- [ ] 单元测试

---

### Phase 6: 端到端集成测试

目标：完整的端到端测试，确保一比一复刻

- [ ] 集成所有组件
- [ ] 端到端测试
  - [ ] 生成网页测试
  - [ ] 生成 PPT 测试
  - [ ] 子代理委托测试
  - [ ] Sandbox 路径测试
  - [ ] 中间件链测试
- [ ] 性能优化
- [ ] 文档完善

---

## 四、关键设计决策

### 4.1 向后兼容策略

**原则**：现有功能不受影响，新功能可选启用

- ContextBuilder 保持现有 API，内部逐步迁移
- 现有工具继续工作，新 Sandbox 工具作为增强
- MasterAgent 保持现有接口，中间件可选插入

### 4.2 Eino  vs LangGraph 映射

| DeerFlow (LangGraph) | nanobot-go (Eino) |
|----------------------|-------------------|
| State | Session + Custom State |
| Middleware | AgentMiddleware + Hooks |
| Checkpoint | CheckPointStore |
| Tool | tool.BaseTool |
| Subagent | 嵌套 Agent |

### 4.3 目录结构

```
pkg/
├── agent/
│   ├── prompts/          # 新：模块化提示词系统
│   ├── middleware/       # 新：中间件链
│   ├── subagent/         # 新：子代理系统
│   ├── master_agent.go   # 现有：增强集成
│   └── context.go        # 现有：重构
├── sandbox/              # 新：Sandbox 系统
│   ├── interface.go
│   ├── provider.go
│   ├── path.go
│   └── local/
└── session/              # 现有：扩展状态
```

---

## 五、风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 向后兼容性破坏 | 高 | 中 | 严格保持现有 API，新功能可选 |
| 性能下降 | 中 | 中 | 分阶段集成，每阶段性能测试 |
| 复杂度增加 | 高 | 高 | 完善文档，渐进式迁移 |
| Eino 特性限制 | 中 | 低 | 提前验证 Eino 能力 |

---

## 六、参考资料

- DeerFlow 源码：`deer-flow/backend/packages/harness/deerflow/`
- Eino 文档：https://github.com/cloudwego/eino
- 现有 nanobot-go 代码：`pkg/agent/`
