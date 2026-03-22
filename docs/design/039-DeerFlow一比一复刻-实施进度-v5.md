# DeerFlow 一比一复刻 - 实施进度 v5

## 变更记录表

| 日期 | 版本 | 变更说明 |
|------|------|----------|
| 2026-03-22 | v5 | `chain.go` 装配真实中间件类型（替换 `NewBaseMiddleware` 占位）；`MiddlewareConfig` 增加可选 `SandboxProvider`；子代理链使用 `ThreadDataMiddleware` + 可配置循环检测；多份中间件与 Eino `schema.Message` 对齐（`ToolCalls`/`ToolCallID`、`schema.Assistant`/`Tool`/`User`/`System`）；消除 `BeforeAgent` 参数名 `state` 与 `state` 包冲突；`WorkerPool.Stop` 后重建 `stopChan` 以支持多次 `Start`/`Stop`；`LeadAgentFactory.BuildSystemPrompt` 改为 `prompts.BuildLeadAgentPromptString`；承接 [v4](./038-DeerFlow一比一复刻-实施进度-v4.md) 基线。 |

---

## 最新更新（相对 v4）

### 中间件链装配修正 ✅

- **问题**：`BuildLeadAgentMiddlewares` / `BuildSubagentMiddlewares` 中多数步骤使用 `NewBaseMiddleware("name")`，仅注册名称，**未挂载**各文件的 `*Middleware` 实现（`BeforeAgent` / `WrapToolCall` 等逻辑不会通过类型断言参与后续集成）。
- **处理**：
  - Lead 链按 DeerFlow 顺序实例化：`NewThreadDataMiddleware`、`NewUploadsMiddleware`、`NewSandboxMiddleware` / `NewDefaultSandboxMiddleware`、`NewDanglingToolCallMiddleware`、可选 `Summarization` / `TodoList` / `Title` / `Memory`、`NewViewImageMiddleware`、可选 `NewSubagentLimitMiddleware`、`NewLoopDetectionMiddlewareWithConfig`（尊重 `LoopWarnThreshold` / `LoopHardLimit`）、`NewToolErrorHandlingMiddleware`、`NewDeferredToolFilterMiddleware`、`NewClarificationMiddleware`。
  - 子代理链：`NewThreadDataMiddleware` + `NewLoopDetectionMiddlewareWithConfig`（与 Lead 共用阈值解析）。
- **配置扩展**：`MiddlewareConfig` 增加 `SandboxProvider sandbox.SandboxProvider`；为 `nil` 时沿用 `NewDefaultSandboxMiddleware()`（与此前占位行为一致，便于后续接入真实 Provider）。
- **工厂**：`LeadAgentFactory.BuildMiddlewareConfig` 无需改动即可编译；若将来有全局 `SandboxProvider`，只需在构建 `MiddlewareConfig` 时赋值。

### 测试修复 ✅

- `factory_test.go`：`len(chain.Middlewares)` 修正为 `len(chain.Middlewares())`；`BuildSystemPrompt` 断言改为 `<role>` / `<thinking_style>`；`TestRuntimeConfig_Partial` 移除未使用变量。
- **WorkerPool**：`Stop()` 结束后再 `make(chan struct{})` 赋回 `stopChan`，避免连续多次 `MakeLeadAgent` + `Cleanup` 时重复 `close` panic。

### 工厂与提示词对齐 ✅

- `factory.BuildSystemPrompt()` 使用 `prompts.LeadAgentConfig` + `BuildLeadAgentPromptString()`，与当前 `sections.go` / `builder.go` API 一致。

### Eino 消息模型对齐 ✅

- 涉及消息与工具调用的中间件改为使用 `github.com/cloudwego/eino/schema` 的 `Message`、`ToolCall` 字段，不再使用已不存在的 `Meta` / `state.Message` 占位。

---

## 完整实施清单（与 v4 一致部分略）

### 中间件系统 (pkg/agent/middleware/) — 更新说明

- ✅ `types.go`：接口与链结构（不变）
- ✅ **`chain.go`：Lead / Subagent 链使用完整中间件类型装配**（本次）
- ✅ 各 `*_middleware.go`：实现文件保持 v4 状态；**链路与实现类已对齐**，后续 Eino 回调桥接可按接口分派

---

## 剩余工作（摘录，与 v4 一致）

P0 仍为：各中间件**业务逻辑**深化、与 **Eino `callbacks.Handler`** 的完整对齐。本次变更解决的是「链上对象是否正确」而非「每个钩子是否已接 LLM/文件系统」。

---

## 完成度估算（相对 v4 微调）

| 模块 | v4 | v5 | 变化 |
|------|-----|-----|------|
| 中间件系统 | ~50% | **~55%** | 装配与类型对齐 +5%（逻辑完整度未变） |
| 其他模块 | 同 v4 | 同 v4 | - |

---

## 总结

1. ✅ **中间件链使用真实类型**：与 `pkg/agent/middleware` 下各实现一致，便于后续桥接与测试断言具体 `Name()` 与行为。
2. ✅ **可注入 `SandboxProvider`**：为工厂/运行时接入沙箱实现预留字段。
3. ✅ **循环检测阈值**：Lead 与子代理链均通过 `NewLoopDetectionMiddlewareWithConfig` 应用配置。

前一版详细清单与 P0–P2 条目见：[实施进度 v4](./038-DeerFlow一比一复刻-实施进度-v4.md)。
