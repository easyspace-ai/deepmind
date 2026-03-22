# DeerFlow 一比一复刻 - 完整实施计划

## 概述

本文档详细说明如何在 nanobot-go 中一比一复刻 DeerFlow 的完整功能和数据流转路径。

---

## 一、完整架构对比

### DeerFlow 完整架构（Python + LangGraph）

```
┌─────────────────────────────────────────────────────────────────┐
│                         LangGraph Server                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Middleware Chain (12个，严格顺序!)          │  │
│  │  1. ThreadDataMiddleware       - 创建线程目录           │  │
│  │  2. UploadsMiddleware          - 上传文件追踪           │  │
│  │  3. DanglingToolCallMiddleware - 处理挂起工具调用       │  │
│  │  4. SummarizationMiddleware     - 上下文摘要(可选)       │  │
│  │  5. TodoListMiddleware          - 任务追踪(可选)        │  │
│  │  6. TitleMiddleware             - 自动生成标题          │  │
│  │  7. MemoryMiddleware            - 记忆提取(可选)        │  │
│  │  8. ViewImageMiddleware         - 图像处理              │  │
│  │  9. SubagentLimitMiddleware     - 限制并发(可选)       │  │
│  │ 10. LoopDetectionMiddleware     - 循环检测              │  │
│  │ 11. ClarificationMiddleware     - 询问澄清(最后!)       │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Lead Agent (主代理)                    │  │
│  │  - 系统提示词 (模块化构建)                               │  │
│  │  - 可用工具 (5类)                                        │  │
│  │    1. Sandbox Tools (bash, ls, read_file, write_file,  │  │
│  │                        str_replace)                        │  │
│  │    2. Built-in Tools (present_files, ask_clarification,  │  │
│  │                         view_image)                        │  │
│  │    3. MCP Tools (动态加载)                               │  │
│  │    4. Community Tools (tavily, jina_ai, firecrawl,     │  │
│  │                          image_search)                     │  │
│  │    5. Subagent Tool (task - 子代理委托)                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                  ThreadState (状态)                        │  │
│  │  - messages: [Message]                                     │  │
│  │  - sandbox: {sandbox_id}                                  │  │
│  │  - thread_data: {workspace_path, uploads_path,           │  │
│  │                  outputs_path}                             │  │
│  │  - title: str                                              │  │
│  │  - artifacts: [str] (去重合并)                            │  │
│  │  - todos: []                                              │  │
│  │  - uploaded_files: []                                     │  │
│  │  - viewed_images: {path: {base64, mime_type}}           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Subagent System (子代理系统)                 │  │
│  │  - SubagentExecutor                                       │  │
│  │    - 双线程池: _scheduler_pool (3) + _execution_pool (3)│  │
│  │    - MAX_CONCURRENT_SUBAGENTS = 3                        │  │
│  │    - timeout_seconds = 900 (15分钟)                      │  │
│  │  - 事件: task_started / task_running / task_completed /  │  │
│  │          task_failed / task_timed_out                     │  │
│  │  - 工具过滤: 移除 task() 防止递归嵌套                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │               Sandbox System (沙箱系统)                    │  │
│  │  - Sandbox 接口: execute_command, read_file,              │  │
│  │                   write_file, list_dir                     │  │
│  │  - LocalSandboxProvider: 单例本地实现                     │  │
│  │  - 虚拟路径: /mnt/user-data/{workspace,uploads,outputs}  │  │
│  │  - 路径翻译: 虚拟 ↔ 物理 双向                             │  │
│  │  - 路径验证: 防止路径遍历                                 │  │
│  │  - 路径掩码: 输出隐藏物理路径                             │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### nanobot-go 目标架构（Go + Eino）

```
┌─────────────────────────────────────────────────────────────────┐
│                        nanobot-go Server                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Middleware Chain (Eino-compatible)           │  │
│  │  (同样12个，同样严格顺序!)                                 │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              MasterAgent (DeerFlow-Lead 模式)            │  │
│  │  - 系统提示词: 使用 pkg/agent/prompts (已完成!)          │  │
│  │  - 可用工具: 升级为 DeerFlow 风格                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              ThreadState (扩展 Session 管理)               │  │
│  │  (完整 DeerFlow 状态字段)                                 │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Subagent System (子代理系统)                 │  │
│  │  (完整 DeerFlow 子代理功能)                               │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Sandbox System (沙箱系统)                    │  │
│  │  - pkg/sandbox (已完成!)                                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 二、完整差距清单

### 已完成 ✅

| 模块 | 状态 | 位置 |
|------|------|------|
| 模块化提示词系统 | ✅ | `pkg/agent/prompts/` |
| Sandbox 虚拟路径系统 | ✅ | `pkg/sandbox/` |

### 待实施 📋

| 优先级 | 模块 | 工作量 | 说明 |
|--------|------|--------|------|
| P0 | ThreadState 完整状态 | M | 扩展 Session 管理，添加所有 DeerFlow 字段 |
| P0 | 中间件链系统 | L | 12 个中间件，严格顺序 |
| P0 | Built-in 工具 | M | present_files, ask_clarification, view_image |
| P0 | 子代理系统 | L | SubagentExecutor, 双线程池, 事件系统 |
| P1 | Sandbox 工具集成 | M | 升级现有工具使用 Sandbox 路径翻译 |
| P1 | Memory 系统 | L | 记忆提取、队列、更新器 |
| P1 | Title 自动生成 | M | TitleMiddleware |
| P2 | Summarization 摘要 | L | 上下文摘要中间件 |
| P2 | TodoList 任务追踪 | M | TodoListMiddleware |

---

## 三、详细实施计划

### Phase 0: 准备工作（重构现有代码）

**目标**: 清理已创建的不完整代码，重新开始完整复刻

- [ ] 删除 `pkg/sandbox/` 下的所有文件（重新设计）
- [ ] 删除 `pkg/agent/prompts/` 下的所有文件（重新设计）
- [ ] 创建新的目录结构（完全对齐 DeerFlow）

---

### Phase 1: ThreadState 完整状态系统

**目标**: 实现 DeerFlow 完整的 ThreadState，包括所有 reducer

**目录**: `pkg/agent/state/`

**文件**:

```
pkg/agent/state/
├── types.go        # ThreadState 定义
├── reducers.go     # merge_artifacts, merge_viewed_images
└── state_test.go   # 单元测试
```

**关键实现**:

```go
// ThreadState DeerFlow 完整状态
type ThreadState struct {
    Messages       []*schema.Message
    Sandbox        *SandboxState
    ThreadData     *ThreadData
    Title          string
    Artifacts      []string          // 使用 merge_artifacts reducer
    Todos          []TodoItem
    UploadedFiles  []UploadedFile
    ViewedImages   map[string]ViewedImage // 使用 merge_viewed_images reducer
}

// Reducers
func MergeArtifacts(existing, new []string) []string
func MergeViewedImages(existing, new map[string]ViewedImage) map[string]ViewedImage
```

---

### Phase 2: 中间件链系统（核心！）

**目标**: 实现 DeerFlow 完整的 12 个中间件，严格顺序

**目录**: `pkg/agent/middleware/`

**文件**:

```
pkg/agent/middleware/
├── interface.go             # Middleware 接口 (Eino-compatible)
├── chain.go                 # MiddlewareChain 执行器
├── thread_data.go           # 1. ThreadDataMiddleware
├── uploads.go               # 2. UploadsMiddleware
├── dangling_tool_call.go    # 3. DanglingToolCallMiddleware
├── summarization.go         # 4. SummarizationMiddleware (可选)
├── todo_list.go             # 5. TodoListMiddleware (可选)
├── title.go                 # 6. TitleMiddleware
├── memory.go                # 7. MemoryMiddleware (可选)
├── view_image.go            # 8. ViewImageMiddleware
├── subagent_limit.go        # 9. SubagentLimitMiddleware (可选)
├── loop_detection.go        # 10. LoopDetectionMiddleware
├── clarification.go         # 11. ClarificationMiddleware (必须最后!)
└── middleware_test.go       # 单元测试
```

**关键设计**: Eino Middleware 映射

| DeerFlow (LangGraph) | nanobot-go (Eino) |
|----------------------|-------------------|
| `before_agent()` | `BeforeInvoke()` |
| `after_model()` | `AfterModel()` |
| `wrap_tool_call()` | `AroundToolCall()` |

**中间件执行顺序（严格！）**:

```go
// 必须按此顺序执行！
middlewareChain := []Middleware{
    NewThreadDataMiddleware(),     // 1
    NewUploadsMiddleware(),         // 2
    NewDanglingToolCallMiddleware(), // 3
    NewSummarizationMiddleware(),   // 4 (可选)
    NewTodoListMiddleware(),        // 5 (可选)
    NewTitleMiddleware(),           // 6
    NewMemoryMiddleware(),          // 7 (可选)
    NewViewImageMiddleware(),       // 8
    NewSubagentLimitMiddleware(),   // 9 (可选)
    NewLoopDetectionMiddleware(),   // 10
    NewClarificationMiddleware(),   // 11 (必须最后！)
}
```

---

### Phase 3: DeerFlow 风格工具系统

**目标**: 实现 DeerFlow 所有工具，包括 Sandbox 工具和 Built-in 工具

**目录**: `pkg/agent/tools/deerflow/`

**文件**:

```
pkg/agent/tools/deerflow/
├── sandbox/
│   ├── bash.go          # bash 工具（带路径翻译）
│   ├── ls.go            # ls 工具（树状格式，最多2层）
│   ├── read_file.go     # read_file 工具（支持行范围）
│   ├── write_file.go    # write_file 工具（支持 append）
│   └── str_replace.go   # str_replace 工具
├── builtin/
│   ├── present_files.go # present_files 工具
│   ├── ask_clarification.go # ask_clarification 工具
│   └── view_image.go    # view_image 工具
└── tools.go             # GetAvailableTools() 工厂函数
```

**工具清单**:

| 工具 | 说明 | 状态 |
|------|------|------|
| `bash(description, command)` | 执行命令 | ⏳ |
| `ls(description, path)` | 列出目录（树状） | ⏳ |
| `read_file(description, path, start_line?, end_line?)` | 读取文件 | ⏳ |
| `write_file(description, path, content, append?)` | 写入文件 | ⏳ |
| `str_replace(description, path, old_str, new_str, replace_all?)` | 替换字符串 | ⏳ |
| `present_files(filepaths)` | 展示输出文件 | ⏳ |
| `ask_clarification(question, type, context?, options?)` | 询问澄清 | ⏳ |
| `view_image(path)` | 查看图像 | ⏳ |
| `task(description, prompt, subagent_type, max_turns?)` | 子代理委托 | ⏳ (Phase 4) |

---

### Phase 4: 子代理系统（完整！）

**目标**: 实现 DeerFlow 完整的子代理系统，包括双线程池和事件系统

**目录**: `pkg/agent/subagent/`

**文件**:

```
pkg/agent/subagent/
├── types.go           # SubagentConfig, SubagentStatus, SubagentResult
├── executor.go        # SubagentExecutor (核心！)
├── registry.go        # 子代理注册 (general-purpose, bash)
├── task_tool.go       # task() 工具实现
├── events.go          # 事件系统 (task_started, task_running, etc.)
└── subagent_test.go   # 单元测试
```

**关键实现**:

```go
// SubagentExecutor 子代理执行器
type SubagentExecutor struct {
    config         *SubagentConfig
    tools          []tool.BaseTool
    parentModel    string
    sandboxState   *SandboxState
    threadData     *ThreadData
    // ...
}

// 双线程池
var (
    _schedulerPool = workerpool.New(3) // scheduler 线程池
    _executionPool = workerpool.New(3) // execution 线程池
    MAX_CONCURRENT_SUBAGENTS = 3
)

// 执行流程
// 1. task() 工具被调用
// 2. 创建 SubagentExecutor
// 3. executor.execute_async() → 后台任务
// 4. 发送 task_started 事件
// 5. 轮询（每 5 秒）
//    - 检查状态
//    - 发送 task_running 事件（新 AI 消息）
// 6. 完成/失败/超时 → 发送对应事件
```

**子代理类型**:

| 类型 | 说明 | 工具 |
|------|------|------|
| `general-purpose` | 通用子代理 | 所有工具 except task() |
| `bash` | Bash 专家 | bash, ls, read_file, write_file, str_replace |

---

### Phase 5: Memory 记忆系统

**目标**: 实现 DeerFlow 完整的记忆系统

**目录**: `pkg/agent/memory/`

**文件**:

```
pkg/agent/memory/
├── types.go       # MemoryData, Fact, Context
├── queue.go       # DebouncedUpdateQueue (去抖队列)
├── updater.go     # MemoryUpdater (LLM 更新)
├── prompt.go      # 记忆提示词模板
└── storage.go     # JSON 文件存储
```

**记忆数据结构**:

```go
type MemoryData struct {
    UserContext struct {
        WorkContext     string
        PersonalContext string
        TopOfMind       string
    }
    History struct {
        RecentMonths   string
        EarlierContext string
        LongTermBackground string
    }
    Facts []Fact
}

type Fact struct {
    ID        string
    Content   string
    Category  string // preference, knowledge, context, behavior, goal
    Confidence float64 // 0-1
    CreatedAt time.Time
    Source    string
}
```

---

### Phase 6: 端到端集成与测试

**目标**: 集成所有组件，完整测试

- [ ] 集成 MiddlewareChain 到 MasterAgent
- [ ] 集成 ThreadState 到 Session 管理
- [ ] 集成所有工具
- [ ] 集成子代理系统
- [ ] 端到端测试
  - [ ] 生成网页测试
  - [ ] 生成 PPT 测试
  - [ ] 子代理委托测试
  - [ ] Sandbox 路径测试
  - [ ] 中间件链顺序测试

---

## 四、目录结构（最终目标）

```
pkg/agent/
├── prompts/              # ✅ 模块化提示词系统（重构）
│   ├── types.go
│   ├── sections.go
│   ├── builder.go
│   └── prompts_test.go
├── state/                # Phase 1: ThreadState
│   ├── types.go
│   ├── reducers.go
│   └── state_test.go
├── middleware/           # Phase 2: 中间件链
│   ├── interface.go
│   ├── chain.go
│   ├── thread_data.go
│   ├── uploads.go
│   ├── dangling_tool_call.go
│   ├── summarization.go
│   ├── todo_list.go
│   ├── title.go
│   ├── memory.go
│   ├── view_image.go
│   ├── subagent_limit.go
│   ├── loop_detection.go
│   ├── clarification.go
│   └── middleware_test.go
├── tools/deerflow/       # Phase 3: DeerFlow 工具
│   ├── sandbox/
│   │   ├── bash.go
│   │   ├── ls.go
│   │   ├── read_file.go
│   │   ├── write_file.go
│   │   └── str_replace.go
│   ├── builtin/
│   │   ├── present_files.go
│   │   ├── ask_clarification.go
│   │   └── view_image.go
│   └── tools.go
├── subagent/             # Phase 4: 子代理系统
│   ├── types.go
│   ├── executor.go
│   ├── registry.go
│   ├── task_tool.go
│   ├── events.go
│   └── subagent_test.go
├── memory/               # Phase 5: 记忆系统
│   ├── types.go
│   ├── queue.go
│   ├── updater.go
│   ├── prompt.go
│   └── storage.go
├── master_agent.go       # 升级：DeerFlow-Lead 模式
└── context.go            # 保持向后兼容

pkg/sandbox/              # ✅ Sandbox 系统（重构）
├── types.go
├── path.go
├── sandbox.go
├── local/
│   ├── local_sandbox.go
│   └── provider.go
└── sandbox_test.go
```

---

## 五、实施顺序建议

### 第一阶段：核心基础设施（Week 1-2）

1. **Phase 0**: 清理和重构现有代码
2. **Phase 1**: ThreadState 完整状态系统
3. **Phase 2**: 中间件链系统（前 6 个中间件）
   - ThreadDataMiddleware
   - UploadsMiddleware
   - DanglingToolCallMiddleware
   - TitleMiddleware
   - ViewImageMiddleware
   - ClarificationMiddleware

### 第二阶段：工具系统（Week 3）

4. **Phase 3**: DeerFlow 风格工具系统
   - Sandbox 工具（bash, ls, read_file, write_file, str_replace）
   - Built-in 工具（present_files, ask_clarification, view_image）

### 第三阶段：高级功能（Week 4-5）

5. **Phase 2 (续)**: 剩余中间件
   - SummarizationMiddleware
   - TodoListMiddleware
   - MemoryMiddleware
   - SubagentLimitMiddleware
   - LoopDetectionMiddleware
6. **Phase 4**: 子代理系统
7. **Phase 5**: Memory 记忆系统

### 第四阶段：集成测试（Week 6）

8. **Phase 6**: 端到端集成与测试

---

## 六、关键技术决策

### 1. Eino vs LangGraph 映射

| DeerFlow (LangGraph) | nanobot-go (Eino) |
|----------------------|-------------------|
| `AgentState` | `ThreadState` |
| `AgentMiddleware` | `AgentMiddleware` (Eino 接口) |
| `CheckpointSaver` | `CheckPointStore` |
| `Command(goto=END)` | `Interrupt` 系统 |
| `ToolMessage` | `schema.Message` (RoleTool) |

### 2. 向后兼容策略

- 现有 `MasterAgent` 接口保持不变
- 新增加 `DeerFlowLeadAgent` 作为可选替代
- 现有工具继续工作，新工具作为增强
- 渐进式迁移，不破坏现有功能

### 3. 测试策略

- 每个新包都有单元测试（目标覆盖率 > 80%）
- 中间件链集成测试
- 端到端完整流程测试
- 与原版 DeerFlow 行为对比测试

---

## 七、风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| Eino 中间件能力限制 | 高 | 中 | 提前验证，设计备用方案 |
| 性能下降 | 中 | 中 | 每阶段性能测试，优化热点 |
| 复杂度过高 | 高 | 高 | 分阶段交付，完善文档 |
| 向后兼容破坏 | 高 | 低 | 严格保持现有 API |

---

## 总结

本计划详细说明了如何在 nanobot-go 中一比一复刻 DeerFlow 的完整功能。通过 6 个 Phase 的实施，可以实现：

- ✅ 完整的中间件链（12 个，严格顺序）
- ✅ DeerFlow 风格的模块化提示词
- ✅ 完整的 Sandbox 虚拟路径系统
- ✅ 所有 DeerFlow 工具
- ✅ 子代理系统（双线程池 + 事件）
- ✅ Memory 记忆系统
- ✅ ThreadState 完整状态管理

最终目标：**数据流转路径和 Agent 流程与原版 DeerFlow 一比一**，可以实现相同的效果（生成网页、PPT 等）。
