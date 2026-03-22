# 0. 文件修改记录表

| 修改人 | 修改时间 | 修改内容 |
| ------ | -------- | -------- |
| AI Assistant | 2026-03-21 | 初始版本 - Phase 1 基础自动规划 |

# 1. 背景（Why）

基于已集成的 Eino 框架，需要实现类似 Claude Code 的自动规划能力。用户输入自然语言后，系统能自动分析意图、分解任务、编排工作流并执行，降低使用 Agent 的门槛。

# 2. 目标（What，必须可验证）

- [ ] 实现意图分析 Agent，能识别任务类型、复杂度、范围
- [ ] 实现任务分解 Agent（基于 Eino DeepAgent），能将复杂任务拆分为子任务
- [ ] 实现基本的工作流编排，支持顺序和简单的并行执行
- [ ] 提供后端 API 供前端调用和显示自动生成的工作流

# 3. 非目标（Explicitly Out of Scope）

- 不实现动态 Agent 生成（留到 Phase 3）
- 不实现执行历史学习与优化（留到 Phase 4）
- 不实现前端 xyflow 工作流可视化（留到后续阶段）
- 不实现 Checkpoint 迁移机制（已有 Eino 原生支持）

# 4. 使用场景 / 用户路径

## 场景 1：基础任务规划

1. 用户通过 API 发送自然语言请求："给登录功能添加单元测试"
2. 系统调用意图分析 Agent，识别任务类型为"test"、复杂度"medium"
3. 系统调用任务分解 Agent，生成子任务列表
4. 系统编排工作流（顺序或并行）
5. 返回工作流结构给调用方

## 场景 2：查看生成的任务

1. 调用方获取生成的任务列表
2. 可以查看每个任务的依赖关系
3. 可以查看建议的执行顺序

# 5. 功能需求清单（Checklist）

- [ ] 意图分析数据模型（IntentAnalysis）
- [ ] 任务数据模型（SubTask）
- [ ] 意图分析 Agent（IntentAnalyzerAgent）
- [ ] 任务分解 Agent（TaskDecomposerAgent）
- [ ] 工作流编排器（WorkflowOrchestrator）
- [ ] 自动规划 API 端点
- [ ] 单元测试覆盖核心功能

# 6. 约束条件（非常关键）

## 技术约束

- 必须基于已集成的 Eino 框架（pkg/agent/eino）
- 必须使用 Go 1.21+
- 必须使用 zap 记录日志
- API 必须遵循现有 REST API 规范

## 架构约束

- 新代码放在 pkg/planner/ 目录下
- 与现有 Agent 管理系统解耦
- 支持扩展到 Phase 2-4

## 安全约束

- API 需要认证（使用现有会话机制）
- 不执行任意代码
- 输入必须验证

# 7. 可修改 / 不可修改项

- ❌ 不可修改：
  - pkg/agent/eino/ 下的现有集成代码
  - 现有 API 认证机制
- ✅ 可调整：
  - 任务分解的提示词
  - 工作流编排策略

# 8. 接口与数据约定（如适用）

## API 定义

### POST /api/v1/planner/analyze

分析用户意图

**请求体：**
```json
{
  "query": "给登录功能添加单元测试"
}
```

**响应体：**
```json
{
  "intent": {
    "taskType": "test",
    "complexity": "medium",
    "scope": "package",
    "technologies": ["go", "testing"],
    "successCriteria": ["覆盖主要登录逻辑", "包含正常和异常场景"]
  }
}
```

### POST /api/v1/planner/decompose

分解任务

**请求体：**
```json
{
  "query": "给登录功能添加单元测试",
  "intent": {...}
}
```

**响应体：**
```json
{
  "tasks": [
    {
      "id": "analyze",
      "name": "分析登录代码",
      "description": "分析现有登录功能代码结构",
      "type": "analyze",
      "dependsOn": [],
      "parallel": false
    }
  ]
}
```

### POST /api/v1/planner/workflow

生成工作流

**请求体：**
```json
{
  "tasks": [...]
}
```

**响应体：**
```json
{
  "workflow": {
    "type": "sequential",
    "stages": [
      {
        "type": "task",
        "task": {...}
      },
      {
        "type": "parallel",
        "tasks": [...]
      }
    ]
  }
}
```

## 数据模型

```go
// IntentAnalysis 意图分析结果
type IntentAnalysis struct {
    TaskType        string   `json:"taskType"`        // 任务类型：code, debug, refactor, document, test, other
    Complexity      string   `json:"complexity"`      // 复杂度：simple, medium, complex
    Scope           string   `json:"scope"`           // 范围：file, package, project
    Technologies    []string `json:"technologies"`    // 涉及的技术栈
    Dependencies    []string `json:"dependencies"`    // 依赖项
    Constraints     []string `json:"constraints"`     // 约束条件
    SuccessCriteria []string `json:"successCriteria"` // 成功标准
}

// SubTask 子任务
type SubTask struct {
    ID          string   `json:"id"`          // 任务唯一标识
    Name        string   `json:"name"`        // 任务名称
    Description string   `json:"description"` // 任务描述
    Type        string   `json:"type"`        // 类型：analyze, create, modify, delete, test, verify
    AgentType   string   `json:"agentType"`   // Agent 类型：chat, tool, custom
    Tools       []string `json:"tools"`       // 需要的工具
    DependsOn   []string `json:"dependsOn"`   // 依赖的任务 ID
    Parallel    bool     `json:"parallel"`    // 是否可并行
}
```

# 9. 验收标准（Acceptance Criteria）

- 如果提供查询"给登录功能添加单元测试"，则意图分析应返回 taskType="test"
- 如果任务有依赖关系，则工作流编排器应按正确顺序排列
- 如果多个任务无依赖且标记为 parallel=true，则应放在同一并行阶段
- 所有 API 端点应返回 200 或恰当的错误码
- 单元测试覆盖率 >= 70%

# 10. 风险与已知不确定点

- 任务分解的质量依赖于 LLM 的能力，可能需要调整提示词
- 遇到歧义时，记录日志并返回需要人工确认的标识
- 初期不实际执行任务，只生成规划

# 11. 非目标（再次强调）

- 不实际执行生成的任务（只规划）
- 不实现前端界面
- 不实现 Agent 动态生成
- 不实现学习优化
