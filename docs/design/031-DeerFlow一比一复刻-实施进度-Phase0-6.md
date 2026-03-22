# DeerFlow 一比一复刻 - 实施进度 (Phase 0-6)

## 概述

本文档总结 DeerFlow 一比一复刻的完整实施进度。我们已完成核心基础设施的构建，包括：

- ✅ **Phase 0**: 源码分析
- ✅ **Phase 1**: ThreadState 完整状态系统
- ✅ **Phase 2**: 模块化提示词系统
- ✅ **Phase 3**: Sandbox 虚拟路径系统
- ✅ **Phase 4**: 中间件链系统（框架）
- 🔄 **Phase 5**: DeerFlow 风格工具系统（框架）
- ⏳ **Phase 6**: 子代理系统（待完整实现）

---

## 已完成的工作 ✅

### Phase 0: 源码分析 ✅

**目标**: 深入分析 DeerFlow 完整源码实现细节

**完成内容**:
- 分析了 `deer-flow/backend/packages/harness/deerflow/` 完整源码
- 理解了 12 个中间件的执行顺序
- 分析了 ThreadState 完整结构和 reducers
- 分析了 Sandbox 虚拟路径系统
- 分析了模块化提示词系统
- 分析了子代理执行器（双线程池）

**关键源码文件**:
- `agents/thread_state.py` - ThreadState 定义
- `agents/lead_agent/agent.py` - Lead Agent 工厂
- `agents/middlewares/*.py` - 12 个中间件
- `sandbox/sandbox.py` - Sandbox 抽象接口
- `subagents/executor.py` - 子代理执行器
- `tools/builtins/*.py` - 内置工具

---

### Phase 1: ThreadState 完整状态系统 ✅

**目录**: `pkg/agent/state/`

**文件**:
```
pkg/agent/state/
├── types.go         # ThreadState 定义
├── reducers.go      # Reducers 和辅助方法
└── state_test.go    # 21 个单元测试 ✅
```

**核心类型**:

1. **SandboxState** - 沙箱状态
   ```go
   type SandboxState struct {
       SandboxID string `json:"sandbox_id,omitempty"`
   }
   ```

2. **ThreadDataState** - 线程数据状态
   ```go
   type ThreadDataState struct {
       WorkspacePath string `json:"workspace_path,omitempty"`
       UploadsPath   string `json:"uploads_path,omitempty"`
       OutputsPath   string `json:"outputs_path,omitempty"`
   }
   ```

3. **ViewedImageData** - 已查看图像数据
   ```go
   type ViewedImageData struct {
       Base64   string `json:"base64"`
       MimeType string `json:"mime_type"`
   }
   ```

4. **TodoItem** - 待办事项
   ```go
   type TodoItem struct {
       ID          string    `json:"id"`
       Description string    `json:"description"`
       Status      string    `json:"status"` // pending, in_progress, completed
       CreatedAt   time.Time `json:"created_at"`
       UpdatedAt   time.Time `json:"updated_at"`
   }
   ```

5. **UploadedFile** - 已上传文件
   ```go
   type UploadedFile struct {
       Filename    string    `json:"filename"`
       Path        string    `json:"path"`
       Size        int64     `json:"size"`
       ContentType string    `json:"content_type"`
       UploadedAt  time.Time `json:"uploaded_at"`
   }
   ```

6. **ThreadState** - 完整线程状态
   ```go
   type ThreadState struct {
       Messages       []*schema.Message
       Sandbox        *SandboxState
       ThreadData     *ThreadDataState
       Title          string
       Artifacts      []string
       Todos          []TodoItem
       UploadedFiles  []UploadedFile
       ViewedImages   map[string]ViewedImageData
   }
   ```

**Reducers**:
- `MergeArtifacts()` - 合并产物列表，去重并保持顺序
- `MergeViewedImages()` - 合并已查看图像，支持清空操作

**辅助方法** (20+):
- `AddArtifacts()` / `AddViewedImages()` / `ClearViewedImages()`
- `AddTodo()` / `UpdateTodoStatus()` / `GetTodoByID()`
- `AddUploadedFile()` / `RemoveUploadedFile()` / `GetUploadedFile()`
- `AddMessage()` / `GetLastMessage()` / `GetUserMessages()` / `GetAssistantMessages()`
- `SetTitle()` / `HasTitle()`
- `SetSandbox()` / `SetThreadData()`
- `GetWorkspacePath()` / `GetUploadsPath()` / `GetOutputsPath()`

**单元测试**: 21 个测试全部通过 ✅

---

### Phase 2: 模块化提示词系统 ✅

**目录**: `pkg/agent/prompts/`

**文件**:
```
pkg/agent/prompts/
├── types.go         # 类型定义（PromptSection, BaseSection, NamedSection, Prompt）
├── sections.go      # 所有提示词分段实现
├── builder.go       # Builder 模式 + 预设函数
└── prompts_test.go  # 34 个单元测试 ✅
```

**核心接口**:

1. **PromptSection 接口**
   ```go
   type PromptSection interface {
       Name() string
       Render() string
   }
   ```

2. **BaseSection** - 基础分段实现
3. **NamedSection** - 函数式分段
4. **Prompt** - 完整提示词容器

**所有提示词分段** (11个):

| 分段 | 说明 |
|------|------|
| `RoleSection` | 角色定义 |
| `ThinkingStyleSection` | 思考方式（支持带子代理） |
| `ClarificationSection` | 澄清询问系统 |
| `SubagentSection` | 子代理说明（并发限制） |
| `WorkingDirSection` | 工作目录说明 |
| `SkillsSection` | 技能列表 |
| `EnvInfoSection` | 环境信息（日期、OS、Go 版本） |
| `ResponseStyleSection` | 响应风格 |
| `CitationsSection` | 引用格式 |
| `CriticalRemindersSection` | 关键提醒（支持带子代理） |
| `CustomSection` | 自定义分段 |

**Builder 模式**:
```go
prompt := prompts.NewBuilder().
    WithRole("DeerFlow 2.0").
    WithThinkingStyle().
    WithClarification().
    WithWorkingDir().
    WithEnvInfo().
    Build()
```

**预设函数**:
- `BuildLeadAgentPrompt()` - 构建 Lead Agent 提示词
- `BuildGeneralPurposeSubagentPrompt()` - 构建通用子代理提示词
- `BuildBashSubagentPrompt()` - 构建 Bash 子代理提示词
- `DefaultLeadAgentConfig()` - 默认配置

**单元测试**: 34 个测试全部通过 ✅

---

### Phase 3: Sandbox 虚拟路径系统 ✅

**目录**: `pkg/sandbox/`

**文件**:
```
pkg/sandbox/
├── types.go                  # Sandbox 接口和类型定义
├── path.go                   # 路径翻译系统
├── sandbox_test.go           # 6 个单元测试 ✅
└── local/
    ├── local_sandbox.go      # LocalSandbox 实现
    └── local_sandbox_test.go # 2 个单元测试 ✅
```

**核心接口**:

1. **Sandbox 接口**
   ```go
   type Sandbox interface {
       ID() string
       ExecuteCommand(command string) (string, error)
       ReadFile(path string) (string, error)
       WriteFile(path, content string, append bool) error
       ListDir(path string, maxDepth int) ([]string, error)
       UpdateFile(path string, content []byte) error
   }
   ```

2. **SandboxProvider 接口**
   ```go
   type SandboxProvider interface {
       Acquire(threadID string) (Sandbox, error)
       Get(threadID string) (Sandbox, bool)
       Release(threadID string) error
   }
   ```

**虚拟路径系统**:

| 虚拟路径 | 说明 |
|----------|------|
| `/mnt/user-data/workspace` | 工作区 |
| `/mnt/user-data/uploads` | 上传文件 |
| `/mnt/user-data/outputs` | 输出文件 |
| `/mnt/skills` | 技能目录 |

**PathTranslator 核心功能**:
- `ToPhysical()` - 虚拟路径 → 物理路径
- `ToVirtual()` - 物理路径 → 虚拟路径（掩码）
- `ValidatePath()` - 路径验证（防止遍历）
- `MaskPathsInOutput()` - 输出路径掩码
- `TranslatePathsInCommand()` - 命令中路径翻译
- `IsVirtualPath()` - 检查是否是虚拟路径
- `ExtractVirtualPaths()` - 从文本提取虚拟路径

**LocalSandbox 实现**:
- `ExecuteCommand()` - 执行 bash 命令（带路径翻译）
- `ReadFile()` - 读取文件
- `WriteFile()` - 写入/追加文件
- `ListDir()` - 树状目录列表（最多 2 层）
- `UpdateFile()` - 更新二进制文件

**LocalSandboxProvider (单例)**:
- `Acquire()` - 获取沙箱（创建线程目录）
- `Get()` - 获取已存在的沙箱
- `Release()` - 释放沙箱
- 线程安全（sync.RWMutex）
- 自动创建目录结构

**单元测试**: 8 个测试全部通过 ✅

---

### Phase 4: 中间件链系统（框架）✅

**目录**: `pkg/agent/middleware/`

**文件**:
```
pkg/agent/middleware/
├── types.go            # 中间件接口和链
├── loop_detection.go   # 循环检测中间件（框架）
└── middleware_test.go  # 4 个单元测试 ✅
```

**核心接口**:

1. **Middleware 接口**
   ```go
   type Middleware interface {
       Name() string
   }
   ```

2. **BaseMiddleware** - 基础中间件实现

3. **MiddlewareChain** - 中间件链
   - `Add()` - 添加中间件
   - `AddAll()` - 添加多个中间件
   - `Middlewares()` - 获取所有中间件

4. **EinoCallbackBridge** - Eino 回调桥接器
   - 将 DeerFlow 风格中间件适配到 Eino 的 callbacks.Handler

**已实现中间件框架**:

- **LoopDetectionMiddleware** - 循环检测中间件框架
  - 配置: warnThreshold, hardLimit, windowSize, maxTrackedThreads
  - 完整实现见 `loop_detection.go`

**中间件执行顺序（DeerFlow 严格顺序）**:

1. `ThreadDataMiddleware` - 创建线程目录
2. `UploadsMiddleware` - 上传文件追踪
3. `DanglingToolCallMiddleware` - 处理挂起工具调用
4. `SummarizationMiddleware` - 上下文摘要（可选）
5. `TodoListMiddleware` - 任务追踪（可选）
6. `TitleMiddleware` - 自动生成标题
7. `MemoryMiddleware` - 记忆提取（可选）
8. `ViewImageMiddleware` - 图像处理
9. `SubagentLimitMiddleware` - 限制并发（可选）
10. `LoopDetectionMiddleware` - 循环检测
11. `ClarificationMiddleware` - 询问澄清（必须最后！）

**单元测试**: 4 个测试全部通过 ✅

---

### Phase 5: DeerFlow 风格工具系统（框架）🔄

**目录**: `pkg/agent/tools/deerflow/`

**文件**:
```
pkg/agent/tools/deerflow/
└── types.go  # 工具类型和工厂函数
```

**工具组**:
- `ToolGroupSandbox` - Sandbox 工具
- `ToolGroupBuiltin` - 内置工具
- `ToolGroupMCP` - MCP 工具
- `ToolGroupCommunity` - 社区工具
- `ToolGroupSubagent` - 子代理工具

**核心函数**:
- `GetAvailableTools()` - 获取可用工具
- `BuildLeadAgentTools()` - 构建 Lead Agent 工具列表
- `BuildGeneralPurposeSubagentTools()` - 构建通用子代理工具列表
- `BuildBashSubagentTools()` - 构建 Bash 子代理工具列表
- `BuildLeadAgentPromptWithTools()` - 构建带工具提示的 Lead Agent 提示词

**Sandbox 工具（占位符）**:
- `bash` - 执行命令
- `ls` - 列出目录（树状）
- `read_file` - 读取文件（支持行范围）
- `write_file` - 写入文件（支持 append）
- `str_replace` - 字符串替换

**Built-in 工具（占位符）**:
- `present_files` - 展示输出文件
- `ask_clarification` - 询问澄清
- `view_image` - 查看图像

**Subagent 工具（占位符）**:
- `task` - 子代理委托

---

### Phase 6: 子代理系统（待完整实现）⏳

**待创建目录**: `pkg/agent/subagent/`

**计划文件**:
```
pkg/agent/subagent/
├── types.go           # SubagentConfig, SubagentStatus, SubagentResult
├── executor.go        # SubagentExecutor (核心！)
├── registry.go        # 子代理注册 (general-purpose, bash)
├── task_tool.go       # task() 工具实现
├── events.go          # 事件系统 (task_started, task_running, etc.)
└── subagent_test.go   # 单元测试
```

**关键设计**:

1. **SubagentExecutor** - 子代理执行器
   - 双线程池: `_scheduler_pool` (3) + `_execution_pool` (3)
   - `MAX_CONCURRENT_SUBAGENTS = 3`
   - `timeout_seconds = 900` (15分钟)

2. **事件系统**:
   - `task_started`
   - `task_running`
   - `task_completed`
   - `task_failed`
   - `task_timed_out`

3. **子代理类型**:
   - `general-purpose` - 通用子代理（所有工具 except task）
   - `bash` - Bash 专家（bash, ls, read_file, write_file, str_replace）

4. **工具过滤**: 移除 task() 防止递归嵌套

---

## 目录结构（最终目标）

```
pkg/
├── agent/
│   ├── state/              # ✅ Phase 1: ThreadState
│   │   ├── types.go
│   │   ├── reducers.go
│   │   └── state_test.go
│   ├── prompts/            # ✅ Phase 2: 模块化提示词
│   │   ├── types.go
│   │   ├── sections.go
│   │   ├── builder.go
│   │   └── prompts_test.go
│   ├── middleware/         # ✅ Phase 4: 中间件链（框架）
│   │   ├── types.go
│   │   ├── loop_detection.go
│   │   └── middleware_test.go
│   └── tools/deerflow/     # 🔄 Phase 5: 工具系统（框架）
│       └── types.go
└── sandbox/                # ✅ Phase 3: Sandbox 系统
    ├── types.go
    ├── path.go
    ├── sandbox_test.go
    └── local/
        ├── local_sandbox.go
        └── local_sandbox_test.go
```

---

## 测试覆盖统计

| 模块 | 测试数 | 状态 |
|------|--------|------|
| `pkg/agent/state` | 21 | ✅ 全部通过 |
| `pkg/agent/prompts` | 34 | ✅ 全部通过 |
| `pkg/sandbox` | 6 | ✅ 全部通过 |
| `pkg/sandbox/local` | 2 | ✅ 全部通过 |
| `pkg/agent/middleware` | 4 | ✅ 全部通过 |
| **总计** | **67** | **✅ 全部通过** |

---

## 关键一比一复刻点

### 1. ThreadState 结构
- ✅ 完全对齐 DeerFlow 的字段
- ✅ `MergeArtifacts()` reducer - 去重并保持顺序
- ✅ `MergeViewedImages()` reducer - 支持清空操作

### 2. 模块化提示词
- ✅ 11 个提示词分段完整复刻
- ✅ Builder 模式与 DeerFlow 一致
- ✅ 预设函数与 DeerFlow 对齐

### 3. Sandbox 虚拟路径
- ✅ 虚拟路径 `/mnt/user-data/{workspace,uploads,outputs}`
- ✅ 路径双向翻译（虚拟 ↔ 物理）
- ✅ 路径验证（防止遍历）
- ✅ 路径掩码（输出隐藏物理路径）
- ✅ LocalSandboxProvider 单例模式

### 4. 中间件链（框架）
- ✅ 中间件接口定义
- ✅ 中间件链实现
- ✅ Eino 回调桥接器
- ✅ LoopDetectionMiddleware 框架

### 5. 工具系统（框架）
- ✅ 工具分组定义
- ✅ 工厂函数框架
- ✅ 所有工具占位符

---

## 后续工作建议

### 优先级 P0
1. **完整中间件实现**
   - 实现剩余 11 个中间件
   - 严格按照 DeerFlow 顺序
   - 集成到 Eino 回调系统

2. **完整工具实现**
   - 实现 Sandbox 工具（带路径翻译）
   - 实现 Built-in 工具
   - 集成到现有工具系统

3. **子代理系统**
   - 实现 SubagentExecutor（双线程池）
   - 实现事件系统
   - 实现 task() 工具

### 优先级 P1
4. **端到端集成**
   - 集成所有组件到 MasterAgent
   - 完整端到端测试
   - 性能优化

---

## 参考资料

- [差距分析文档](./027-DeerFlow深度集成-差距分析.md)
- [完整实施计划](./029-DeerFlow一比一复刻-完整实施计划.md)
- [Phase 0-3 进度](./030-DeerFlow一比一复刻-实施进度-Phase0-3.md)
- DeerFlow 源码: `deer-flow/backend/packages/harness/deerflow/`
