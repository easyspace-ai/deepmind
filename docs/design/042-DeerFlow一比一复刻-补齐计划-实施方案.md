# DeerFlow 一比一复刻 - 补齐计划与实施方案

> **文档目的**：基于当前对齐现状（约 60% 完成度），制定分阶段补齐计划，明确每个阶段的任务、验收标准和时间估算。

---

## 变更记录表

| 日期 | 版本 | 变更说明 |
|------|------|----------|
| 2026-03-22 | v1 | 初版：基于 v6 进度，制定四阶段补齐计划 |

---

## 一、现状概览

### 当前完成度

| 模块 | 完成度 | 说明 |
|------|--------|------|
| 状态系统 | 95% | ThreadState, reducers, ApplyMiddlewareUpdates |
| 提示词系统 | 95% | 17 个分段完整 |
| Sandbox 系统 | 70% | 接口、PathTranslator、LocalSandbox |
| 中间件系统 | 55% | 链已装配，业务逻辑待完善 |
| 工具系统 | 60% | 基础工具完成，缺 tool_search/社区工具 |
| 子代理系统 | 80% | Executor/WorkerPool/事件系统 |
| Memory 系统 | 20% | 队列框架，缺 LLM 提取 |
| Skills 系统 | 50% | Loader 基础，缺完整集成 |
| MCP 系统 | 30% | 有基础，缺多服/OAuth/mtime |
| **总体** | **~58%** | - |

### 关键差距排序（按影响度）

1. **中间件业务逻辑** - 14 个中间件都需要从占位符 → 完整实现
2. **Memory LLM 事实提取** - 核心功能缺失
3. **tool_search + DeferredTool** - 工具生态缺失
4. **社区工具** (tavily/jina/firecrawl/image_search)
5. **MCP 深度对齐** (多服/OAuth/mtime 缓存)

---

## 二、补齐策略

### 指导原则

1. **先闭环，后完善**：先让 Lead Agent 能端到端跑通，再补全边缘功能
2. **自底向上**：从底层依赖（中间件）→ 上层功能（Memory/MCP）
3. **测试驱动**：每个模块补充单元测试，确保不回退
4. **文档同步**：每个阶段更新进度文档

---

## 三、分阶段实施计划

### Phase 0: 准备阶段（1 天）

#### 目标
- 搭建开发环境，确认所有现有测试通过
- 建立 DeerFlow Python 代码对照索引
- 制定详细的中间件实现规范

#### 任务清单

| 任务 | 说明 |
|------|------|
| 0.1 | 运行 `go test ./pkg/...` 确认现有测试通过 |
| 0.2 | 创建 `docs/design/deerflow-reference/` 目录，存放 Python 关键代码片段 |
| 0.3 | 编写《中间件实现规范》，明确 BeforeAgent/AfterModel/WrapToolCall 等钩子的职责 |

#### 交付物
- `docs/design/deerflow-reference/` - Python 对照代码库
- `docs/design/042-1-中间件实现规范.md`

---

### Phase 1: 中间件业务逻辑补齐（5-7 天）- **P0 最高优先级**

#### 目标
将 14 个中间件从"基础框架"提升到"完整业务逻辑"，与 DeerFlow Python 行为对齐。

#### 任务清单（按依赖顺序）

##### 1.1 ThreadDataMiddleware（1 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/thread_data_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| 创建线程目录结构 | `{workspace,uploads,outputs}` 目录自动创建 |
| 计算路径 | `thread_data.{workspace_path,uploads_path,outputs_path}` 正确填充 |
| Lazy init 支持 | 目录只在首次访问时创建 |
| 与 pkg/config/paths 集成 | 使用统一的路径解析逻辑 |
| 单元测试 | 覆盖创建、复用、清理场景 |

**文件**：`pkg/agent/middleware/thread_data_middleware.go`

---

##### 1.2 UploadsMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/uploads_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| 从 `message.additional_kwargs.files` 提取新上传文件 | 正确解析并加入 `uploaded_files` |
| 扫描历史上传文件 | 合并历史与新上传，去重 |
| 生成 `<uploaded_files>` 消息块 | XML 格式正确 |
| 预处理 HumanMessage 内容 | 注入文件列表到消息 |
| 单元测试 | 覆盖新增、历史、重复场景 |

**文件**：`pkg/agent/middleware/uploads_middleware.go`

---

##### 1.3 SandboxMiddleware（1 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/sandbox/middleware.py`

| 任务 | 验收标准 |
|------|----------|
| 从 `SandboxProvider` acquire sandbox | 成功获取 sandbox_id |
| 存储 sandbox_id 到 state | `state.Sandbox.SandboxID` 正确设置 |
| Lazy init 支持 | 只在首次需要时 acquire |
| 与 `pkg/sandbox` 集成 | 使用统一的 Sandbox 接口 |
| Release 在 cleanup 时调用 | 资源正确释放 |
| 单元测试 | 覆盖 acquire/release/复用场景 |

**文件**：`pkg/agent/middleware/sandbox_middleware.go`

**依赖**：Phase 1.1 (ThreadDataMiddleware)

---

##### 1.4 DanglingToolCallMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/dangling_tool_call_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| 检测 AIMessage.tool_calls 缺失对应 ToolMessage | 正确识别挂起调用 |
| `wrap_model_call` / `awrap_model_call` 实现 | 钩子正确挂载 |
| 插入 synthetic ToolMessage | 包含正确的 message ID |
| 状态 "error"，内容 "tool call interrupted" | 格式与 DeerFlow 一致 |
| 单元测试 | 覆盖检测、插入、无挂起场景 |

**文件**：`pkg/agent/middleware/dangling_tool_call_middleware.go`

---

##### 1.5 SummarizationMiddleware（1 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/lead_agent/agent.py` (SummarizationMiddleware 创建)

| 任务 | 验收标准 |
|------|----------|
| Token/message 触发检测 | 按 config.trigger 正确触发 |
| 模型调用生成摘要 | 使用 config.model_name |
| Keep 策略实现 | `keep.last_n` / `keep.first_n` / `keep.fraction` |
| 与 `pkg/config` 集成 | 从 AppConfig.Summarization 读取配置 |
| 单元测试 | 覆盖触发、keep 策略、禁用场景 |

**文件**：`pkg/agent/middleware/summarization_middleware.go`

**依赖**：需要模型调用能力（可先用 mock）

---

##### 1.6 TodoListMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/todo_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| `write_todos` 工具注入 | 工具正确绑定到 state |
| 系统提示词注入 | `<todo_list_system>` 完整 |
| 实时 todo 状态管理 | `state.Todos` 实时更新 |
| `is_plan_mode` 条件启用 | 只在 plan_mode 时启用 |
| 单元测试 | 覆盖启用/禁用、更新场景 |

**文件**：`pkg/agent/middleware/todo_list_middleware.go`

---

##### 1.7 TitleMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/title_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| 首次消息交换检测 | 正确识别第一次完整交互 |
| 标题模型调用 | 使用 config.title.model_name |
| 规范化结构化内容 | 清理消息后再送模型 |
| 标题生成提示词 | 与 DeerFlow 一致 |
| 单元测试 | 覆盖首次、非首次、禁用场景 |

**文件**：`pkg/agent/middleware/title_middleware.go`

---

##### 1.8 MemoryMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/memory_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| 消息过滤（用户输入 + 最终 AI 响应） | 正确过滤，忽略中间工具调用 |
| 去抖队列 | 使用 `pkg/agent/memory` 的队列 |
| 异步内存更新 | 不阻塞主流程 |
| 与 `pkg/agent/memory` 集成 | 调用 `EnqueueFromThreadState` |
| 单元测试 | 覆盖过滤、去抖、禁用场景 |

**文件**：`pkg/agent/middleware/memory_middleware.go`

**依赖**：Phase 2 (Memory 系统) 可并行

---

##### 1.9 ViewImageMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/view_image_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| 图像文件读取 | 从 uploads 路径读取 |
| Base64 编码 | 正确编码为 data URL |
| 注入 viewed_images 到 state | `state.ViewedImages` 正确填充 |
| 视觉支持模型条件判断 | 只在 `supports_vision` 时启用 |
| 单元测试 | 覆盖有图、无图、无视觉支持场景 |

**文件**：`pkg/agent/middleware/view_image_middleware.go`

---

##### 1.10 SubagentLimitMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/subagent_limit_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| `task()` 工具调用计数 | 正确统计数量 |
| 超出限制时截断 | 保留前 N 个，丢弃超出 |
| `after_model` / `aafter_model` 实现 | 钩子正确挂载 |
| 2-4 范围限制 | 默认 3，可配置 |
| 单元测试 | 覆盖未超限、超限、禁用场景 |

**文件**：`pkg/agent/middleware/subagent_limit_middleware.go`

---

##### 1.11 LoopDetectionMiddleware（已有基础，完善 0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/loop_detection_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| Order-independent 哈希完整实现 | 工具调用顺序不影响哈希 |
| LRU 缓存 | 最近 N 个状态的缓存 |
| 警告/硬停止阈值 | `warn_threshold` / `hard_limit` |
| 单元测试 | 覆盖重复、相似、非重复场景 |

**文件**：`pkg/agent/middleware/loop_detection.go`

---

##### 1.12 ToolErrorHandlingMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/tool_error_handling_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| `wrap_tool_call` / `awrap_tool_call` 实现 | 钩子正确挂载 |
| 异常捕获与转换 | panic → error 转换 |
| 生成 ToolMessage (status="error") | 格式与 DeerFlow 一致 |
| GraphBubbleUp 透传 | 错误正确向上传递 |
| 单元测试 | 覆盖异常、正常、透传场景 |

**文件**：`pkg/agent/middleware/tool_error_handling_middleware.go`

---

##### 1.13 DeferredToolFilterMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/deferred_tool_filter_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| 延迟工具 schema 隐藏 | 模型不可见延迟工具 |
| `tool_search` 启用条件 | 只在 config.tool_search.enabled 时启用 |
| `<available-deferred-tools>` 注入 | 提示词中列出延迟工具名 |
| 单元测试 | 覆盖启用/禁用、隐藏场景 |

**文件**：`pkg/agent/middleware/deferred_tool_filter_middleware.go`

**依赖**：Phase 3.1 (DeferredTool 系统)

---

##### 1.14 ClarificationMiddleware（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/middlewares/clarification_middleware.py`

| 任务 | 验收标准 |
|------|----------|
| `ask_clarification` 工具调用拦截 | 正确识别该工具 |
| 中断执行 | 使用 Eino 的中断机制 |
| 必须在最后位置 | 链顺序正确 |
| 单元测试 | 覆盖拦截、无请求、禁用场景 |

**文件**：`pkg/agent/middleware/clarification_middleware.go`

---

#### Phase 1 交付物
- 14 个中间件完整业务逻辑
- 每个中间件的单元测试（覆盖率 ≥ 80%）
- `docs/design/042-2-Phase1-中间件补齐总结.md`
- **完成度提升**：中间件系统 55% → **95%**，总体 58% → **70%**

---

### Phase 2: Memory 系统完整实现（2-3 天）- **P0 高优先级**

#### 目标
实现 Memory 系统的 LLM 事实提取、去重、原子写入等核心功能。

#### 任务清单

##### 2.1 Memory Updater（1.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/memory/updater.py`

| 任务 | 验收标准 |
|------|----------|
| LLM 事实提取提示词 | 与 DeerFlow 一致 |
| UserContext 提取 | workContext/personalContext/topOfMind |
| History 提取 | recentMonths/earlierContext/longTermBackground |
| Facts 提取 | id/content/category/confidence/source |
| 空格规范化事实去重 | trim 后比较，避免重复 |
| 原子文件 I/O | temp file + rename，避免损坏 |
| 单元测试 | 覆盖提取、去重、写入场景 |

**文件**：`pkg/agent/memory/updater.go`

---

##### 2.2 Memory Queue 完善（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/memory/queue.py`

| 任务 | 验收标准 |
|------|----------|
| 每线程去重 | 同一线程短时间内多次更新只处理一次 |
| 可配置等待时间 | debounce_seconds 配置 |
| 后台处理 goroutine | 不阻塞主流程 |
| 单元测试 | 覆盖去重、等待、关闭场景 |

**文件**：`pkg/agent/memory/queue.go`

---

##### 2.3 Memory Prompt（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/agents/memory/prompt.py`

| 任务 | 验收标准 |
|------|----------|
| 内存更新提示词模板 | 与 DeerFlow 一致 |
| `format_memory_for_injection()` | 按 token 限制格式化 |
| 提示词分段集成 | `prompts.MemorySection` 使用此函数 |
| 单元测试 | 覆盖格式化、截断、空场景 |

**文件**：`pkg/agent/memory/prompt.go`

---

#### Phase 2 交付物
- Memory Updater 完整实现
- Memory Queue 完善
- Memory Prompt 模板
- 单元测试覆盖率 ≥ 80%
- `docs/design/042-3-Phase2-Memory补齐总结.md`
- **完成度提升**：Memory 系统 20% → **90%**，总体 70% → **75%**

---

### Phase 3: 工具系统补齐（3-4 天）- **P1 中优先级**

#### 目标
补齐 tool_search、DeferredTool 注册表、社区工具。

#### 任务清单

##### 3.1 DeferredTool 系统（1 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/tools/builtins/tool_search.py`

| 任务 | 验收标准 |
|------|----------|
| `DeferredToolRegistry` 类型 | 存储延迟工具 |
| `get_deferred_registry()` 函数 | 全局单例 |
| `tool_search` 工具实现 | 搜索延迟工具 |
| 与 `DeferredToolFilterMiddleware` 集成 | 正确隐藏/显示 |
| 单元测试 | 覆盖注册、搜索、过滤场景 |

**文件**：
- `pkg/agent/tools/deerflow/deferred_tool_registry.go`
- `pkg/agent/tools/deerflow/tool_search_tool.go`

---

##### 3.2 社区工具 - Tavily（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/community/tavily/tools.py`

| 任务 | 验收标准 |
|------|----------|
| `tavily_search` 工具 | Web 搜索（默认 5 结果）|
| `tavily_fetch` 工具 | Web 获取（4KB 限制）|
| API Key 配置 | 从环境变量读取 |
| 单元测试 | 使用 mock 客户端 |

**文件**：`pkg/agent/tools/community/tavily/`

---

##### 3.3 社区工具 - Jina AI（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/community/jina_ai/`

| 任务 | 验收标准 |
|------|----------|
| `jina_fetch` 工具 | 通过 Jina Reader API 获取 |
| Readability 提取 | 提取主要内容 |
| API Key 配置 | 从环境变量读取 |
| 单元测试 | 使用 mock 客户端 |

**文件**：`pkg/agent/tools/community/jina/`

---

##### 3.4 社区工具 - Firecrawl（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/community/firecrawl/`

| 任务 | 验收标准 |
|------|----------|
| `firecrawl_scrape` 工具 | 网页爬取 |
| API Key 配置 | 从环境变量读取 |
| 单元测试 | 使用 mock 客户端 |

**文件**：`pkg/agent/tools/community/firecrawl/`

---

##### 3.5 社区工具 - Image Search（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/community/image_search/`

| 任务 | 验收标准 |
|------|----------|
| `image_search` 工具 | DuckDuckGo 图片搜索 |
| 返回格式 | URL + 缩略图 |
| 单元测试 | 使用 mock 客户端 |

**文件**：`pkg/agent/tools/community/imagesearch/`

---

##### 3.6 工具链完整组装（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/tools/tools.py`

| 任务 | 验收标准 |
|------|----------|
| `GetAvailableTools()` 函数 | 组装所有工具 |
| 配置定义工具 | 从 config.tools 加载 |
| MCP 工具（可选）| 预留集成点 |
| 内置工具 | present_files/ask_clarification/view_image |
| 子代理工具 | task()（如启用）|
| 社区工具 | tavily/jina/firecrawl/image_search |
| 单元测试 | 覆盖组装、过滤、禁用场景 |

**文件**：`pkg/agent/tools/deerflow/tools.go`

---

#### Phase 3 交付物
- DeferredTool 系统 + tool_search
- 4 个社区工具
- 完整工具链组装
- 单元测试覆盖率 ≥ 75%
- `docs/design/042-4-Phase3-工具补齐总结.md`
- **完成度提升**：工具系统 60% → **90%**，总体 75% → **82%**

---

### Phase 4: MCP 系统深度对齐（3-4 天）- **P1 中优先级**

#### 目标
对齐 DeerFlow 的 MCP 多服务器管理、mtime 缓存失效、OAuth 支持。

#### 任务清单

##### 4.1 MCP MultiServerMCPClient（1.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/mcp/client.py`

| 任务 | 验收标准 |
|------|----------|
| 多服务器管理 | 同时连接多个 MCP 服务器 |
| Transports: stdio/SSE/HTTP | 三种传输方式支持 |
| 工具懒加载 | 首次调用时加载 |
| 单元测试 | 使用 mock 服务器 |

**文件**：`pkg/agent/mcp/client.go`

---

##### 4.2 MCP mtime 缓存失效（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/mcp/cache.py`

| 任务 | 验收标准 |
|------|----------|
| 配置文件 mtime 检测 | 比较修改时间 |
| 缓存失效 | 配置变更时自动重载 |
| `get_cached_mcp_tools()` | 缓存工具列表 |
| 单元测试 | 覆盖缓存、失效、命中场景 |

**文件**：`pkg/agent/mcp/cache.go`

---

##### 4.3 MCP OAuth 支持（1 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/mcp/oauth.py`

| 任务 | 验收标准 |
|------|----------|
| `client_credentials` 流程 | 客户端凭证模式 |
| `refresh_token` 流程 | 刷新令牌模式 |
| 自动 token 刷新 | 过期前自动刷新 |
| Authorization header 注入 | 自动添加到请求 |
| 单元测试 | 使用 mock OAuth 服务器 |

**文件**：`pkg/agent/mcp/oauth.go`

---

##### 4.4 MCP 工具集成（0.5 天）
**DeerFlow 参考**：`deer-flow/backend/packages/harness/deerflow/mcp/tools.py`

| 任务 | 验收标准 |
|------|----------|
| MCP 工具转换 | MCP tool → Eino tool |
| 与 `GetAvailableTools()` 集成 | 可选包含 MCP 工具 |
| 懒加载触发 | 首次调用时初始化 |
| 单元测试 | 覆盖转换、调用、错误场景 |

**文件**：`pkg/agent/mcp/tools.go`

---

#### Phase 4 交付物
- MultiServerMCPClient 完整实现
- mtime 缓存失效
- OAuth 支持（两种流程）
- MCP 工具集成
- 单元测试覆盖率 ≥ 75%
- `docs/design/042-5-Phase4-MCP补齐总结.md`
- **完成度提升**：MCP 系统 30% → **90%**，总体 82% → **88%**

---

### Phase 5: 收尾与集成测试（2-3 天）- **P2 低优先级**

#### 目标
端到端测试、文档更新、已知限制整理。

#### 任务清单

| 任务 | 说明 |
|------|------|
| 5.1 | Lead Agent 端到端集成测试（使用 mock 模型） |
| 5.2 | 子代理端到端集成测试 |
| 5.3 | 示例 config.yaml 完整示例 |
| 5.4 | 更新 AGENTS.md / README.md |
| 5.5 | 整理已知限制与待改进点 |
| 5.6 | 最终代码走查与清理 |

#### Phase 5 交付物
- 端到端集成测试
- 完整配置示例
- 功能总结文档
- **总体完成度**：88% → **90%+**

---

## 四、总体时间估算

| 阶段 | 估算时间 | 依赖 | 完成度提升 |
|------|----------|------|-----------|
| Phase 0: 准备 | 1 天 | - | - |
| Phase 1: 中间件补齐 | 5-7 天 | Phase 0 | +12% (58%→70%) |
| Phase 2: Memory 系统 | 2-3 天 | Phase 1 可并行 | +5% (70%→75%) |
| Phase 3: 工具补齐 | 3-4 天 | Phase 1 | +7% (75%→82%) |
| Phase 4: MCP 对齐 | 3-4 天 | Phase 1 | +6% (82%→88%) |
| Phase 5: 收尾 | 2-3 天 | Phase 1-4 | +2% (88%→90%) |
| **总计** | **16-21 天** | - | **+32%** |

---

## 五、风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| Eino 框架 API 变更 | 高 | 中 | 锁定 eino 版本，先做 API 兼容性测试 |
| Memory LLM 提取效果不确定 | 中 | 低 | 先用 prompt 对齐，后续可迭代优化 |
| 社区工具 API 变更 | 中 | 低 | 使用抽象接口，易于替换 |
| 时间估算不足 | 中 | 中 | 分阶段交付，优先保证 P0 功能 |

---

## 六、验收标准

### 整体验收

1. **编译通过**：`go build ./...` 无错误
2. **测试通过**：`go test ./... -cover` 覆盖率 ≥ 75%
3. **端到端跑通**：Lead Agent 可以完成简单任务（读文件、写文件、bash）
4. **文档完整**：各模块有使用说明，进度文档更新到最新

### 各阶段验收

每个阶段完成后需：
1. 该阶段所有任务打勾 ✅
2. 新增单元测试通过
3. 编写阶段总结文档
4. 代码 review 通过（如需要）

---

## 七、后续优化方向（v2）

达到 90% 完成度后，可考虑：

1. **性能优化**：中间件链缓存、Memory 批量更新
2. **可观测性**：结构化日志、Trace 集成
3. **高级功能**：Checkpointer、Agents 配置持久化
4. **前端深度对齐**：WebSocket 事件、Artifact 展示

---

**文档结束**

下一步：执行 Phase 0（准备阶段）
