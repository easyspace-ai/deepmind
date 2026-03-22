package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	apiBase  = "https://ark.cn-beijing.volces.com/api/coding/v3"
	apiKey   = "82c9ade2-b73a-4c5f-8ec6-5c507e0b6036"
	modelName = "doubao-seed-2-0-pro-260215"
)

func main() {
	// 设置 logger
	logger := setupLogger()
	defer logger.Sync()

	logger.Info("🚀 启动 DeerFlow Go Agent 测试")
	logger.Info("配置",
		zap.String("api_base", apiBase),
		zap.String("model", modelName))

	// 创建 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("收到终止信号，正在退出...")
		cancel()
	}()

	// 创建 OpenAI 兼容的 chat model
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: apiBase,
	})
	if err != nil {
		logger.Fatal("创建 chat model 失败", zap.Error(err))
	}
	logger.Info("✅ ChatModel 创建成功")

	// 创建测试工具
	tools := []tool.BaseTool{
		NewWriteFileTool(),
		NewReadFileTool(),
	}
	logger.Info("✅ 工具加载完成", zap.Int("count", len(tools)))

	// 创建 ADK Agent
	agentConfig := &adk.ChatModelAgentConfig{
		Name:        "deerflow_test_agent",
		Description: "测试 Agent，用于验证 DeerFlow Go 的功能",
		Instruction: `你是一个有帮助的 AI 助手。你可以使用以下工具：

1. write_file - 写入文件到本地文件系统
2. read_file - 从本地文件系统读取文件

请根据用户的要求执行相应的操作。如果用户要求生成 HTML，请使用 write_file 工具将其保存为 .html 文件。

执行任务时，请按照以下步骤：
1. 理解用户需求
2. 如果需要生成文件，使用 write_file 工具
3. 完成后给用户总结

重要：不要生成 Markdown 代码块，直接使用工具。`,
		Model: chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		MaxIterations: 10,
	}

	agent, err := adk.NewChatModelAgent(ctx, agentConfig)
	if err != nil {
		logger.Fatal("创建 ADK Agent 失败", zap.Error(err))
	}
	logger.Info("✅ ADK Agent 创建成功")

	// 测试对话
	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════")
	logger.Info("开始测试对话")
	logger.Info("═══════════════════════════════════════════════════")
	logger.Info("")

	// 测试消息
	input := &adk.AgentInput{
		Messages: []*schema.Message{
			schema.UserMessage("你好！请生成一个漂亮的个人主页 HTML，包含：标题、简介、技能展示、联系方式。保存为 portfolio.html"),
		},
	}

	logger.Info("📤 发送消息", zap.String("content", input.Messages[0].Content))

	// 执行 agent
	iterator := agent.Run(ctx, input)
	if iterator == nil {
		logger.Fatal("Agent Run 返回 nil")
	}

	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════")
	logger.Info("✅ Agent 执行开始")
	logger.Info("═══════════════════════════════════════════════════")
	logger.Info("")

	var finalMessage *schema.Message
	var toolCalls []schema.ToolCall

	for {
		event, ok := iterator.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			logger.Error("Agent 事件错误", zap.Error(event.Err))
			break
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg := event.Output.MessageOutput.Message
			if msg != nil {
				logger.Info("📥 收到消息",
					zap.String("role", string(msg.Role)),
					zap.String("content", truncate(msg.Content, 200)))

				if msg.Role == schema.Assistant {
					finalMessage = msg
					toolCalls = msg.ToolCalls
				}
			}
		}
	}

	logger.Info("")
	logger.Info("═══════════════════════════════════════════════════")
	logger.Info("✅ Agent 执行完成")
	logger.Info("═══════════════════════════════════════════════════")
	logger.Info("")

	if finalMessage != nil {
		logger.Info("📝 最终响应", zap.String("content", finalMessage.Content))
	}

	logger.Info("🔧 工具调用", zap.Int("count", len(toolCalls)))
	for i, tc := range toolCalls {
		logger.Info(fmt.Sprintf("  工具 %d:", i+1),
			zap.String("name", tc.Function.Name),
			zap.String("args", tc.Function.Arguments))
	}

	logger.Info("")
	logger.Info("🎉 测试完成！")
}

func setupLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, _ := config.Build()
	return logger
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ============================================
// 测试工具
// ============================================

// WriteFileTool 写入文件工具
type WriteFileTool struct {
	name string
	desc string
}

// NewWriteFileTool 创建写入文件工具
func NewWriteFileTool() *WriteFileTool {
	return &WriteFileTool{
		name: "write_file",
		desc: "写入内容到文件。参数：filename（文件名）, content（内容）",
	}
}

// Info 实现 tool.BaseTool 接口
func (t *WriteFileTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"filename": {
				Desc:     "要写入的文件名",
				Required: true,
				Type:     schema.String,
			},
			"content": {
				Desc:     "文件内容",
				Required: true,
				Type:     schema.String,
			},
		}),
	}, nil
}

// InvokableRun 实现 tool.BaseTool 接口
func (t *WriteFileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	filename, _ := args["filename"].(string)
	content, _ := args["content"].(string)

	if filename == "" {
		return "", fmt.Errorf("filename is required")
	}

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("write file failed: %w", err)
	}

	return fmt.Sprintf("Successfully wrote to %s (%d bytes)", filename, len(content)), nil
}

// ReadFileTool 读取文件工具
type ReadFileTool struct {
	name string
	desc string
}

// NewReadFileTool 创建读取文件工具
func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{
		name: "read_file",
		desc: "读取文件内容。参数：filename（文件名）",
	}
}

// Info 实现 tool.BaseTool 接口
func (t *ReadFileTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"filename": {
				Desc:     "要读取的文件名",
				Required: true,
				Type:     schema.String,
			},
		}),
	}, nil
}

// InvokableRun 实现 tool.BaseTool 接口
func (t *ReadFileTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	filename, _ := args["filename"].(string)

	if filename == "" {
		return "", fmt.Errorf("filename is required")
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("read file failed: %w", err)
	}

	return string(content), nil
}
