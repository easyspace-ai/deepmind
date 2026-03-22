# DeerFlow Go - 源码级对齐差距报告

| 修改人 | 修改时间 | 修改内容 |
| ------ | -------- | -------- |
| AI Assistant | 2026-03-22 | 初始版本 - 源码级深度对比差距分析 |

---

## 对比概览

本次对比基于：
- **DeerFlow Python**: `deer-flow/backend/packages/harness/deerflow/`
- **DeerFlow Go**: `/Users/leven/space/test/nanobot-go/`

---

## 核心差距清单

### 1. 中间件系统差距

#### 1.1 缺失的中间件

| 中间件 | DeerFlow Python | DeerFlow Go | 优先级 |
|--------|-----------------|-------------|--------|
| **SummarizationMiddleware** | ✅ 完整实现 | ❌ 缺失 | P0 |
| **DeferredToolFilterMiddleware** | ✅ 完整实现 | ❌ 缺失 | P0 |
| **LoopDetectionMiddleware** | ✅ 完整实现 | ❌ 缺失 | P0 |
| **ToolErrorHandlingMiddleware** | ✅ 完整实现 | ❌ 缺失 | P0 |

**总计**: 4 个关键中间件缺失

#### 1.2 中间件链构建顺序

DeerFlow Python 有严格的中间件链顺序（`_build_middlewares()`）：

```python
# DeerFlow Python 顺序
1. build_lead_runtime_middlewares()
   ├── ThreadDataMiddleware
   ├── UploadsMiddleware
   ├── SandboxMiddleware
   ├── DanglingToolCallMiddleware
   └── ToolErrorHandlingMiddleware  ← 缺失
2. SummarizationMiddleware (可选)        ← 缺失
3. TodoListMiddleware (可选)
4. TitleMiddleware
5. MemoryMiddleware
6. ViewImageMiddleware (条件)
7. DeferredToolFilterMiddleware (条件)  ← 缺失
8. SubagentLimitMiddleware (条件)
9. LoopDetectionMiddleware             ← 缺失
10. ClarificationMiddleware (最后)
```

---

### 2. Lead Agent 工厂差距

#### 2.1 `make_lead_agent()` 函数

**DeerFlow Python** (`agents/lead_agent/agent.py`):
- ✅ `_resolve_model_name()` - 模型名称解析
- ✅ `_create_summarization_middleware()` - 摘要中间件创建
- ✅ `_create_todo_list_middleware()` - TodoList 中间件创建
- ✅ `_build_middlewares()` - 中间件链构建
- ✅ `make_lead_agent()` - Lead Agent 工厂函数
- ✅ 支持 `thinking_enabled`, `is_plan_mode`, `subagent_enabled` 等运行时配置
- ✅ 支持 per-agent 配置加载
- ✅ 支持 bootstrap agent 模式

**DeerFlow Go**:
- ❌ 缺少对应的工厂函数
- ❌ 缺少运行时配置支持
- ❌ 缺少 per-agent 配置集成

---

### 3. 工具系统差距

#### 3.1 `get_available_tools()` 集成

**DeerFlow Python** (`tools/__init__.py`):
- ✅ 支持 `groups` 参数过滤工具组
- ✅ 支持 `include_mcp` 参数
- ✅ 支持 `model_name` 参数
- ✅ 支持 `subagent_enabled` 参数
- ✅ 集成 MCP 工具
- ✅ 集成社区工具
- ✅ 集成子代理工具

**DeerFlow Go**:
- ⚠️ 已有 `GetAvailableTools()`，但需要完整对齐

#### 3.2 `setup_agent` 工具

**DeerFlow Python**:
- ✅ `setup_agent` 工具用于初始化自定义 agent
- ❌ DeerFlow Go 缺失

---

### 4. 配置系统差距

#### 4.1 `AgentsConfig`

**DeerFlow Python** (`config/agents_config.py`):
- ✅ per-agent 配置加载
- ✅ agent model 配置
- ✅ agent tool_groups 配置
- ❌ DeerFlow Go 缺失

---

### 5. 运行时中间件构建器差距

#### 5.1 `build_lead_runtime_middlewares()`

**DeerFlow Python** (`agents/middlewares/tool_error_handling_middleware.py`):
- ✅ `_build_runtime_middlewares()` - 构建基础中间件链
- ✅ `build_lead_runtime_middlewares()` - Lead Agent 中间件链
- ✅ `build_subagent_runtime_middlewares()` - 子代理中间件链
- ❌ DeerFlow Go 缺失

---

## 各中间件详细差距

### SummarizationMiddleware

**位置**: `agents/middlewares/summarization_middleware.py` (DeerFlow Python)

**功能**:
- Token/message 触发检测
- Keep 策略: `keep.last_n`, `keep.first_n`, `keep.fraction`
- 模型调用生成摘要
- 与 `config.summarization` 集成

**关键代码**:
```python
def _create_summarization_middleware() -> SummarizationMiddleware | None:
    config = get_summarization_config()
    if not config.enabled:
        return None
    # trigger, keep, model 配置
    return SummarizationMiddleware(...)
```

---

### DeferredToolFilterMiddleware

**位置**: `agents/middlewares/deferred_tool_filter_middleware.py`

**功能**:
- 从 `request.tools` 中移除延迟工具
- 只在 `app_config.tool_search.enabled` 时启用
- LLM 只看到 active tool schemas
- 延迟工具通过 `tool_search` 在运行时发现

**关键代码**:
```python
def _filter_tools(self, request: ModelRequest) -> ModelRequest:
    registry = get_deferred_registry()
    deferred_names = {e.name for e in registry.entries}
    active_tools = [t for t in request.tools if t.name not in deferred_names]
    return request.override(tools=active_tools)
```

---

### LoopDetectionMiddleware

**位置**: `agents/middlewares/loop_detection_middleware.py`

**功能**:
- Order-independent 哈希（工具调用顺序不影响哈希）
- LRU 缓存（最近 N 个状态）
- 警告阈值（默认 3 次）
- 硬停止阈值（默认 5 次）
- 滑动窗口（默认 20）
- 最大跟踪线程数（默认 100，LRU 淘汰）

**关键代码**:
```python
def _hash_tool_calls(tool_calls: list[dict]) -> str:
    # 按 name 和 args 排序后哈希
    normalized.sort(key=lambda tc: (
        tc["name"],
        json.dumps(tc["args"], sort_keys=True)
    ))
    return hashlib.md5(blob.encode()).hexdigest()[:12]
```

---

### ToolErrorHandlingMiddleware

**位置**: `agents/middlewares/tool_error_handling_middleware.py`

**功能**:
- `wrap_tool_call()` / `awrap_tool_call()` - 同步/异步包装
- 异常捕获与转换
- 生成错误 ToolMessage
- 保留 GraphBubbleUp 控制流信号
- 错误详情截断（最大 500 字符）

**关键代码**:
```python
def wrap_tool_call(self, request, handler):
    try:
        return handler(request)
    except GraphBubbleUp:
        raise  # 保留控制流信号
    except Exception as exc:
        return self._build_error_message(request, exc)
```

---

## 完整中间件链对比表

| 阶段 | DeerFlow Python | DeerFlow Go | 状态 |
|------|-----------------|-------------|------|
| **1. Runtime 基础链** | | | |
| ThreadDataMiddleware | ✅ | ✅ | 已对齐 |
| UploadsMiddleware | ✅ | ✅ | 已对齐 |
| SandboxMiddleware | ✅ | ✅ | 已对齐 |
| DanglingToolCallMiddleware | ✅ | ✅ | 已对齐 |
| **ToolErrorHandlingMiddleware** | ✅ | ❌ | 缺失 |
| **2. Lead 专用链** | | | |
| **SummarizationMiddleware** (可选) | ✅ | ❌ | 缺失 |
| TodoListMiddleware (可选) | ✅ | ✅ | 已对齐 |
| TitleMiddleware | ✅ | ✅ | 已对齐 |
| MemoryMiddleware | ✅ | ✅ | 已对齐 |
| ViewImageMiddleware (条件) | ✅ | ✅ | 已对齐 |
| **DeferredToolFilterMiddleware** (条件) | ✅ | ❌ | 缺失 |
| SubagentLimitMiddleware (条件) | ✅ | ✅ | 已对齐 |
| **LoopDetectionMiddleware** | ✅ | ❌ | 缺失 |
| ClarificationMiddleware (最后) | ✅ | ✅ | 已对齐 |

**对齐统计**: 10/14 (71%) 对齐，4 个缺失

---

## 优先级排序

### P0 - 必须立即补齐（核心功能）

1. **ToolErrorHandlingMiddleware** - 工具错误处理，防止运行中断
2. **DeferredToolFilterMiddleware** - 延迟工具过滤，节省 token
3. **LoopDetectionMiddleware** - 循环检测，安全关键
4. **SummarizationMiddleware** - 上下文摘要，防止 token 溢出

### P1 - 重要功能

5. **Lead Agent 工厂函数** (`make_lead_agent()`)
6. **运行时配置支持** (`thinking_enabled`, `is_plan_mode`, etc.)
7. **AgentsConfig** - per-agent 配置

### P2 - 增强功能

8. **setup_agent 工具** - 自定义 agent 初始化

---

## 实施建议

### 阶段 A: 补齐缺失的中间件（2-3 天）

1. Day 1: ToolErrorHandlingMiddleware + DeferredToolFilterMiddleware
2. Day 2: LoopDetectionMiddleware
3. Day 3: SummarizationMiddleware

### 阶段 B: Lead Agent 工厂（1-2 天）

1. Day 1: make_lead_agent() 基础框架
2. Day 2: 运行时配置 + per-agent 配置

### 阶段 C: 测试与集成（1 天）

1. 完整中间件链集成测试
2. 端到端测试

---

## 总结

### 当前对齐状态

| 指标 | 数值 |
|------|------|
| 中间件对齐 | 10/14 (71%) |
| 总体完成度 | ~90% → **75-80%**（修正后） |
| 缺失关键组件 | 4 个中间件 + 工厂函数 |

### 修正后的完成度估算

考虑到缺失的 4 个关键中间件和 Lead Agent 工厂，**实际完成度约为 75-80%**，而非之前估计的 90%。

### 建议

优先补齐 P0 中间件（ToolErrorHandling, DeferredToolFilter, LoopDetection, Summarization），然后补齐 Lead Agent 工厂函数，以达到真正的 90%+ 完成度。
