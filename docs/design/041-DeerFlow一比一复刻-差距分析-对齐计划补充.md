# DeerFlow 一比一复刻 - 差距分析（对齐计划补充）

> 本文档在 [037 v3](./037-DeerFlow一比一复刻-完整差距分析-v3.md) 与 [038 v4 进度](./038-DeerFlow一比一复刻-实施进度-v4.md) 基础上，记录 **对齐计划** 落地后的状态切片；**不**逐条重述 v3 全文。

## 变更记录表

| 日期 | 说明 |
|------|------|
| 2026-03-22 | 初版：标注「已解决 / 仍缺 / nanobot 替代」。 |

---

## 已解决或显著推进（相对 v3 过时表述）

| 项 | 说明 |
|----|------|
| Eino ↔ 中间件链 | `NewDeerflowEinoHandler` + `EinoCallbackBridge` 默认实现；`RunBeforeAgentPhase` 等供子代理复用（见 `pkg/agent/middleware/eino_callbacks.go`）。 |
| 子代理真实 Agent | `SubagentExecutor` 可选 `WithChatModel` + `BuildSubagentMiddlewares`（见 v5/v6 进度文）。 |
| Memory 包 | `pkg/agent/memory` + `MemoryMiddleware.AfterModel` 入队。 |
| DeerFlow Skills 合同 | `pkg/agent/skills` loader；与现有 `pkg/agent/skills.go` / `tools/skill` **并存**，需在集成层择一或映射。 |
| write_todos | `pkg/agent/tools/deerflow/write_todos_tool.go` + `getBuiltinTools` 注册。 |
| 模型工厂 | **nanobot 替代**：`pkg/models.DeerFlowModelEntry` 仅语义锚点；真实创建用 `pkg/agent/provider`。 |
| deerflow 包编译 | 去除重复构造函数；`tool_security` 正则字面量修复；`ToolInfo` API 与 eino 0.7 对齐。 |

---

## 仍缺 / 部分对齐

| 项 | 说明 |
|----|------|
| 中间件业务深度 | Summarization、Title、ViewImage、DeferredTool+registry、ToolErrorHandling 等与 DeerFlow 行为仍有差距（同 v4 表）。 |
| tool_search / 社区工具 | Tavily/Jina 等未在 `deerflow` 包内等价落地；可声明由 nanobot 其他工具替代并固定工具表。 |
| MCP 对等 | DeerFlow：多服务器、mtime 缓存、OAuth 刷新等；nanobot：`pkg/agent/tools/mcp` — **需逐项 diff 后补缺**。 |
| 统一运行时 | `hooks/eino` 与 DeerFlow 中间件链 **合并** 仍为可选路线（计划默认先复刻轨闭环）。 |

---

## 总结

- **复刻轨**（middleware + deerflow tools + subagent + state）与 **nanobot 主 Agent**（provider、hooks）关系已在 [040 v6](./040-DeerFlow一比一复刻-实施进度-v6.md) 进度文中保持「并行存在、文档标注替代」策略。
