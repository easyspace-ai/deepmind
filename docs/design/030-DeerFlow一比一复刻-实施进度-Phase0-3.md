# DeerFlow 一比一复刻 - 实施进度 (Phase 0-3)

## 概述

本文档总结 DeerFlow 一比一复刻的前 3 个 Phase 实施进度。

---

## 已完成的工作 ✅

### Phase 0: 准备工作 ✅

**目标**: 分析 DeerFlow 完整源码实现细节

**完成内容**:
- 深入分析了 `deer-flow/backend/packages/harness/deerflow/` 源码
- 理解了完整的数据流转路径
- 识别了 12 个中间件的执行顺序
- 分析了 ThreadState 完整结构
- 分析了 Sandbox 虚拟路径系统
- 分析了模块化提示词系统

**关键源码文件分析**:
- `agents/thread_state.py` - ThreadState 定义和 reducers
- `agents/lead_agent/agent.py` - Lead Agent 工厂和中间件链
- `agents/middlewares/*.py` - 12 个中间件实现
- `sandbox/sandbox.py` - Sandbox 抽象接口
- `sandbox/tools.py` - Sandbox 工具实现
- `subagents/executor.py` - 子代理执行器
- `tools/builtins/` - 内置工具

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

**关键实现**:

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

**辅助方法**:
- `AddArtifacts()` - 添加产物
- `AddViewedImages()` / `ClearViewedImages()` - 图像管理
- `AddTodo()` / `UpdateTodoStatus()` / `GetTodoByID()` - 待办管理
- `AddUploadedFile()` / `RemoveUploadedFile()` / `GetUploadedFile()` - 文件管理
- `AddMessage()` / `GetLastMessage()` / `GetUserMessages()` / `GetAssistantMessages()` - 消息管理
- `SetTitle()` / `HasTitle()` - 标题管理
- `SetSandbox()` / `SetThreadData()` - 状态设置
- `GetWorkspacePath()` / `GetUploadsPath()` / `GetOutputsPath()` - 路径获取

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

**核心组件**:

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

**所有提示词分段**:

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

## 目录结构（已完成）

```
pkg/
├── agent/
│   ├── state/              # ✅ Phase 1: ThreadState
│   │   ├── types.go
│   │   ├── reducers.go
│   │   └── state_test.go
│   └── prompts/            # ✅ Phase 2: 模块化提示词
│       ├── types.go
│       ├── sections.go
│       ├── builder.go
│       └── prompts_test.go
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
| **总计** | **63** | **✅ 全部通过** |

---

## 下一步计划

### Phase 4: 中间件链系统（12个中间件）
1. ThreadDataMiddleware - 创建线程目录
2. UploadsMiddleware - 上传文件追踪
3. DanglingToolCallMiddleware - 处理挂起工具调用
4. SummarizationMiddleware - 上下文摘要（可选）
5. TodoListMiddleware - 任务追踪（可选）
6. TitleMiddleware - 自动生成标题
7. MemoryMiddleware - 记忆提取（可选）
8. ViewImageMiddleware - 图像处理
9. SubagentLimitMiddleware - 限制并发（可选）
10. LoopDetectionMiddleware - 循环检测
11. ClarificationMiddleware - 询问澄清（必须最后！）

### Phase 5: DeerFlow 风格工具系统
- Sandbox 工具: bash, ls, read_file, write_file, str_replace
- Built-in 工具: present_files, ask_clarification, view_image
- Subagent 工具: task()

### Phase 6: 子代理系统
- SubagentExecutor - 双线程池
- 事件系统 - task_started/running/completed/failed/timed_out
- 工具过滤 - 移除 task() 防止嵌套

---

## 关键设计决策

### 1. 向后兼容
- 新模块与现有系统完全独立
- 现有 `ContextBuilder` 和工具继续工作
- 新功能可以渐进式集成

### 2. 模块化设计
- 每个包独立可测试
- 清晰的接口定义
- 易于扩展和维护

### 3. 测试覆盖
- 每个新包都有单元测试
- 所有测试通过 ✅
- 总计 63 个测试

### 4. 一比一复刻
- 严格对齐 DeerFlow 的数据结构
- 严格对齐 DeerFlow 的 API 设计
- 严格对齐 DeerFlow 的行为逻辑

---

## 参考资料

- [差距分析文档](./027-DeerFlow深度集成-差距分析.md)
- [完整实施计划](./029-DeerFlow一比一复刻-完整实施计划.md)
- DeerFlow 源码: `deer-flow/backend/packages/harness/deerflow/`
