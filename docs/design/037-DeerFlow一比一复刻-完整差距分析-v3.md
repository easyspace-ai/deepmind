# DeerFlow 一比一复刻 - 完整差距分析 v3

> **校准**：v3 中部分条目（如「中间件皆占位」「配置仅 paths」）已随 v4–v6 过期。请以 [038 v4 实施进度](./038-DeerFlow一比一复刻-实施进度-v4.md)、[040 v6 实施进度](./040-DeerFlow一比一复刻-实施进度-v6.md) 与 [041 对齐计划补充差距切片](./041-DeerFlow一比一复刻-差距分析-对齐计划补充.md) 为准。

## 概述

本文档详细对比 nanobot-go 当前实现与 DeerFlow 原版的完整差距。

## 已完成的工作

### ✅ 1. 状态系统 (pkg/agent/state/)
- **types.go**: ThreadState, SandboxState, ThreadDataState, ViewedImageData, TodoItem, UploadedFile
- **reducers.go**: MergeArtifacts, MergeViewedImages, 20+ 辅助方法
- **state_test.go**: 21 个单元测试 ✅

### ✅ 2. 提示词系统 (pkg/agent/prompts/)
- **types.go**: PromptSection 接口, BaseSection, NamedSection, Prompt 容器
- **sections.go**: 11 个提示词分段完整实现
- **builder.go**: Builder 模式 + 预设函数
- **prompts_test.go**: 34 个单元测试 ✅

### ✅ 3. Sandbox 系统 (pkg/sandbox/)
- **types.go**: Sandbox 接口, SandboxProvider 接口
- **path.go**: PathTranslator, 路径翻译/验证/掩码
- **local/local_sandbox.go**: LocalSandbox 实现
- **sandbox_test.go, local_sandbox_test.go**: 8 个单元测试 ✅

### ✅ 4. 中间件系统 (pkg/agent/middleware/)
- **types.go**: Middleware 接口, MiddlewareChain, EinoCallbackBridge
- **chain.go**: BuildLeadAgentMiddlewares(), 14 个中间件严格顺序
- **loop_detection.go**: LoopDetectionMiddleware 完整实现
- **thread_data_middleware.go, uploads_middleware.go, title_middleware.go, dangling_tool_call_middleware.go, view_image_middleware.go, clarification_middleware.go, sandbox_middleware.go, subagent_limit_middleware.go, summarization_middleware.go, todo_list_middleware.go, memory_middleware.go, tool_error_handling_middleware.go, deferred_tool_filter_middleware.go**: 所有 14 个中间件的占位符/部分实现
- **middleware_test.go**: 4 个单元测试 ✅

### ✅ 5. 工具系统 (pkg/agent/tools/deerflow/)
- **types.go**: 工具类型和工厂函数
- **tool_security.go**: 工具安全层（路径验证、翻译、掩码）
- **sandbox_tools.go**: 5 个 Sandbox 工具
- **builtin_tools.go**: 3 个 Built-in 工具
- **task_tool.go**: task() 工具实现 ✅

### ✅ 6. 子代理系统 (pkg/agent/subagent/)
- **types.go**: SubagentStatus, SubagentConfig, SubagentResult, 全局任务存储
- **executor.go**: SubagentExecutor, WorkerPool, 双线程池设计
- **events.go**: TaskEventBus, 事件系统
- **builtins.go**: general-purpose, bash 内置配置
- **registry.go**: SubagentRegistry, 全局注册表
- **subagent_test.go**: 21 个单元测试 ✅

### ✅ 7. Lead Agent 工厂 (pkg/agent/factory/)
- **factory.go**: LeadAgentConfig, LeadAgentFactory, MakeLeadAgent()
- **factory_test.go**: 11 个单元测试 ✅

---

## 剩余差距

### 1. 中间件完整业务逻辑实现 ⚠️

当前所有中间件都是占位符或部分实现，缺少完整的业务逻辑：

| 中间件 | 状态 | 缺失内容 |
|--------|------|----------|
| **ThreadDataMiddleware** | 占位符 | - 创建线程目录结构<br>- 计算 workspace/uploads/outputs 路径<br>- lazy_init 支持<br>- 与 Paths 集成 |
| **UploadsMiddleware** | 占位符 | - 从 message.additional_kwargs.files 提取新上传文件<br>- 扫描历史上传文件<br>- 生成 &lt;uploaded_files&gt; 消息块<br>- 预处理 HumanMessage 内容 |
| **SandboxMiddleware** | 占位符 | - 从 SandboxProvider acquire sandbox<br>- 存储 sandbox_id 到 state<br>- lazy_init 支持<br>- 与 pkg/sandbox 集成 |
| **DanglingToolCallMiddleware** | 占位符 | - 检测 AIMessage.tool_calls 缺失对应 ToolMessage<br>- wrap_model_call / awrap_model_call 实现<br>- 插入 synthetic ToolMessage<br>- 使用正确的 message ID |
| **SummarizationMiddleware** | 占位符 | - token/message 触发检测<br>- 模型调用生成摘要<br>- keep 策略实现<br>- 与 config 集成 |
| **TodoListMiddleware** | 占位符 | - write_todos 工具注入<br>- 系统提示词注入<br>- 实时 todo 状态管理<br>- is_plan_mode 条件启用 |
| **TitleMiddleware** | 占位符 | - 首次消息交换检测<br>- 标题模型调用<br>- 规范化结构化内容<br>- 标题生成提示词 |
| **MemoryMiddleware** | 占位符 | - 消息过滤（用户输入 + 最终 AI 响应）<br>- 去抖队列<br>- 异步内存更新<br>- 与 pkg/agent/memory 集成 |
| **ViewImageMiddleware** | 占位符 | - 图像文件读取<br>- base64 编码<br>- 注入 viewed_images 到 state<br>- 视觉支持模型条件判断 |
| **SubagentLimitMiddleware** | 占位符 | - task() 工具调用计数<br>- 超出限制时截断<br>- after_model / aafter_model 实现<br>- 2-4 范围限制 |
| **LoopDetectionMiddleware** | ⚠️ 部分 | - order-independent 哈希完整实现<br>- LRU 缓存<br>- 警告/硬停止阈值<br>- 当前只有基础框架 |
| **ToolErrorHandlingMiddleware** | 占位符 | - wrap_tool_call / awrap_tool_call 实现<br>- 异常捕获与转换<br>- 生成 ToolMessage (status="error")<br>- GraphBubbleUp 透传 |
| **DeferredToolFilterMiddleware** | 占位符 | - 延迟工具 schema 隐藏<br>- tool_search 启用条件<br>- &lt;available-deferred-tools&gt; 注入 |
| **ClarificationMiddleware** | 占位符 | - ask_clarification 工具调用拦截<br>- Command(goto=END) 中断<br>- 必须在最后位置 |

### 2. Eino 框架集成 ⚠️

**中间件 Eino 适配**:
- 当前中间件接口设计为兼容 Eino，但实际的桥接实现不完整
- `EinoCallbackBridge` 需要完整实现所有 callbacks
- 需要与 `github.com/cloudwego/eino/schema` 类型完全对齐

**SubagentExecutor Eino 集成**:
- `SubagentExecutor.createAgent()` 是占位符
- 需要使用 Eino 的 `DeepAgent` 或 `ChatModelAgent`
- 需要正确的 state 传递与中间件集成

### 3. 提示词系统完整性 ⚠️

**缺失的提示词分段**:
- **Subagent 分段**: 完整的 &lt;subagent_system&gt; 分段（148 行）
  - 并行分解指导
  - 硬并发限制说明
  - 多批次执行策略
  - 详细的示例代码

- **Skills 分段**: &lt;skill_system&gt; 分段
  - 渐进式加载模式
  - 可用技能列表
  - 容器路径注入

- **Deferred Tools 分段**: &lt;available-deferred-tools&gt; 分段
  - 工具搜索启用条件
  - 延迟工具名称列表

- **Soul 分段**: &lt;soul&gt; 分段
  - Agent 个性注入
  - SOUL.md 加载

- **Memory 分段**: &lt;memory&gt; 分段
  - 记忆数据格式化
  - token 限制注入

- **Current Date**: &lt;current_date&gt; 标签

**提示词模板系统**:
- 需要完整的 `SYSTEM_PROMPT_TEMPLATE`（DeerFlow 有 335 行）
- 动态变量替换: `agent_name`, `soul`, `skills_section`, `deferred_tools_section`, `memory_context`, `subagent_section`, `subagent_reminder`, `subagent_thinking`

### 4. 内存系统 (Memory) ❌

完全缺失，需要创建:
```
pkg/agent/memory/
├── types.go       - MemoryData, UserContext, History, Fact
├── updater.go     - LLM 事实提取，去重，原子写入
├── queue.go       - 去抖队列 (debounce)
├── prompt.go      - 内存更新提示词
└── memory_test.go
```

**功能需求**:
- UserContext: workContext, personalContext, topOfMind
- History: recentMonths, earlierContext, longTermBackground
- Facts: id, content, category, confidence, createdAt, source
- 去抖更新队列 (30s 默认)
- 空格规范化事实去重
- 原子文件 I/O (temp file + rename)

### 5. Skills 系统 ❌

完全缺失，需要创建:
```
pkg/agent/skills/
├── types.go       - Skill, SkillMetadata
├── loader.go      - load_skills(), 递归扫描
├── parser.go      - SKILL.md 解析 (YAML frontmatter)
└── skills_test.go
```

**功能需求**:
- SKILL.md 格式: YAML frontmatter + 内容
- 目录扫描: `skills/{public,custom}/`
- enabled 状态管理
- 容器路径注入
- `get_skills_prompt_section()` 函数

### 6. 配置系统 ⚠️

**当前状态**: 只有 `pkg/config/paths.go`

**缺失的配置模块**:
```
pkg/config/
├── app_config.go      - AppConfig (models, tools, sandbox, skills, title, summarization, subagents, memory)
├── model_config.go    - ModelConfig (use, supports_thinking, supports_vision)
├── sandbox_config.go  - SandboxConfig (use)
├── skills_config.go   - SkillsConfig (path, container_path)
├── title_config.go    - TitleConfig (enabled, max_words, max_chars, prompt_template)
├── summarization_config.go - SummarizationConfig (enabled, trigger, keep, model_name, trim_tokens_to_summarize, summary_prompt)
├── subagents_config.go - SubagentsConfig (enabled, timeouts map)
├── memory_config.go   - MemoryConfig (enabled, storage_path, debounce_seconds, model_name, max_facts, fact_confidence_threshold, injection_enabled, max_injection_tokens)
├── loader.go          - FromFile(), 环境变量解析 ($VAR), config_version 检查
└── config_test.go
```

**config.yaml 完整 schema**:
```yaml
config_version: 1
models: []
tools: []
tool_groups: []
sandbox:
  use: ...
skills:
  path: ...
  container_path: ...
title:
  enabled: true
  ...
summarization:
  enabled: false
  ...
subagents:
  enabled: true
  timeouts: {...}
memory:
  enabled: false
  ...
```

### 7. 工具系统完整性 ⚠️

**缺失的工具**:
- **tool_search 工具**: 搜索延迟工具
- **write_todos 工具**: 任务列表管理 (TodoListMiddleware)
- **社区工具**: tavily, jina_ai, firecrawl, image_search

**工具安全层**:
- 当前有基础框架，但需要与实际工具调用集成
- `ValidateLocalToolPath()` 需要完整实现
- `MaskLocalPathsInOutput()` 需要完整实现

### 8. Deferred Tool 系统 ❌

完全缺失，需要:
- DeferredToolRegistry
- tool_search 工具实现
- `get_deferred_registry()` 函数
- `DeferredToolFilterMiddleware` 完整实现

### 9. MCP 系统 ❌

完全缺失，需要:
```
pkg/agent/mcp/
├── types.go       - MCPServerConfig
├── client.go      - MultiServerMCPClient
├── tools.go       - get_cached_mcp_tools(), mtime 缓存失效
└── cache.go
```

**功能需求**:
- stdio, SSE, HTTP transports
- OAuth 支持 (client_credentials, refresh_token)
- 自动 token 刷新
- 懒加载 + mtime 缓存失效

### 10. 模型工厂 ❌

完全缺失，需要:
```
pkg/models/
├── factory.go     - create_chat_model(name, thinking_enabled)
├── types.go       - ModelConfig
└── reflection.go  - resolve_variable(), resolve_class()
```

**功能需求**:
- 反射加载 provider 模块
- thinking_enabled 支持
- supports_vision 检测
- per-model when_thinking_enabled 覆盖
- 环境变量解析 ($VAR)
- 缺失模块时提供安装提示

### 11. 完整的工具链构建 ❌

DeerFlow 的 `get_available_tools()` 组合:
1. Config-defined tools
2. MCP tools (懒加载, mtime 缓存)
3. Built-in tools (present_files, ask_clarification, view_image)
4. Subagent tool (task, 如果 enabled)
5. Community tools (tavily, jina_ai, firecrawl, image_search)

当前只有基础框架，缺少完整的组装逻辑。

### 12. ThreadState 的完整 Reducers ⚠️

当前只有 MergeArtifacts, MergeViewedImages，需要确认:
- messages reducer
- sandbox reducer
- thread_data reducer
- title reducer
- todos reducer
- uploaded_files reducer
- viewed_images reducer (已有)
- artifacts reducer (已有)

### 13. 端到端集成测试 ❌

完全缺失，需要:
- 完整的 Agent 创建流程测试
- 中间件链执行测试
- 工具调用测试
- 子代理执行测试

### 14. 前端集成 ⚠️

需要确认:
- WebSocket 事件流 (task_started, task_running, task_completed 等)
- ThreadState 与前端状态的对齐
- Artifact 展示
- TodoList 实时更新

---

## 优先级建议

### P0 - 核心功能 (必须有)
1. **中间件完整业务逻辑** - 14 个中间件都需要完整实现
2. **Eino 框架深度集成** - 中间件桥接 + SubagentExecutor
3. **提示词系统完整性** - Subagent/Skills/Memory/Soul 分段
4. **配置系统** - AppConfig + 所有子配置 + config.yaml 加载

### P1 - 重要功能 (应该有)
5. **内存系统** - MemoryMiddleware + updater + queue
6. **Skills 系统** - 技能加载 + 提示词注入
7. **工具系统完整性** - tool_search, write_todos, 社区工具
8. **模型工厂** - create_chat_model() + 反射加载

### P2 - 增强功能 (可以有)
9. **MCP 系统** - 多服务器管理 + OAuth
10. **Deferred Tool 系统** - 延迟工具 + tool_search
11. **端到端集成测试**
12. **前端集成** - WebSocket 事件 + 状态对齐

---

## 总结

### 已完成度估算

| 模块 | 完成度 |
|------|--------|
| 状态系统 | 95% |
| 提示词系统 | 60% (缺少 6 个分段) |
| Sandbox 系统 | 70% (缺少中间件集成) |
| 中间件系统 | 30% (都是占位符) |
| 工具系统 | 50% (缺少很多工具) |
| 子代理系统 | 80% (缺少 Eino 集成) |
| Lead Agent 工厂 | 70% (缺少完整集成) |
| 内存系统 | 0% |
| Skills 系统 | 0% |
| 配置系统 | 20% |
| MCP 系统 | 0% |
| 模型工厂 | 0% |

**总体完成度: ~40%**

### 关键差距

1. **中间件业务逻辑**: 所有 14 个中间件都需要从占位符变为完整实现
2. **提示词完整性**: 缺少 6 个重要的提示词分段
3. **Eino 集成**: 需要深度集成 Eino 框架的实际类型
4. **配置系统**: 需要完整的 config.yaml 加载和所有子配置
5. **内存系统**: 完全缺失
6. **Skills 系统**: 完全缺失

### 建议

如果目标是"一比一复刻，达到企业级可用"，建议按以下顺序:
1. **先完成所有中间件的完整业务逻辑**
2. **完成提示词系统的所有分段**
3. **深度集成 Eino 框架**
4. **实现完整的配置系统**
5. **实现内存系统**
6. **实现 Skills 系统**

这样可以达到 ~80-85% 的完成度，基本可用。
