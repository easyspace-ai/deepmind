# DeerFlow 一比一复刻 - 实施进度 v6

## 变更记录表

| 日期 | 版本 | 变更说明 |
|------|------|----------|
| 2026-03-22 | v6 | 对齐「DeerFlow 与 nanobot-go 全面对比及对齐计划」后续落地：`pkg/agent/memory` 去抖队列与 `MemoryMiddleware` 对接；`pkg/agent/skills` DeerFlow 式 SKILL.md 扫描与 `PromptSection`；`write_todos` 工具与内置工具表注册；子代理 `runWithEino` 在 `MiddlewareConfig.SandboxProvider` 非空时 `defer Release`；`state.ApplyMiddlewareUpdates` 单测；`deerflow` 包去除 `types.go` 与 `sandbox_tools`/`builtin`/`task` 的重复符号；`BaseDeerFlowTool.Info` 改为 `schema.ToolInfo` + `NewParamsOneOfByJSONSchema`；`tool_security` 修复 raw string 内反引号导致的编译错误并改用 RE2 兼容路径正则；`pkg/models` 增加 `DeerFlowModelEntry` 语义锚点（实际建模型仍走 `pkg/agent/provider`）；Planner 页 Playwright 冒烟；承接 [v5](./039-DeerFlow一比一复刻-实施进度-v5.md)。 |

---

## 相对 v5 的增量

### 运行时与状态

- **Memory**：`pkg/agent/memory` 提供 `DefaultManager`、`EnqueueFromThreadState`（按 workspace 聚合文本片段，供后续 LLM 事实提取扩展）。
- **Skills**：`pkg/agent/skills` 提供 `LoadDir`、`PromptSection`（YAML frontmatter + 正文解析）。
- **状态合并测试**：`pkg/agent/state/middleware_updates_test.go` 覆盖 `ApplyMiddlewareUpdates` 主要键。

### 工具与 deerflow 包

- **write_todos**：`NewWriteTodosTool` 写入 `ToolConfig.ThreadState.Todos`（需调用方绑定共享 `ThreadState`）。
- **getSubagentTools**：`NewTaskTool(nil)` 与 `task_tool.go` 签名一致。
- **ToolInfo**：与 eino v0.7 `ToolInfo`（`Desc`、`ParamsOneOf`）一致。

### 子代理

- **沙箱释放**：子线程 ID 执行结束后对 `SandboxProvider.Release` 做 best-effort 调用（避免测试/无 Provider 时误伤）。

### 模型工厂（nanobot 替代说明）

- **`pkg/models`**：仅保留 `DeerFlowModelEntry` 配置结构说明；**不**重复实现反射式 `create_chat_model`，与 nanobot 统一入口 `pkg/agent/provider` 一致，避免双栈。

### 前端 E2E

- **`web/e2e/tests/planner.spec.ts`**：登录后访问 `/planner`，断言「Eino 自动规划系统」可见。

---

## 仍为 P0/P1 的已知缺口（摘录）

- 各中间件与 DeerFlow 的 **业务深度**（Summarization LLM、Title 独立模型、ViewImage 读盘 base64、DeferredTool 注册表 + tool_search、Clarification 中断与图状态等）仍按 v4 优先级逐步补齐。
- 子代理 **逐轮** 与 Eino `compose.WithCallbacks` 全量挂钩（相对当前「Before/After 各一轮」）仍为架构项。
- **MCP**：与 DeerFlow 多服/OAuth/mtime 缓存的逐项对比见 [041](./041-DeerFlow一比一复刻-差距分析-对齐计划补充.md)。

---

## 完成度估算（相对 v5）

| 模块 | v5 | v6 | 说明 |
|------|-----|-----|------|
| Memory / Skills | 缺失 / 缺失 | **初版** | 可接入中间件与提示词 |
| deerflow 工具 | 缺 write_todos | **+write_todos** | 仍缺 community / tool_search 等 |
| 子代理沙箱 | 未释放 | **可选 Release** | 依赖注入 Provider |
| 文档 / E2E | - | **+v6 + Planner 冒烟** | - |

---

前一版基线：[实施进度 v5](./039-DeerFlow一比一复刻-实施进度-v5.md)。
