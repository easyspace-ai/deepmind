# 045-DeerFlow-Frontend-API-对接规划

| 修改人 | 修改时间 | 修改内容 |
| ------ | -------- | -------- |
| AI Assistant | 2026-03-23 | 初始版本 - API 对接规划 |

---

## 1. 需求概述

DeerFlow Go 后端需要对接 DeerFlow 原版前端（位于 `frontend/` 目录）。本文档梳理前端调用的 API，规划前后端 API 路径，并制定实现方案。

---

## 2. 前端 API 调用梳理

### 2.1 环境变量配置

前端通过以下环境变量配置后端地址：

```bash
NEXT_PUBLIC_BACKEND_BASE_URL=http://localhost:8001  # 普通 API
NEXT_PUBLIC_LANGGRAPH_BASE_URL=http://localhost:2024  # LangGraph API
```

### 2.2 API 分类

前端调用的 API 分为两类：

1. **普通 REST API** (`/api/*`) - 配置、列表类接口
2. **LangGraph SDK API** (`/api/langgraph/*`) - 对话流、线程管理

---

## 3. API 清单与规划

### 3.1 普通 REST API (`/api/*`)

| 序号 | API 路径 | 方法 | 前端文件 | 功能 | 后端状态 | 优先级 |
|-----|---------|------|---------|------|---------|-------|
| 1 | `/api/models` | GET | `core/models/api.ts` | 获取模型列表 | ❌ 缺失 | P0 |
| 2 | `/api/agents` | GET | `core/agents/api.ts` | 获取 Agent 列表 | ⚠️ 格式不匹配 | P0 |
| 3 | `/api/agents/:name` | GET | `core/agents/api.ts` | 获取单个 Agent | ⚠️ 格式不匹配 | P0 |
| 4 | `/api/agents` | POST | `core/agents/api.ts` | 创建 Agent | ⚠️ 格式不匹配 | P1 |
| 5 | `/api/agents/:name` | PUT | `core/agents/api.ts` | 更新 Agent | ⚠️ 格式不匹配 | P1 |
| 6 | `/api/agents/:name` | DELETE | `core/agents/api.ts` | 删除 Agent | ⚠️ 格式不匹配 | P1 |
| 7 | `/api/agents/check?name=xxx` | GET | `core/agents/api.ts` | 检查 Agent 名称可用性 | ❌ 缺失 | P1 |
| 8 | `/api/skills` | GET | `core/skills/api.ts` | 获取技能列表 | ⚠️ 格式不匹配 | P0 |
| 9 | `/api/skills/:skillName` | PUT | `core/skills/api.ts` | 启用/禁用技能 | ❌ 缺失 | P1 |
| 10 | `/api/skills/install` | POST | `core/skills/api.ts` | 安装技能 | ❌ 缺失 | P2 |
| 11 | `/api/memory` | GET | `core/memory/api.ts` | 获取用户记忆 | ❌ 缺失 | P2 |
| 12 | `/api/mcp/config` | GET | `core/mcp/api.ts` | 获取 MCP 配置 | ❌ 缺失 | P1 |
| 13 | `/api/mcp/config` | PUT | `core/mcp/api.ts` | 更新 MCP 配置 | ❌ 缺失 | P1 |
| 14 | `/api/threads/:threadId/uploads` | POST | `core/uploads/api.ts` | 上传文件 | ❌ 缺失 | P0 |
| 15 | `/api/threads/:threadId/uploads/list` | GET | `core/uploads/api.ts` | 列出上传文件 | ❌ 缺失 | P1 |
| 16 | `/api/threads/:threadId/uploads/:filename` | DELETE | `core/uploads/api.ts` | 删除上传文件 | ❌ 缺失 | P1 |
| 17 | `/api/threads/:threadId/artifacts/*` | GET | mock 路由 | 获取 artifact 文件 | ❌ 缺失 | P1 |

### 3.2 LangGraph SDK API (`/api/langgraph/*`)

前端使用 `@langchain/langgraph-sdk` 客户端调用以下 API：

| 序号 | LangGraph SDK 方法 | 对应 HTTP 路径 | 功能 | 后端状态 | 优先级 |
|-----|-------------------|---------------|------|---------|-------|
| 1 | `client.runs.stream()` | `POST /threads/:threadId/runs/stream` | 流式运行 Agent | ❌ 缺失 | P0 |
| 2 | `client.runs.joinStream()` | `POST /threads/:threadId/runs/:runId/join` | 加入流式运行 | ❌ 缺失 | P1 |
| 3 | `client.threads.search()` | `POST /threads/search` | 搜索线程列表 | ❌ 缺失 | P0 |
| 4 | `client.threads.create()` | `POST /threads` | 创建线程 | ❌ 缺失 | P0 |
| 5 | `client.threads.get()` | `GET /threads/:threadId` | 获取线程 | ❌ 缺失 | P0 |
| 6 | `client.threads.updateState()` | `POST /threads/:threadId/state` | 更新线程状态 | ❌ 缺失 | P1 |
| 7 | `client.threads.delete()` | `DELETE /threads/:threadId` | 删除线程 | ❌ 缺失 | P0 |
| 8 | `client.threads.getHistory()` | `POST /threads/:threadId/history` | 获取线程历史 | ❌ 缺失 | P1 |

---

## 4. API 响应格式对比

### 4.1 现有后端格式 (`/api/v1/*`)

```go
// 列表响应
type ListResponse struct {
    Items    interface{} `json:"items"`
    Total    int64       `json:"total"`
    Page     int         `json:"page,omitempty"`
    PageSize int         `json:"page_size,omitempty"`
}

// 错误响应
type ErrorResponse struct {
    Error string `json:"error"`
}
```

### 4.2 前端期望格式 (`/api/*`)

```typescript
// Models API
GET /api/models
{
  models: Model[]
}

// Agents API
GET /api/agents
{
  agents: Agent[]
}

GET /api/agents/:name
{ /* Agent 对象 */ }

// Skills API
GET /api/skills
{
  skills: Skill[]
}

// Memory API
GET /api/memory
{ /* UserMemory 对象 */ }

// MCP Config API
GET /api/mcp/config
{ /* MCPConfig 对象 */ }

// Upload API
POST /api/threads/:threadId/uploads
{
  success: boolean,
  files: UploadedFileInfo[],
  message: string
}
```

### 4.3 LangGraph API 格式

LangGraph SDK 期望的响应格式需完全兼容 LangGraph Server 的 API 规范。

---

## 5. 实现方案

### 5.1 方案一：双路由组（推荐）

同时维护两套 API：

```
/api/v1/*    - 现有后端 API（保持不变，供管理后台使用）
/api/*        - 新的前端 API（格式与 DeerFlow 前端匹配）
/api/langgraph/* - LangGraph 兼容 API
```

**优点：**
- 保持现有功能不破坏
- 前端 API 完全对齐 DeerFlow
- 渐进式迁移

### 5.2 API 实现优先级

**Phase 0 (P0 - 核心对话)**:
1. `/api/models` - 模型列表
2. `/api/agents` - Agent 列表（返回 `{ agents: [...] }` 格式）
3. `/api/skills` - 技能列表（返回 `{ skills: [...] }` 格式）
4. LangGraph Thread API - 创建、搜索、删除线程
5. LangGraph Runs API - 流式运行

**Phase 1 (P1 - 完善功能)**:
1. `/api/mcp/config` - MCP 配置
2. `/api/threads/:threadId/uploads` - 文件上传
3. 其余 Agent CRUD API
4. 技能启用/禁用

**Phase 2 (P2 - 高级功能)**:
1. `/api/memory` - 用户记忆
2. 技能安装
3. Artifact 文件访问

---

## 6. 文件清单

### 新增文件

| 文件路径 | 说明 |
|---------|------|
| `internal/api/deerflow_handler.go` | DeerFlow 前端 API 处理器 |
| `internal/api/langgraph_handler.go` | LangGraph 兼容 API 处理器 |
| `internal/api/models_handler.go` | Models API 处理器 |
| `internal/api/uploads_handler.go` | Uploads API 处理器 |
| `pkg/langgraph/` | LangGraph SDK 兼容层 |

### 修改文件

| 文件路径 | 说明 |
|---------|------|
| `internal/api/server.go` | 注册新的 API 路由组 |
| `internal/app/gateway.go` | 配置 API 端口（8001） |

---

## 7. 总结

- 前端需要两类 API：普通 REST API (`/api/*`) 和 LangGraph API (`/api/langgraph/*`)
- 建议采用双路由组方案，保持现有 `/api/v1/*` 不变
- 优先实现 Phase 0 的核心对话功能 API
