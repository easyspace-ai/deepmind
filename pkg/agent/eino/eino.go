// Package eino 提供 Eino 框架的便捷集成
//
// 这个包提供了 Eino 框架核心类型的便捷导出，
// 使得在 nanobot-go 项目中使用 Eino 更加方便。
//
// 主要特性：
//   - 中间件优先：通过 ChatModelAgentMiddleware 接口无限扩展
//   - 中断可组合：CompositeInterrupt 支持嵌套智能体中断
//   - 恢复目标化：ResumeWithParams 可精确寻址恢复点
//   - 工作流内置：顺序/并行/循环开箱即用
//   - 状态向前兼容：Checkpoint 迁移机制
//
// 基本用法：
//
//	import "github.com/weibaohui/nanobot-go/pkg/agent/eino"
//
//	// 创建一个 ChatModelAgent
//	agent, err := eino.NewChatModelAgent(ctx, &eino.ChatModelAgentConfig{
//	    Name:        "my-agent",
//	    Description: "A helpful agent",
//	    Model:       myModel,
//	    Instruction: "You are a helpful assistant.",
//	})
//
//	// 创建 Runner 并运行
//	runner := eino.NewRunner(ctx, eino.RunnerConfig{
//	    Agent: agent,
//	})
//	iter := runner.Query(ctx, "Hello!")
package eino

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ==================== 核心类型导出 ====================

type (
	// Agent Eino Agent 接口
	Agent = adk.Agent

	// ResumableAgent 可恢复的 Agent
	ResumableAgent = adk.ResumableAgent

	// ChatModelAgent 基于聊天模型的 Agent
	ChatModelAgent = adk.ChatModelAgent

	// AgentEvent Agent 事件
	AgentEvent = adk.AgentEvent

	// AgentAction Agent 动作
	AgentAction = adk.AgentAction

	// AgentInput Agent 输入
	AgentInput = adk.AgentInput

	// AgentOutput Agent 输出
	AgentOutput = adk.AgentOutput

	// Message 消息类型
	Message = schema.Message

	// AsyncIterator 异步迭代器
	AsyncIterator[T any] = adk.AsyncIterator[T]

	// Runner Agent 运行器
	Runner = adk.Runner

	// RunnerConfig 运行器配置
	RunnerConfig = adk.RunnerConfig

	// ResumeParams 恢复参数
	ResumeParams = adk.ResumeParams

	// ChatModelAgentConfig ChatModelAgent 配置
	ChatModelAgentConfig = adk.ChatModelAgentConfig

	// AgentMiddleware Agent 中间件（结构体形式）
	AgentMiddleware = adk.AgentMiddleware

	// ChatModelAgentState ChatModelAgent 状态
	ChatModelAgentState = adk.ChatModelAgentState

	// ToolsConfig 工具配置
	ToolsConfig = adk.ToolsConfig

	// CheckPointStore 检查点存储
	CheckPointStore = adk.CheckPointStore

	// ResumeInfo 恢复信息
	ResumeInfo = adk.ResumeInfo

	// InterruptInfo 中断信息
	InterruptInfo = adk.InterruptInfo

	// InterruptCtx 中断上下文
	InterruptCtx = adk.InterruptCtx

	// RoleType 角色类型
	RoleType = schema.RoleType

	// ToolCall 工具调用
	ToolCall = schema.ToolCall

	// BaseChatModel 基础聊天模型接口
	BaseChatModel = model.BaseChatModel

	// ToolCallingChatModel 支持工具调用的聊天模型接口
	ToolCallingChatModel = model.ToolCallingChatModel

	// BaseTool 基础工具接口
	BaseTool = tool.BaseTool

	// DeepConfig DeepAgent 配置
	DeepConfig = deep.Config

	// PlanExecuteConfig PlanExecute 配置
	PlanExecuteConfig = planexecute.Config

	// PlannerConfig Planner 配置
	PlannerConfig = planexecute.PlannerConfig

	// ExecutorConfig Executor 配置
	ExecutorConfig = planexecute.ExecutorConfig

	// ReplannerConfig Replanner 配置
	ReplannerConfig = planexecute.ReplannerConfig

	// SupervisorConfig Supervisor 配置
	SupervisorConfig = supervisor.Config

	// SequentialAgentConfig 顺序 Agent 配置
	SequentialAgentConfig = adk.SequentialAgentConfig

	// ParallelAgentConfig 并行 Agent 配置
	ParallelAgentConfig = adk.ParallelAgentConfig

	// LoopAgentConfig 循环 Agent 配置
	LoopAgentConfig = adk.LoopAgentConfig
)

// ==================== 常量导出 ====================

const (
	// RoleUser 用户角色
	RoleUser = schema.User
	// RoleAssistant 助手角色
	RoleAssistant = schema.Assistant
	// RoleSystem 系统角色
	RoleSystem = schema.System
	// RoleTool 工具角色
	RoleTool = schema.Tool
)

// ==================== 基础构造函数 ====================

// NewChatModelAgent 创建 ChatModelAgent
func NewChatModelAgent(ctx context.Context, cfg *ChatModelAgentConfig) (*ChatModelAgent, error) {
	return adk.NewChatModelAgent(ctx, cfg)
}

// NewRunner 创建 Runner
func NewRunner(ctx context.Context, cfg RunnerConfig) *Runner {
	return adk.NewRunner(ctx, cfg)
}

// UserMessage 创建用户消息
func UserMessage(content string) *Message {
	return schema.UserMessage(content)
}

// SystemMessage 创建系统消息
func SystemMessage(content string) *Message {
	return schema.SystemMessage(content)
}

// AssistantMessage 创建助手消息
func AssistantMessage(content string, toolCalls []ToolCall) *Message {
	return schema.AssistantMessage(content, toolCalls)
}

// ToolMessage 创建工具消息
func ToolMessage(content string, toolCallID string) *Message {
	return schema.ToolMessage(content, toolCallID)
}

// ==================== 工作流 Agent ====================

// NewSequentialAgent 创建顺序执行 Agent
func NewSequentialAgent(ctx context.Context, cfg *SequentialAgentConfig) (ResumableAgent, error) {
	return adk.NewSequentialAgent(ctx, cfg)
}

// NewParallelAgent 创建并行执行 Agent
func NewParallelAgent(ctx context.Context, cfg *ParallelAgentConfig) (ResumableAgent, error) {
	return adk.NewParallelAgent(ctx, cfg)
}

// NewLoopAgent 创建循环执行 Agent
func NewLoopAgent(ctx context.Context, cfg *LoopAgentConfig) (ResumableAgent, error) {
	return adk.NewLoopAgent(ctx, cfg)
}

// NewBreakLoopAction 创建中断循环的动作
func NewBreakLoopAction(agentName string) *AgentAction {
	return adk.NewBreakLoopAction(agentName)
}

// NewExitAction 创建退出动作
func NewExitAction() *AgentAction {
	return adk.NewExitAction()
}

// NewTransferToAgentAction 创建移交到其他 Agent 的动作
func NewTransferToAgentAction(destAgentName string) *AgentAction {
	return adk.NewTransferToAgentAction(destAgentName)
}

// ==================== 预构建 Agent ====================

// NewDeepAgent 创建 DeepAgent
func NewDeepAgent(ctx context.Context, cfg *DeepConfig) (ResumableAgent, error) {
	return deep.New(ctx, cfg)
}

// NewPlanExecuteAgent 创建 PlanExecuteAgent
func NewPlanExecuteAgent(ctx context.Context, cfg *PlanExecuteConfig) (ResumableAgent, error) {
	return planexecute.New(ctx, cfg)
}

// NewPlanner 创建规划者 Agent
func NewPlanner(ctx context.Context, cfg *PlannerConfig) (Agent, error) {
	return planexecute.NewPlanner(ctx, cfg)
}

// NewExecutor 创建执行者 Agent
func NewExecutor(ctx context.Context, cfg *ExecutorConfig) (Agent, error) {
	return planexecute.NewExecutor(ctx, cfg)
}

// NewReplanner 创建重规划者 Agent
func NewReplanner(ctx context.Context, cfg *ReplannerConfig) (Agent, error) {
	return planexecute.NewReplanner(ctx, cfg)
}

// NewSupervisorAgent 创建监督者 Agent
func NewSupervisorAgent(ctx context.Context, cfg *SupervisorConfig) (ResumableAgent, error) {
	return supervisor.New(ctx, cfg)
}

// ==================== 中断系统 ====================

// Interrupt 创建基本中断
func Interrupt(ctx context.Context, info any) *AgentEvent {
	return adk.Interrupt(ctx, info)
}

// StatefulInterrupt 创建带状态的中断
func StatefulInterrupt(ctx context.Context, info any, state any) *AgentEvent {
	return adk.StatefulInterrupt(ctx, info, state)
}

// WithCheckPointID 设置检查点 ID
func WithCheckPointID(id string) adk.AgentRunOption {
	return adk.WithCheckPointID(id)
}

// ==================== 会话值 ====================

// AddSessionValue 添加会话值
func AddSessionValue(ctx context.Context, key string, value any) {
	adk.AddSessionValue(ctx, key, value)
}

// AddSessionValues 批量添加会话值
func AddSessionValues(ctx context.Context, values map[string]any) {
	adk.AddSessionValues(ctx, values)
}

// GetSessionValue 获取会话值
func GetSessionValue(ctx context.Context, key string) (any, bool) {
	return adk.GetSessionValue(ctx, key)
}

// GetSessionValues 获取所有会话值
func GetSessionValues(ctx context.Context) map[string]any {
	return adk.GetSessionValues(ctx)
}

// ==================== 工具函数 ====================

// SendEvent 发送 Agent 事件
func SendEvent(ctx context.Context, event *AgentEvent) error {
	return adk.SendEvent(ctx, event)
}

// SendToolGenAction 发送工具生成的动作
func SendToolGenAction(ctx context.Context, toolName string, action *AgentAction) error {
	return adk.SendToolGenAction(ctx, toolName, action)
}

// ==================== Agent 工具 ====================

// AgentWithDeterministicTransferTo 创建具有确定性移交的 Agent
func AgentWithDeterministicTransferTo(ctx context.Context, agent Agent, toAgentNames []string) Agent {
	return adk.AgentWithDeterministicTransferTo(ctx, &adk.DeterministicTransferConfig{
		Agent:        agent,
		ToAgentNames: toAgentNames,
	})
}

// SetSubAgents 设置子 Agent
func SetSubAgents(ctx context.Context, parent Agent, subAgents []Agent) (ResumableAgent, error) {
	return adk.SetSubAgents(ctx, parent, subAgents)
}
