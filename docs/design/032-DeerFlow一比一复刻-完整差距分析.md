# DeerFlow 一比一复刻 - 完整差距分析

## 概述

本文档基于对 DeerFlow 完整源码的深入分析，对比当前 nanobot-go 实现与原版 DeerFlow 的详细差距。

---

## 一、完整差距对比表

### 1.1 ThreadState 状态系统 ✅

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| `sandbox` 字段 | ✅ SandboxState | ✅ SandboxState | ✅ 完成 |
| `thread_data` 字段 | ✅ ThreadDataState | ✅ ThreadDataState | ✅ 完成 |
| `title` 字段 | ✅ string | ✅ string | ✅ 完成 |
| `artifacts` 字段 | ✅ []string + MergeArtifacts reducer | ✅ []string + MergeArtifacts reducer | ✅ 完成 |
| `todos` 字段 | ✅ []TodoItem | ✅ []TodoItem | ✅ 完成 |
| `uploaded_files` 字段 | ✅ []UploadedFile | ✅ []UploadedFile | ✅ 完成 |
| `viewed_images` 字段 | ✅ map[string]ViewedImageData + MergeViewedImages reducer | ✅ map[string]ViewedImageData + MergeViewedImages reducer | ✅ 完成 |
| 辅助方法（20+） | ✅ | ✅ | ✅ 完成 |
| 单元测试 | 未统计 | 21 个 ✅ | ✅ 完成 |

**结论**：ThreadState 状态系统**100% 完成**，一比一对齐。

---

### 1.2 模块化提示词系统 ✅

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| PromptSection 接口 | ✅ | ✅ | ✅ 完成 |
| RoleSection | ✅ | ✅ | ✅ 完成 |
| ThinkingStyleSection | ✅ | ✅ | ✅ 完成 |
| ClarificationSection | ✅ | ✅ | ✅ 完成 |
| SubagentSection | ✅ | ✅ | ✅ 完成 |
| WorkingDirSection | ✅ | ✅ | ✅ 完成 |
| SkillsSection | ✅ | ✅ | ✅ 完成 |
| EnvInfoSection | ✅ | ✅ | ✅ 完成 |
| ResponseStyleSection | ✅ | ✅ | ✅ 完成 |
| CitationsSection | ✅ | ✅ | ✅ 完成 |
| CriticalRemindersSection | ✅ | ✅ | ✅ 完成 |
| CustomSection | ✅ | ✅ | ✅ 完成 |
| Builder 模式 | ✅ | ✅ | ✅ 完成 |
| BuildLeadAgentPrompt() | ✅ | ✅ | ✅ 完成 |
| BuildGeneralPurposeSubagentPrompt() | ✅ | ✅ | ✅ 完成 |
| BuildBashSubagentPrompt() | ✅ | ✅ | ✅ 完成 |
| 单元测试 | 未统计 | 34 个 ✅ | ✅ 完成 |

**结论**：模块化提示词系统**100% 完成**，11 个分段全部实现，一比一对齐。

---

### 1.3 Sandbox 虚拟路径系统 ✅

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| Sandbox 接口 | ✅ ExecuteCommand, ReadFile, WriteFile, ListDir, UpdateFile | ✅ ExecuteCommand, ReadFile, WriteFile, ListDir, UpdateFile | ✅ 完成 |
| SandboxProvider 接口 | ✅ Acquire, Get, Release | ✅ Acquire, Get, Release | ✅ 完成 |
| 虚拟路径系统 | ✅ /mnt/user-data/{workspace,uploads,outputs} | ✅ /mnt/user-data/{workspace,uploads,outputs} | ✅ 完成 |
| PathTranslator.ToPhysical() | ✅ replace_virtual_path() | ✅ ToPhysical() | ✅ 完成 |
| PathTranslator.ToVirtual() | ✅ mask_local_paths_in_output() | ✅ ToVirtual() | ✅ 完成 |
| PathTranslator.ValidatePath() | ✅ validate_local_tool_path() + _reject_path_traversal() | ✅ ValidatePath() | ✅ 完成 |
| PathTranslator.TranslatePathsInCommand() | ✅ replace_virtual_paths_in_command() | ✅ TranslatePathsInCommand() | ✅ 完成 |
| PathTranslator.MaskPathsInOutput() | ✅ mask_local_paths_in_output() | ✅ MaskPathsInOutput() | ✅ 完成 |
| LocalSandbox 实现 | ✅ | ✅ | ✅ 完成 |
| LocalSandboxProvider 单例 | ✅ | ✅ | ✅ 完成 |
| /mnt/skills 路径支持 | ✅ | ❌ 缺失 | ⚠️ 部分完成 |
| 单元测试 | 未统计 | 8 个 ✅ | ✅ 完成 |

**结论**：Sandbox 虚拟路径系统**95% 完成**，核心功能已对齐，仅缺少 `/mnt/skills` 路径支持（低优先级）。

---

### 1.4 中间件链系统 🔄

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| Middleware 接口框架 | ✅ | ✅ | ✅ 完成 |
| MiddlewareChain | ✅ | ✅ | ✅ 完成 |
| EinoCallbackBridge | ✅ Eino 回调桥接器 | ✅ | ✅ 完成 |
| **ThreadDataMiddleware** | ✅ 创建线程目录 | ❌ 仅框架 | ⏳ 待实现 |
| **UploadsMiddleware** | ✅ 上传文件追踪 | ❌ 仅框架 | ⏳ 待实现 |
| **DanglingToolCallMiddleware** | ✅ 处理挂起工具调用 | ❌ 仅框架 | ⏳ 待实现 |
| **SummarizationMiddleware** | ✅ 上下文摘要（可选） | ❌ 仅框架 | ⏳ 待实现 |
| **TodoListMiddleware** | ✅ 任务追踪（可选） | ❌ 仅框架 | ⏳ 待实现 |
| **TitleMiddleware** | ✅ 自动生成标题 | ❌ 仅框架 | ⏳ 待实现 |
| **MemoryMiddleware** | ✅ 记忆提取（可选） | ❌ 仅框架 | ⏳ 待实现 |
| **ViewImageMiddleware** | ✅ 图像处理 | ❌ 仅框架 | ⏳ 待实现 |
| **SubagentLimitMiddleware** | ✅ 限制并发（可选） | ❌ 仅框架 | ⏳ 待实现 |
| **LoopDetectionMiddleware** | ✅ 循环检测 | ✅ 框架实现（无完整逻辑） | ⚠️ 部分完成 |
| **ClarificationMiddleware** | ✅ 询问澄清（必须最后！） | ❌ 仅框架 | ⏳ 待实现 |
| **SandboxMiddleware** | ✅ 沙箱生命周期 | ❌ 完全缺失 | ⏳ 待实现 |
| **ToolErrorHandlingMiddleware** | ✅ 工具错误处理 | ❌ 完全缺失 | ⏳ 待实现 |
| **DeferredToolFilterMiddleware** | ✅ 延迟工具过滤 | ❌ 完全缺失 | ⏳ 待实现 |
| 严格执行顺序 | ✅ 固定 12-13 个中间件顺序 | ✅ 框架支持 | ✅ 完成 |
| 单元测试 | 未统计 | 4 个 ✅ | ✅ 完成 |

**关键发现**：DeerFlow 实际有 **13 个中间件**（而非之前认为的 11 个）！

**结论**：中间件链系统**20% 完成**，仅框架和 LoopDetection 框架已完成，**12 个完整中间件逻辑待实现**。

---

### 1.5 DeerFlow 风格工具系统 🔄

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| 工具分组框架 | ✅ sandbox, builtin, mcp, community, subagent | ✅ | ✅ 完成 |
| GetAvailableTools() | ✅ | ✅ | ✅ 完成 |
| BuildLeadAgentTools() | ✅ | ✅ | ✅ 完成 |
| BuildGeneralPurposeSubagentTools() | ✅ | ✅ | ✅ 完成 |
| BuildBashSubagentTools() | ✅ | ✅ | ✅ 完成 |
| **bash 工具** | ✅ 完整实现（带路径翻译） | ❌ 占位符 | ⏳ 待实现 |
| **ls 工具** | ✅ 树状格式，最多 2 层 | ❌ 占位符 | ⏳ 待实现 |
| **read_file 工具** | ✅ 支持行范围 | ❌ 占位符 | ⏳ 待实现 |
| **write_file 工具** | ✅ 支持 append | ❌ 占位符 | ⏳ 待实现 |
| **str_replace 工具** | ✅ 单处或全部替换 | ❌ 占位符 | ⏳ 待实现 |
| **present_files 工具** | ✅ 展示输出文件 | ❌ 占位符 | ⏳ 待实现 |
| **ask_clarification 工具** | ✅ 询问澄清 | ❌ 占位符 | ⏳ 待实现 |
| **view_image 工具** | ✅ 查看图像 | ❌ 占位符 | ⏳ 待实现 |
| **task 工具** | ✅ 子代理委托 | ❌ 占位符 | ⏳ 待实现 |
| **MCP 工具** | ✅ 完整实现 | ❌ 完全缺失 | ⏳ 待实现 |
| **社区工具** | ✅ tavily, jina_ai, firecrawl, image_search | ❌ 完全缺失 | ⏳ 待实现 |
| **write_todos 工具** | ✅ 任务列表管理 | ❌ 完全缺失 | ⏳ 待实现 |
| 工具路径安全验证 | ✅ validate_local_tool_path() | ❌ 占位符无 | ⏳ 待实现 |
| 工具路径翻译 | ✅ replace_virtual_paths_in_command() | ❌ 占位符无 | ⏳ 待实现 |
| 工具路径掩码 | ✅ mask_local_paths_in_output() | ❌ 占位符无 | ⏳ 待实现 |
| 工具懒加载沙箱 | ✅ ensure_sandbox_initialized() | ❌ 占位符无 | ⏳ 待实现 |
| 工具错误处理 | ✅ _sanitize_error() | ❌ 占位符无 | ⏳ 待实现 |

**结论**：DeerFlow 风格工具系统**15% 完成**，仅框架和工厂函数已完成，**13+ 个完整工具实现待编写**。

---

### 1.6 子代理系统 ⏳

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| SubagentConfig 类型 | ✅ | ❌ 缺失 | ⏳ 待创建 |
| SubagentStatus 枚举 | ✅ PENDING, RUNNING, COMPLETED, FAILED, TIMED_OUT | ❌ 缺失 | ⏳ 待创建 |
| SubagentResult 类型 | ✅ task_id, trace_id, status, result, error, started_at, completed_at, ai_messages | ❌ 缺失 | ⏳ 待创建 |
| **SubagentExecutor** | ✅ 完整核心执行器 | ❌ 缺失 | ⏳ 待创建 |
| 双线程池设计 | ✅ _scheduler_pool (3) + _execution_pool (3) | ❌ 缺失 | ⏳ 待实现 |
| MAX_CONCURRENT_SUBAGENTS | ✅ =3 | ❌ 缺失 | ⏳ 待实现 |
| timeout_seconds | ✅ =900 (15分钟) | ❌ 缺失 | ⏳ 待实现 |
| SubagentExecutor.execute() | ✅ 同步执行 | ❌ 缺失 | ⏳ 待实现 |
| SubagentExecutor._aexecute() | ✅ 异步执行（带流式消息捕获） | ❌ 缺失 | ⏳ 待实现 |
| SubagentExecutor.execute_async() | ✅ 后台执行（返回 task_id） | ❌ 缺失 | ⏳ 待实现 |
| 工具过滤系统 | ✅ _filter_tools() - 移除 task 防止嵌套 | ❌ 缺失 | ⏳ 待实现 |
| 模型继承系统 | ✅ _get_model_name() - "inherit" 支持 | ❌ 缺失 | ⏳ 待实现 |
| 状态继承系统 | ✅ sandbox_state, thread_data 继承 | ❌ 缺失 | ⏳ 待实现 |
| 背景任务存储 | ✅ _background_tasks dict | ❌ 缺失 | ⏳ 待实现 |
| get_background_task_result() | ✅ | ❌ 缺失 | ⏳ 待实现 |
| list_background_tasks() | ✅ | ❌ 缺失 | ⏳ 待实现 |
| cleanup_background_task() | ✅ 仅清理终端状态任务 | ❌ 缺失 | ⏳ 待实现 |
| 子代理注册系统 | ✅ general-purpose, bash | ❌ 缺失 | ⏳ 待实现 |
| **事件系统** | ✅ task_started, task_running, task_completed, task_failed, task_timed_out | ❌ 缺失 | ⏳ 待实现 |
| 实时 AI 消息流 | ✅ 流式捕获子代理 AI 消息 | ❌ 缺失 | ⏳ 待实现 |
| task() 工具轮询逻辑 | ✅ 5 秒轮询，超时保护 | ❌ 缺失 | ⏳ 待实现 |
| trace_id 分布式追踪 | ✅ 父子链路追踪 | ❌ 缺失 | ⏳ 待实现 |

**结论**：子代理系统**0% 完成**，**完整系统待创建**。

---

### 1.7 配置系统 ⏳

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| AppConfig | ✅ 完整应用配置 | ❌ 缺失 | ⏳ 待创建 |
| ModelConfig | ✅ 模型工厂配置 | ❌ 缺失 | ⏳ 待创建 |
| SandboxConfig | ✅ 沙箱配置 | ❌ 缺失 | ⏳ 待创建 |
| SkillsConfig | ✅ 技能配置 | ❌ 缺失 | ⏳ 待创建 |
| SubagentsConfig | ✅ 子代理配置 | ❌ 缺失 | ⏳ 待创建 |
| TitleConfig | ✅ 标题生成配置 | ❌ 缺失 | ⏳ 待创建 |
| SummarizationConfig | ✅ 摘要配置 | ❌ 缺失 | ⏳ 待创建 |
| MemoryConfig | ✅ 记忆配置 | ❌ 缺失 | ⏳ 待创建 |
| ToolConfig | ✅ 工具配置 | ❌ 缺失 | ⏳ 待创建 |
| CheckpointerConfig | ✅ Checkpoint 配置 | ❌ 缺失 | ⏳ 待创建 |
| TracingConfig | ✅ 追踪配置 | ❌ 缺失 | ⏳ 待实现 |
| config.yaml 加载 | ✅ 优先级解析 + 环境变量替换 | ❌ 缺失 | ⏳ 待实现 |
| config_version 检查 | ✅ 版本检查 + 升级警告 | ❌ 缺失 | ⏳ 待实现 |
| extensions_config.json | ✅ MCP + Skills 配置 | ❌ 缺失 | ⏳ 待实现 |

**结论**：配置系统**0% 完成**，**完整系统待创建**。

---

### 1.8 Skills 系统 ⏳

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| Skills 目录结构 | ✅ skills/{public,custom}/ | ❌ 缺失 | ⏳ 待创建 |
| SKILL.md 格式 | ✅ YAML frontmatter + 内容 | ❌ 缺失 | ⏳ 待实现 |
| load_skills() | ✅ 递归扫描加载 | ❌ 缺失 | ⏳ 待实现 |
| Skills 注入提示词 | ✅ /mnt/skills 路径列表 | ❌ 缺失 | ⏳ 待实现 |
| Skills API | ✅ list, get, update, install | ❌ 缺失 | ⏳ 待实现 |

**结论**：Skills 系统**0% 完成**，但**低优先级**（可后期集成）。

---

### 1.9 记忆系统 ⏳

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| MemoryUpdater | ✅ LLM 事实提取 | ❌ 缺失 | ⏳ 待创建 |
| MemoryQueue | ✅ 防抖更新队列 | ❌ 缺失 | ⏳ 待创建 |
| memory.json 存储 | ✅ userContext, personalContext, topOfMind, history, facts | ❌ 缺失 | ⏳ 待实现 |
| 事实去重 | ✅ 空白标准化 + 内容比较 | ❌ 缺失 | ⏳ 待实现 |
| 原子文件 I/O | ✅ 临时文件 + 重命名 | ❌ 缺失 | ⏳ 待实现 |
| 记忆注入提示词 | ✅ <memory> 标签 | ❌ 缺失 | ⏳ 待实现 |

**结论**：记忆系统**0% 完成**，但**低优先级**（可后期集成）。

---

### 1.10 MCP 系统 ⏳

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| MultiServerMCPClient | ✅ 多服务器管理 | ❌ 缺失 | ⏳ 待集成 |
| MCP 工具懒加载 | ✅ 首次使用初始化 | ❌ 缺失 | ⏳ 待实现 |
| MCP 缓存失效 | ✅ mtime 比较 | ❌ 缺失 | ⏳ 待实现 |
| stdio 传输 | ✅ | ❌ 缺失 | ⏳ 待实现 |
| SSE/HTTP 传输 | ✅ | ❌ 缺失 | ⏳ 待实现 |
| OAuth 支持 | ✅ client_credentials, refresh_token | ❌ 缺失 | ⏳ 待实现 |
| 自动 Token 刷新 | ✅ | ❌ 缺失 | ⏳ 待实现 |
| MCP API | ✅ get_config, update_config | ❌ 缺失 | ⏳ 待实现 |

**结论**：MCP 系统**0% 完成**，但**低优先级**（可后期集成）。

---

### 1.11 Lead Agent 工厂 ⏳

| 特性 | DeerFlow 原版 | nanobot-go 当前实现 | 状态 |
|------|--------------|---------------------|------|
| make_lead_agent() | ✅ 工厂函数 | ❌ 缺失 | ⏳ 待创建 |
| 中间件链组合 | ✅ 严格 13 个中间件顺序 | ❌ 缺失 | ⏳ 待实现 |
| 动态工具加载 | ✅ get_available_tools() | ❌ 缺失 | ⏳ 待实现 |
| 系统提示词构建 | ✅ apply_prompt_template() | ❌ 缺失 | ⏳ 待实现 |
| 可配置运行时 | ✅ thinking_enabled, model_name, is_plan_mode, subagent_enabled | ❌ 缺失 | ⏳ 待实现 |

**结论**：Lead Agent 工厂**0% 完成**，**核心集成待实现**。

---

## 二、完整中间件列表（DeerFlow 原版 13 个）

### 中间件执行顺序（严格！）

```
1.  ThreadDataMiddleware         - 创建线程目录
2.  UploadsMiddleware             - 上传文件追踪
3.  SandboxMiddleware             - 沙箱生命周期管理（新增！）
4.  DanglingToolCallMiddleware    - 处理挂起工具调用
5.  SummarizationMiddleware       - 上下文摘要（可选）
6.  TodoListMiddleware            - 任务追踪（可选）
7.  TitleMiddleware               - 自动生成标题
8.  MemoryMiddleware              - 记忆提取（可选）
9.  ViewImageMiddleware           - 图像处理
10. SubagentLimitMiddleware       - 限制并发（可选）
11. LoopDetectionMiddleware       - 循环检测
12. ToolErrorHandlingMiddleware   - 工具错误处理（新增！）
13. DeferredToolFilterMiddleware  - 延迟工具过滤（新增！）
14. ClarificationMiddleware       - 询问澄清（必须最后！）
```

**关键发现**：原版 DeerFlow 有 **14 个中间件**，其中 3 个之前未计入：
- SandboxMiddleware
- ToolErrorHandlingMiddleware
- DeferredToolFilterMiddleware

---

## 三、完整工具列表（DeerFlow 原版 15+ 个）

### 3.1 Sandbox 工具（5 个）

| 工具 | 描述 | 关键特性 |
|------|------|---------|
| `bash` | 执行 bash 命令 | 路径验证、路径翻译、路径掩码、错误处理 |
| `ls` | 树状目录列表 | 最多 2 层、路径验证、skills 只读 |
| `read_file` | 读取文件 | 支持行范围、路径验证 |
| `write_file` | 写入文件 | 支持 append、路径验证、自动创建目录 |
| `str_replace` | 字符串替换 | 单处/全部替换、验证 old_str 存在 |

### 3.2 Built-in 工具（5 个）

| 工具 | 描述 | 关键特性 |
|------|------|---------|
| `present_files` | 展示输出文件 | 仅 /mnt/user-data/outputs、路径规范化、artifacts reducer |
| `ask_clarification` | 询问澄清 | 5 种类型、options 支持、被 ClarificationMiddleware 拦截 |
| `view_image` | 查看图像 | base64 转换、viewed_images reducer、格式验证 |
| `write_todos` | 任务列表 | 1 个 in_progress、TodoMiddleware 上下文丢失检测 |

### 3.3 Subagent 工具（1 个）

| 工具 | 描述 | 关键特性 |
|------|------|---------|
| `task` | 子代理委托 | 双线程池、5s 轮询、实时事件、trace_id 追踪、工具过滤 |

### 3.4 MCP 工具（动态）

| 来源 | 描述 |
|------|------|
| MCP 服务器 | 动态加载、缓存失效 |

### 3.5 Community 工具（4 个）

| 工具 | 描述 |
|------|------|
| `tavily` | 网络搜索 |
| `jina_ai` | 网页提取 |
| `firecrawl` | 网页爬取 |
| `image_search` | 图像搜索 |

---

## 四、完整差距总结

### 4.1 完成度统计

| 模块 | 完成度 | 说明 |
|------|--------|------|
| ThreadState 状态系统 | ✅ 100% | 21 个单元测试全部通过 |
| 模块化提示词系统 | ✅ 100% | 34 个单元测试全部通过 |
| Sandbox 虚拟路径系统 | ✅ 95% | 缺少 /mnt/skills 支持（低优先级） |
| 中间件链系统 | 🔄 20% | 框架完成，12 个中间件待实现 |
| DeerFlow 风格工具系统 | 🔄 15% | 框架完成，13+ 个工具待实现 |
| 子代理系统 | ⏳ 0% | 完整系统待创建 |
| 配置系统 | ⏳ 0% | 完整系统待创建 |
| Skills 系统 | ⏳ 0% | 低优先级，可后期集成 |
| 记忆系统 | ⏳ 0% | 低优先级，可后期集成 |
| MCP 系统 | ⏳ 0% | 低优先级，可后期集成 |
| Lead Agent 工厂 | ⏳ 0% | 核心集成待实现 |

**总体完成度：约 25%**

---

### 4.2 剩余工作量预估

| 优先级 | 任务 | 预估工作量 | 关键文件 |
|--------|------|-----------|---------|
| **P0** | **完整中间件实现（12 个）** | **3-4 天** | `pkg/agent/middleware/*.go` |
| **P0** | **完整工具实现（13+ 个）** | **3-4 天** | `pkg/agent/tools/deerflow/*.go` |
| **P0** | **子代理系统** | **3-4 天** | `pkg/agent/subagent/*.go` |
| **P0** | **Lead Agent 工厂集成** | **2 天** | `pkg/agent/eino_agent.go` |
| **P1** | 配置系统 | 2-3 天 | `pkg/config/*.go` |
| **P1** | 端到端测试 | 2-3 天 | 集成测试 |
| **P2** | Skills 系统 | 2 天 | 可后期集成 |
| **P2** | 记忆系统 | 2 天 | 可后期集成 |
| **P2** | MCP 系统 | 2 天 | 可后期集成 |

**总计 P0 工作量：10-14 天**

---

## 五、关键一比一复刻点（尚未完成）

### 5.1 中间件关键细节

1. **ThreadDataMiddleware**
   - `before_agent()` 创建线程目录
   - 支持 lazy_init 模式（仅计算路径）
   - 返回 `thread_data` 状态更新

2. **UploadsMiddleware**
   - 从 `message.additional_kwargs.files` 提取新上传文件
   - 扫描历史上传文件
   -  prepend `<uploaded_files>` 块到最后一条用户消息
   - 保留 `additional_kwargs` 供前端读取

3. **DanglingToolCallMiddleware**
   - 使用 `wrap_model_call()` 而非 `before_model()`（关键！）
   - 检测 AIMessage.tool_calls 无对应 ToolMessage 的情况
   - 插入占位 ToolMessage（`[Tool call was interrupted...]`）
   - 位置正确：紧跟在 AIMessage 后

4. **TitleMiddleware**
   - `after_model()` / `aafter_model()` 生成标题
   - 触发条件：第 1 条用户消息 + 第 1 条 AI 响应
   - 支持 fallback 到用户消息前 50 字符
   - 配置：max_words, max_chars, prompt_template

5. **LoopDetectionMiddleware**
   - `_hash_tool_calls()`: order-independent 哈希（排序后）
   - `_track_and_check()`: LRU 历史记录 + 警告/硬停止阈值
   - `_apply()`: 注入警告消息或剥离 tool_calls
   - `reset(thread_id)`: 清除跟踪状态

6. **ClarificationMiddleware**
   - `wrap_tool_call()` / `awrap_tool_call()` 拦截
   - 识别 `ask_clarification` 工具调用
   - 格式化消息（5 种类型图标）
   - 返回 `Command(goto=END)` 中断执行

### 5.2 工具关键细节

1. **bash 工具**
   - `ensure_sandbox_initialized()` 懒加载沙箱
   - `validate_local_bash_command_paths()` 验证路径
   - `replace_virtual_paths_in_command()` 翻译路径
   - `mask_local_paths_in_output()` 掩码结果
   - 支持 `/mnt/skills` 只读路径

2. **ls 工具**
   - 树状格式，最多 2 层
   - `validate_local_tool_path(..., read_only=True)`
   - `/mnt/skills` 路径特殊处理 → `_resolve_skills_path()`

3. **read_file 工具**
   - 支持 `start_line` / `end_line`（1-indexed, inclusive）
   - `"\n".join(content.splitlines()[start_line-1 : end_line])`

4. **write_file 工具**
   - `append` 参数支持
   - 自动创建目录

5. **str_replace 工具**
   - 默认仅替换 1 处（要求 old_str 唯一！）
   - `replace_all=True` 替换全部

6. **present_files 工具**
   - 仅接受 `/mnt/user-data/outputs/*` 路径
   - `_normalize_presented_filepath()` 规范化
   - 返回 `Command(update={"artifacts": [...]})`

7. **view_image 工具**
   - 验证扩展名：`.jpg`, `.jpeg`, `.png`, `.webp`
   - base64 编码
   - 返回 `Command(update={"viewed_images": {...}})`

8. **task 工具（最复杂！）**
   - `SubagentExecutor.execute_async()` 后台执行
   - 双线程池：_scheduler_pool (3) + _execution_pool (3)
   - 5 秒轮询
   - 实时事件流：`task_started` → `task_running` (每条 AI 消息) → `task_completed/failed/timed_out`
   - `MAX_CONCURRENT_SUBAGENTS = 3`
   - `timeout_seconds = 900` (15 分钟)
   - 工具过滤：移除 `task` 防止递归嵌套
   - 状态继承：`sandbox_state`, `thread_data`
   - trace_id 分布式追踪

### 5.3 子代理系统关键细节

1. **SubagentExecutor**
   - `_create_agent()`: 创建子 agent（复用中间件）
   - `_aexecute()`: 异步执行 + 流式捕获 AI 消息
   - `execute()`: 同步包装（`asyncio.run()`）
   - `execute_async()`: 后台执行（返回 task_id）

2. **双线程池**
   - `_scheduler_pool`: max_workers=3
   - `_execution_pool`: max_workers=3
   - 分离调度和执行防止阻塞

3. **背景任务管理**
   - `_background_tasks`: 全局 dict 存储
   - `get_background_task_result(task_id)`: 查询状态
   - `cleanup_background_task(task_id)`: 仅清理终端状态任务

---

## 六、分阶段实施建议（修订版）

### Phase 0: 准备工作 ✅（已完成）
- ✅ 深入分析 DeerFlow 完整源码
- ✅ 识别 13 个中间件、15+ 个工具、完整子代理系统
- ✅ 本文档

### Phase 1: 完整中间件实现（P0，4 天）
- [ ] ThreadDataMiddleware
- [ ] UploadsMiddleware
- [ ] SandboxMiddleware
- [ ] DanglingToolCallMiddleware
- [ ] SummarizationMiddleware（可选）
- [ ] TodoListMiddleware（可选）
- [ ] TitleMiddleware
- [ ] MemoryMiddleware（可选）
- [ ] ViewImageMiddleware
- [ ] SubagentLimitMiddleware（可选）
- [ ] LoopDetectionMiddleware（完整逻辑）
- [ ] ToolErrorHandlingMiddleware
- [ ] DeferredToolFilterMiddleware
- [ ] ClarificationMiddleware（必须最后！）

### Phase 2: 完整工具实现（P0，4 天）
- [ ] bash 工具（带路径翻译）
- [ ] ls 工具（树状格式）
- [ ] read_file 工具（带行范围）
- [ ] write_file 工具（带 append）
- [ ] str_replace 工具
- [ ] present_files 工具
- [ ] ask_clarification 工具
- [ ] view_image 工具
- [ ] write_todos 工具（可选）
- [ ] 工具安全验证层（路径验证、翻译、掩码）

### Phase 3: 子代理系统（P0，4 天）
- [ ] SubagentConfig, SubagentStatus, SubagentResult 类型
- [ ] SubagentExecutor（核心！）
- [ ] 双线程池（_scheduler_pool + _execution_pool）
- [ ] 事件系统（task_started, task_running, task_completed, task_failed, task_timed_out）
- [ ] 工具过滤（移除 task 防止嵌套）
- [ ] task() 工具完整实现（5s 轮询 + 实时事件）

### Phase 4: Lead Agent 集成（P0，2 天）
- [ ] make_lead_agent() 工厂
- [ ] 中间件链严格顺序组合
- [ ] BuildLeadAgentPromptWithTools() 集成
- [ ] 运行时配置支持（thinking_enabled, model_name, is_plan_mode, subagent_enabled）

### Phase 5: 配置系统（P1，2-3 天）
- [ ] AppConfig, ModelConfig, SandboxConfig 等
- [ ] config.yaml 加载
- [ ] extensions_config.json 加载

### Phase 6: 端到端测试（P1，2-3 天）
- [ ] 完整端到端测试
- [ ] 性能优化

### Phase 7-9: 低优先级功能（P2，按需）
- [ ] Skills 系统
- [ ] 记忆系统
- [ ] MCP 系统

---

## 七、参考资料

### DeerFlow 关键源码文件

```
deer-flow/backend/packages/harness/deerflow/
├── agents/
│   ├── thread_state.py                    # ThreadState 定义
│   ├── lead_agent/
│   │   ├── agent.py                       # Lead Agent 工厂
│   │   └── prompt.py                      # 系统提示词
│   └── middlewares/
│       ├── thread_data_middleware.py      # 1. 线程目录
│       ├── uploads_middleware.py          # 2. 上传文件
│       ├── sandbox_middleware.py          # 3. 沙箱生命周期
│       ├── dangling_tool_call_middleware.py # 4. 挂起工具调用
│       ├── todo_middleware.py             # 6. 任务列表
│       ├── title_middleware.py            # 7. 标题生成
│       ├── memory_middleware.py           # 8. 记忆提取
│       ├── view_image_middleware.py       # 9. 图像处理
│       ├── subagent_limit_middleware.py   # 10. 子代理限制
│       ├── loop_detection_middleware.py   # 11. 循环检测
│       ├── tool_error_handling_middleware.py # 12. 工具错误处理
│       ├── deferred_tool_filter_middleware.py # 13. 延迟工具过滤
│       └── clarification_middleware.py    # 14. 询问澄清（最后！）
├── sandbox/
│   ├── sandbox.py                         # Sandbox 接口
│   ├── tools.py                           # 5 个 Sandbox 工具 + 路径翻译
│   └── local/
│       └── local_sandbox.py               # LocalSandbox 实现
├── tools/
│   ├── tools.py                           # get_available_tools()
│   └── builtins/
│       ├── present_file_tool.py           # present_files
│       ├── clarification_tool.py          # ask_clarification
│       ├── view_image_tool.py             # view_image
│       └── task_tool.py                   # task（子代理委托）
├── subagents/
│   ├── executor.py                        # SubagentExecutor（双线程池）
│   └── registry.py                        # 子代理注册
└── config/
    ├── app_config.py                      # AppConfig
    ├── model_config.py                    # ModelConfig
    ├── sandbox_config.py                  # SandboxConfig
    ├── title_config.py                    # TitleConfig
    └── subagents_config.py                # SubagentsConfig
```

---

## 总结

**当前状态**：
- ✅ ThreadState：100% 完成
- ✅ 模块化提示词：100% 完成
- ✅ Sandbox 虚拟路径：95% 完成
- 🔄 中间件链：20% 完成（框架）
- 🔄 工具系统：15% 完成（框架）
- ⏳ 子代理系统：0% 完成
- ⏳ 配置系统：0% 完成
- ⏳ Lead Agent 集成：0% 完成

**总体完成度：约 25%**

**剩余 P0 工作量：10-14 天**
