package middleware

import (
	"context"
	"errors"

	"github.com/cloudwego/eino/callbacks"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// 中间件错误定义（一比一复刻 DeerFlow 错误行为）
var (
	// ErrThreadIDRequired thread_id is required in the context
	ErrThreadIDRequired = errors.New("thread ID is required in the context")
)

// Middleware DeerFlow 风格中间件接口
// 一比一复刻 DeerFlow 的 AgentMiddleware 概念，但适配 Eino
type Middleware interface {
	// Name 返回中间件名称
	Name() string
}

// BaseMiddleware 基础中间件实现
type BaseMiddleware struct {
	name string
}

// NewBaseMiddleware 创建基础中间件
func NewBaseMiddleware(name string) *BaseMiddleware {
	return &BaseMiddleware{name: name}
}

// Name 实现 Middleware 接口
func (m *BaseMiddleware) Name() string {
	return m.name
}

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []Middleware
}

// NewMiddlewareChain 创建中间件链
func NewMiddlewareChain(middlewares ...Middleware) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: middlewares,
	}
}

// Add 添加中间件
func (c *MiddlewareChain) Add(m Middleware) *MiddlewareChain {
	c.middlewares = append(c.middlewares, m)
	return c
}

// AddAll 添加多个中间件
func (c *MiddlewareChain) AddAll(middlewares ...Middleware) *MiddlewareChain {
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}

// Middlewares 获取所有中间件
func (c *MiddlewareChain) Middlewares() []Middleware {
	return c.middlewares
}

// EinoCallbackBridge Eino 回调桥接器
type EinoCallbackBridge struct {
	chain   *MiddlewareChain
	state   *state.ThreadState
	handler callbacks.Handler
	logger  *zap.Logger
}

// NewEinoCallbackBridge 创建 Eino 回调桥接器
func NewEinoCallbackBridge(chain *MiddlewareChain, state *state.ThreadState) *EinoCallbackBridge {
	return &EinoCallbackBridge{
		chain:  chain,
		state:  state,
		logger: zap.NewNop(),
	}
}

// SetLogger 设置日志（用于 DeerFlow 中间件链 Eino 回调）。
func (b *EinoCallbackBridge) SetLogger(l *zap.Logger) *EinoCallbackBridge {
	if l == nil {
		b.logger = zap.NewNop()
	} else {
		b.logger = l
	}
	return b
}

// SetHandler 设置底层 handler
func (b *EinoCallbackBridge) SetHandler(handler callbacks.Handler) {
	b.handler = handler
}

// Handler 获取组合后的 handler
func (b *EinoCallbackBridge) Handler() callbacks.Handler {
	if b.handler != nil {
		return b.handler
	}
	return NewDeerflowEinoHandler(b.chain, b.state, b.logger)
}

// GetState 获取当前状态
func (b *EinoCallbackBridge) GetState() *state.ThreadState {
	return b.state
}

// ============================================
// 完整中间件接口（DeerFlow 风格）
// ============================================

// BeforeAgentMiddleware 在 Agent 执行前运行的中间件
type BeforeAgentMiddleware interface {
	Middleware
	BeforeAgent(ctx context.Context, state *state.ThreadState) (stateUpdate map[string]interface{}, err error)
}

// AfterModelMiddleware 在模型响应后运行的中间件
type AfterModelMiddleware interface {
	Middleware
	AfterModel(ctx context.Context, state *state.ThreadState) (stateUpdate map[string]interface{}, err error)
}

// BeforeModelMiddleware 在模型调用前运行的中间件
type BeforeModelMiddleware interface {
	Middleware
	BeforeModel(ctx context.Context, state *state.ThreadState) (stateUpdate map[string]interface{}, err error)
}

// ToolCallInfo 工具调用信息（用于所有中间件）
type ToolCallInfo struct {
	ID   string                 `json:"id"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// ToolResultInfo 工具结果信息
type ToolResultInfo struct {
	ToolCallID string `json:"tool_call_id"`
}

// BeforeAgentWithIDMiddleware 带 threadID 的 BeforeAgentMiddleware
type BeforeAgentWithIDMiddleware interface {
	Middleware
	BeforeAgent(ctx context.Context, state *state.ThreadState, threadID string) (stateUpdate map[string]interface{}, err error)
}

// WrapToolCallMiddleware 包装工具调用的中间件
type WrapToolCallMiddleware interface {
	Middleware
	WrapToolCall(ctx context.Context, toolCall *ToolCallInfo, next ToolCallHandler) (*ToolCallResult, error)
}

// ToolCallHandler 工具调用处理器
type ToolCallHandler func(ctx context.Context, req *ToolCallInfo) (*ToolCallResult, error)

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	Content     string
	StateUpdate map[string]interface{}
	Interrupt   bool // 是否中断执行
}

// WrapModelCallMiddleware 包装模型调用的中间件
type WrapModelCallMiddleware interface {
	Middleware
	WrapModelCall(ctx context.Context, req *ModelCallRequest, next ModelCallHandler) (*ModelCallResult, error)
}

// ModelCallRequest 模型调用请求
type ModelCallRequest struct {
	Messages []*state.ThreadState
}

// ModelCallHandler 模型调用处理器
type ModelCallHandler func(ctx context.Context, req *ModelCallRequest) (*ModelCallResult, error)

// ModelCallResult 模型调用结果
type ModelCallResult struct {
	StateUpdate map[string]interface{}
}
