package eino_test

import (
	"context"
	"fmt"

	"github.com/weibaohui/nanobot-go/pkg/agent/eino"
)

// ExampleNewChatModelAgent 展示如何创建一个基本的 ChatModelAgent
func ExampleNewChatModelAgent() {
	ctx := context.Background()

	// 注意：这里需要一个实际的 model.BaseChatModel 实例
	// 例如：使用 eino-ext 的 OpenAI 模型
	// model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
	//     APIKey: "your-api-key",
	//     Model:  "gpt-4",
	// })

	// 创建 ChatModelAgent 配置
	cfg := &eino.ChatModelAgentConfig{
		Name:        "example-agent",
		Description: "一个示例智能体",
		Instruction: "你是一个有帮助的助手。",
		// Model: model, // 实际使用时需要设置模型
		MaxIterations: 10,
	}

	// 创建 Agent（未配置 Model 时 ADK 会返回错误，与生产用法一致）
	agent, err := eino.NewChatModelAgent(ctx, cfg)
	if err != nil {
		fmt.Printf("创建 Agent 失败: %v\n", err)
		// Output: 创建 Agent 失败: agent 'Model' is required
		return
	}

	fmt.Printf("Agent 创建成功: %s\n", agent.Name(ctx))
}

// ExampleNewSequentialAgent 展示如何创建顺序工作流
func ExampleNewSequentialAgent() {
	ctx := context.Background()

	// 注意：实际使用时需要创建真实的子 Agent
	// subAgent1, _ := eino.NewChatModelAgent(ctx, cfg1)
	// subAgent2, _ := eino.NewChatModelAgent(ctx, cfg2)

	cfg := &eino.SequentialAgentConfig{
		Name:        "sequential-workflow",
		Description: "顺序执行工作流",
		// SubAgents: []eino.Agent{subAgent1, subAgent2},
	}

	agent, err := eino.NewSequentialAgent(ctx, cfg)
	if err != nil {
		fmt.Printf("创建顺序工作流失败: %v\n", err)
		return
	}

	fmt.Printf("顺序工作流创建成功: %s\n", agent.Name(ctx))
}

// ExampleNewDeepAgent 展示如何创建 DeepAgent
func ExampleNewDeepAgent() {
	ctx := context.Background()

	cfg := &eino.DeepConfig{
		Name:         "deep-agent",
		Description:  "深度任务编排智能体",
		Instruction:  "你是一个擅长任务分解和编排的智能体。",
		MaxIteration: 20,
		// ChatModel: model, // 实际使用时需要设置模型
	}

	agent, err := eino.NewDeepAgent(ctx, cfg)
	if err != nil {
		fmt.Printf("创建 DeepAgent 失败: %v\n", err)
		return
	}

	fmt.Printf("DeepAgent 创建成功: %s\n", agent.Name(ctx))
}

// ExampleNewSupervisorAgent 展示如何创建监督者智能体系统
func ExampleNewSupervisorAgent() {
	ctx := context.Background()

	// 注意：实际使用时需要创建真实的监督者和子 Agent
	// supervisor, _ := eino.NewChatModelAgent(ctx, supervisorCfg)
	// subAgent1, _ := eino.NewChatModelAgent(ctx, subCfg1)
	// subAgent2, _ := eino.NewChatModelAgent(ctx, subCfg2)

	cfg := &eino.SupervisorConfig{
		// Supervisor: supervisor,
		// SubAgents:  []eino.Agent{subAgent1, subAgent2},
	}

	agent, err := eino.NewSupervisorAgent(ctx, cfg)
	if err != nil {
		fmt.Printf("创建监督者系统失败: %v\n", err)
		return
	}

	// 实际使用时可以获取名称
	_ = agent
	fmt.Println("监督者系统创建示例")
}

// ExampleRunner 展示如何使用 Runner 运行 Agent
func ExampleRunner() {
	_ = context.Background()

	// 注意：实际使用时需要创建真实的 Agent
	// agent, _ := eino.NewChatModelAgent(ctx, cfg)

	// 创建 Runner 配置
	_ = eino.RunnerConfig{
		// Agent: agent,
		EnableStreaming: false,
	}

	// 创建 Runner
	// runner := eino.NewRunner(ctx, runnerCfg)

	// 运行查询
	// iter := runner.Query(ctx, "你好，请介绍一下自己")

	// 迭代处理事件
	// for {
	//     event, ok := iter.Next()
	//     if !ok {
	//         break
	//     }
	//     if event.Err != nil {
	//         fmt.Printf("错误: %v\n", event.Err)
	//         break
	//     }
	//     if event.Output != nil {
	//         fmt.Printf("输出: %s\n", event.Output.Content)
	//     }
	// }

	fmt.Println("Runner 使用示例")
}
