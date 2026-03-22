# Eino 框架集成

这个包提供了 [Eino](https://github.com/cloudwego/eino) 框架的便捷集成，使得在 nanobot-go 项目中使用 Eino 更加方便。

## 主要特性

- **中间件优先**: 通过 ChatModelAgentMiddleware 接口无限扩展
- **中断可组合**: CompositeInterrupt 支持嵌套智能体中断
- **恢复目标化**: ResumeWithParams 可精确寻址恢复点
- **工作流内置**: 顺序/并行/循环开箱即用
- **状态向前兼容**: Checkpoint 迁移机制

## 快速开始

### 1. 基础使用

```go
import "github.com/weibaohui/nanobot-go/pkg/agent/eino"

// 创建一个 ChatModelAgent
agent, err := eino.NewChatModelAgent(ctx, &eino.ChatModelAgentConfig{
    Name:        "my-agent",
    Description: "A helpful agent",
    Model:       myModel, // 需要实现 model.BaseChatModel
    Instruction: "You are a helpful assistant.",
    MaxIterations: 20,
})
```

### 2. 创建 Runner 并运行

```go
// 创建 Runner
runner := eino.NewRunner(ctx, eino.RunnerConfig{
    Agent:           agent,
    EnableStreaming: false,
})

// 运行查询
iter := runner.Query(ctx, "Hello!")

// 处理事件
for {
    event, ok := iter.Next()
    if !ok {
        break
    }
    if event.Err != nil {
        log.Printf("Error: %v", event.Err)
        break
    }
    if event.Output != nil {
        fmt.Printf("Output: %s\n", event.Output.Content)
    }
}
```

### 3. 顺序工作流

```go
// 创建子 Agent
agent1, _ := eino.NewChatModelAgent(ctx, cfg1)
agent2, _ := eino.NewChatModelAgent(ctx, cfg2)

// 创建顺序工作流
workflow, err := eino.NewSequentialAgent(ctx, &eino.SequentialAgentConfig{
    Name:        "sequential-workflow",
    Description: "按顺序执行的工作流",
    SubAgents:   []eino.Agent{agent1, agent2},
})
```

### 4. 并行工作流

```go
workflow, err := eino.NewParallelAgent(ctx, &eino.ParallelAgentConfig{
    Name:        "parallel-workflow",
    Description: "并行执行的工作流",
    SubAgents:   []eino.Agent{agent1, agent2},
})
```

### 5. 循环工作流

```go
workflow, err := eino.NewLoopAgent(ctx, &eino.LoopAgentConfig{
    Name:          "loop-workflow",
    Description:   "循环执行的工作流",
    SubAgents:     []eino.Agent{agent},
    MaxIterations: 10,
})
```

### 6. DeepAgent (深度任务编排)

```go
deepAgent, err := eino.NewDeepAgent(ctx, &eino.DeepConfig{
    Name:         "deep-agent",
    Description:  "深度任务编排智能体",
    ChatModel:    myModel,
    Instruction:  "你是一个擅长任务分解的智能体。",
    SubAgents:    []eino.Agent{specialistAgent1, specialistAgent2},
    MaxIteration: 20,
})
```

### 7. PlanExecute (计划-执行-重计划)

```go
// 创建各阶段 Agent
planner, _ := eino.NewPlanner(ctx, &eino.PlannerConfig{
    ToolCallingChatModel: myToolCallingModel,
})

executor, _ := eino.NewExecutor(ctx, &eino.ExecutorConfig{
    Model:         myModel,
    Tools:         myTools,
    MaxIterations: 10,
})

replanner, _ := eino.NewReplanner(ctx, &eino.ReplannerConfig{
    ChatModel: myToolCallingModel,
})

// 创建 PlanExecute Agent
planExecuteAgent, err := eino.NewPlanExecuteAgent(ctx, &eino.PlanExecuteConfig{
    Planner:   planner,
    Executor:  executor,
    Replanner: replanner,
    MaxIterations: 10,
})
```

### 8. Supervisor (监督者模式)

```go
// 创建监督者和子 Agent
supervisorAgent, _ := eino.NewChatModelAgent(ctx, supervisorCfg)
subAgent1, _ := eino.NewChatModelAgent(ctx, subCfg1)
subAgent2, _ := eino.NewChatModelAgent(ctx, subCfg2)

// 创建监督者系统
supervisorSystem, err := eino.NewSupervisorAgent(ctx, &eino.SupervisorConfig{
    Supervisor: supervisorAgent,
    SubAgents:  []eino.Agent{subAgent1, subAgent2},
})
```

### 9. 使用中间件

```go
// 简单中间件
simpleMw := eino.NewSimpleMiddleware(&eino.SimpleMiddlewareConfig{
    AdditionalInstruction: "额外的系统提示词",
    AdditionalTools:       []eino.BaseTool{myTool},
})

// 在配置中使用
cfg := &eino.ChatModelAgentConfig{
    // ...
    Middlewares: []eino.AgentMiddleware{simpleMw},
}
```

### 10. 中断与恢复

```go
// 发送中断
interruptEvent := eino.Interrupt(ctx, "需要用户确认")

// 带状态的中断
statefulEvent := eino.StatefulInterrupt(ctx, "需要用户确认", myState)

// 使用 Checkpoint
runner := eino.NewRunner(ctx, eino.RunnerConfig{
    Agent:           agent,
    CheckPointStore: myCheckpointStore,
})

// 恢复执行
iter, err := runner.Resume(ctx, checkpointID)
// 或者使用 ResumeWithParams 精确恢复
iter, err := runner.ResumeWithParams(ctx, checkpointID, &eino.ResumeParams{
    Targets: map[string]any{
        "agent-address": resumeData,
    },
})
```

### 11. 会话值

```go
// 添加会话值
eino.AddSessionValue(ctx, "user_id", "12345")

// 批量添加
eino.AddSessionValues(ctx, map[string]any{
    "user_id": "12345",
    "time":    time.Now(),
})

// 获取会话值
value, ok := eino.GetSessionValue(ctx, "user_id")

// 获取所有会话值
allValues := eino.GetSessionValues(ctx)
```

## 更多示例

请查看 [example_test.go](./example_test.go) 获取更多使用示例。

## 相关链接

- [Eino GitHub](https://github.com/cloudwego/eino)
- [Eino 文档](https://github.com/cloudwego/eino#readme)
