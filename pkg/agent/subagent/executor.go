package subagent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/weibaohui/nanobot-go/pkg/agent/middleware"
	"github.com/weibaohui/nanobot-go/pkg/agent/prompts"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"go.uber.org/zap"
)

// ============================================
// 常量定义
// ============================================

const (
	MaxConcurrentSubagents = 3
	DefaultPollInterval    = 5 * time.Second
)

// ============================================
// 全局线程池
// ============================================

var (
	schedulerPool *WorkerPool
	executionPool *WorkerPool
	poolsOnce     sync.Once
)

// initPools 初始化全局线程池
func initPools() {
	poolsOnce.Do(func() {
		schedulerPool = NewWorkerPool(3, "subagent-scheduler")
		executionPool = NewWorkerPool(3, "subagent-exec")
	})
}

// ============================================
// WorkerPool - 工作线程池
// ============================================

// WorkerPool 简单的工作线程池实现
type WorkerPool struct {
	maxWorkers int
	name       string
	taskQueue  chan func()
	wg         sync.WaitGroup
	stopChan   chan struct{}
	started    bool
	mu         sync.Mutex
	logger     *zap.Logger
}

// NewWorkerPool 创建工作线程池
func NewWorkerPool(maxWorkers int, name string) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		name:       name,
		taskQueue:  make(chan func(), 100),
		stopChan:   make(chan struct{}),
		logger:     zap.NewNop(),
	}
}

// SetLogger 设置日志器
func (p *WorkerPool) SetLogger(logger *zap.Logger) {
	p.logger = logger
}

// Start 启动线程池
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return
	}
	p.started = true

	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	p.logger.Debug("Worker pool started",
		zap.String("name", p.name),
		zap.Int("workers", p.maxWorkers))
}

// Stop 停止线程池
func (p *WorkerPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return
	}

	close(p.stopChan)
	p.wg.Wait()
	p.started = false
	// 下一次 Start() 需要新的 channel，否则再次 Stop 会对已关闭的 channel 执行 close 导致 panic
	p.stopChan = make(chan struct{})

	p.logger.Debug("Worker pool stopped", zap.String("name", p.name))
}

// Submit 提交任务
func (p *WorkerPool) Submit(task func()) {
	p.Start() // 确保已启动
	select {
	case p.taskQueue <- task:
	case <-p.stopChan:
		p.logger.Warn("Worker pool stopped, task rejected", zap.String("name", p.name))
	}
}

// worker 工作协程
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	p.logger.Debug("Worker started",
		zap.String("pool", p.name),
		zap.Int("worker_id", id))

	for {
		select {
		case task := <-p.taskQueue:
			func() {
				defer func() {
					if r := recover(); r != nil {
						p.logger.Error("Worker panicked",
							zap.String("pool", p.name),
							zap.Int("worker_id", id),
							zap.Any("panic", r))
					}
				}()
				task()
			}()
		case <-p.stopChan:
			p.logger.Debug("Worker stopping",
				zap.String("pool", p.name),
				zap.Int("worker_id", id))
			return
		}
	}
}

// ============================================
// SubagentExecutor - 子代理执行器
// ============================================

// SubagentExecutor 子代理执行器
// 一比一复刻 DeerFlow 的 SubagentExecutor
type SubagentExecutor struct {
	config        *SubagentConfig
	parentModel   string
	sandboxState  any // TODO: 替换为实际类型
	threadData    any // TODO: 替换为实际类型
	threadID      string
	traceID       string
	tools         []string
	logger        *zap.Logger

	// chatModel 非空时 Execute 走 Eino ChatModelAgent；与 DeerFlow 子代理对齐。
	chatModel model.ToolCallingChatModel
	// mwConfig 子代理中间件配置；nil 时使用 middleware.DefaultMiddlewareConfig()。
	mwConfig *middleware.MiddlewareConfig
}

// NewSubagentExecutor 创建子代理执行器
func NewSubagentExecutor(
	config *SubagentConfig,
	tools []string,
	parentModel string,
	sandboxState any,
	threadData any,
	threadID string,
	traceID string,
	logger *zap.Logger,
) *SubagentExecutor {
	if logger == nil {
		logger = zap.NewNop()
	}
	if traceID == "" {
		traceID = uuid.New().String()[:8]
	}

	// 过滤工具
	filteredTools := FilterTools(tools, config.Tools, config.DisallowedTools)

	logger.Info("SubagentExecutor initialized",
		zap.String("trace_id", traceID),
		zap.String("name", config.Name),
		zap.Int("tools_count", len(filteredTools)))

	return &SubagentExecutor{
		config:        config,
		parentModel:   parentModel,
		sandboxState:  sandboxState,
		threadData:    threadData,
		threadID:      threadID,
		traceID:       traceID,
		tools:         filteredTools,
		logger:        logger,
	}
}

// WithChatModel 设置 Eino ToolCallingChatModel；设置后 Execute 将创建真实 ChatModelAgent。
func (e *SubagentExecutor) WithChatModel(m model.ToolCallingChatModel) *SubagentExecutor {
	e.chatModel = m
	return e
}

// WithMiddlewareConfig 设置子代理中间件配置（BuildSubagentMiddlewares）。
func (e *SubagentExecutor) WithMiddlewareConfig(c *middleware.MiddlewareConfig) *SubagentExecutor {
	e.mwConfig = c
	return e
}

// createAgent 创建 Eino ChatModelAgent；chatModel 未配置时返回 nil。
func (e *SubagentExecutor) createAgent(ctx context.Context) (*adk.ChatModelAgent, error) {
	if e.chatModel == nil {
		return nil, nil
	}
	maxIter := e.config.MaxTurns
	if maxIter <= 0 {
		maxIter = 25
	}
	instruction := e.config.SystemPrompt
	if instruction == "" {
		instruction = prompts.BuildGeneralPurposeSubagentPromptString()
	}
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          e.config.Name,
		Description:   e.config.Description,
		Instruction:   instruction,
		Model:         e.chatModel,
		MaxIterations: maxIter,
	})
}

// buildInitialState 构建初始状态
func (e *SubagentExecutor) buildInitialState(task string) map[string]any {
	state := map[string]any{
		"messages": []any{task}, // TODO: 使用实际的消息类型
	}

	// 传递 sandbox 和 thread_data
	if e.sandboxState != nil {
		state["sandbox"] = e.sandboxState
	}
	if e.threadData != nil {
		state["thread_data"] = e.threadData
	}

	return state
}

// Execute 同步执行任务
func (e *SubagentExecutor) Execute(task string, resultHolder *SubagentResult) *SubagentResult {
	initPools() // 确保线程池已初始化

	var result *SubagentResult
	if resultHolder != nil {
		result = resultHolder
	} else {
		taskID := uuid.New().String()[:8]
		now := time.Now()
		result = NewSubagentResult(taskID, e.traceID)
		result.Status = SubagentStatusRunning
		result.StartedAt = &now
	}

	e.logger.Info("Subagent starting sync execution",
		zap.String("trace_id", e.traceID),
		zap.String("name", e.config.Name),
		zap.Int("max_turns", e.config.MaxTurns))

	if e.chatModel != nil {
		e.runWithEino(context.Background(), task, result)
		return result
	}

	// 无模型时占位完成（便于离线测试线程池与状态机）
	result.Status = SubagentStatusCompleted
	now := time.Now()
	result.CompletedAt = &now
	result.Result = fmt.Sprintf("Task completed: %s", task)

	e.logger.Info("Subagent completed sync execution",
		zap.String("trace_id", e.traceID),
		zap.String("name", e.config.Name))

	return result
}

// runWithEino 使用 ChatModelAgent + 子代理中间件链（RunBefore/After 各一轮；完整逐轮挂钩需 compose.WithCallbacks 注入 ADK Run）。
func (e *SubagentExecutor) runWithEino(ctx context.Context, task string, result *SubagentResult) {
	subThreadID := fmt.Sprintf("%s-%s", e.threadID, result.TaskID)
	execCtx := middleware.WithThreadID(ctx, subThreadID)

	mwCfg := e.mwConfig
	if mwCfg == nil {
		mwCfg = middleware.DefaultMiddlewareConfig()
	}
	defer func() {
		if mwCfg.SandboxProvider != nil {
			if relErr := mwCfg.SandboxProvider.Release(subThreadID); relErr != nil {
				e.logger.Debug("subagent sandbox release", zap.String("thread_id", subThreadID), zap.Error(relErr))
			}
		}
	}()

	ts := state.NewThreadState()
	if sb, ok := e.sandboxState.(*state.SandboxState); ok && sb != nil {
		ts.Sandbox = sb
	}
	if td, ok := e.threadData.(*state.ThreadDataState); ok && td != nil {
		ts.ThreadData = td
	}

	chain := middleware.BuildSubagentMiddlewares(mwCfg)

	if err := middleware.RunBeforeAgentPhase(execCtx, chain, ts, subThreadID); err != nil {
		e.logger.Warn("subagent RunBeforeAgentPhase", zap.Error(err))
	}
	ts.Messages = []*schema.Message{schema.UserMessage(task)}
	if err := middleware.RunBeforeModelPhase(execCtx, chain, ts); err != nil {
		e.logger.Warn("subagent RunBeforeModelPhase", zap.Error(err))
	}

	agent, err := e.createAgent(execCtx)
	if err != nil {
		result.Status = SubagentStatusFailed
		result.Error = err.Error()
		now := time.Now()
		result.CompletedAt = &now
		return
	}

	runner := adk.NewRunner(execCtx, adk.RunnerConfig{Agent: agent})
	iter := runner.Query(execCtx, task)

	var lastText string
	for {
		ev, ok := iter.Next()
		if !ok {
			break
		}
		if ev.Err != nil {
			result.Status = SubagentStatusFailed
			result.Error = ev.Err.Error()
			now := time.Now()
			result.CompletedAt = &now
			return
		}
		if ev.Output != nil && ev.Output.MessageOutput != nil {
			msg, gerr := ev.Output.MessageOutput.GetMessage()
			if gerr == nil && msg != nil && msg.Content != "" {
				lastText = msg.Content
			}
		}
	}

	if lastText != "" {
		ts.Messages = append(ts.Messages, schema.AssistantMessage(lastText, nil))
	}
	if err := middleware.RunAfterModelPhase(execCtx, chain, ts); err != nil {
		e.logger.Warn("subagent RunAfterModelPhase", zap.Error(err))
	}

	result.Status = SubagentStatusCompleted
	result.Result = lastText
	now := time.Now()
	result.CompletedAt = &now
	e.logger.Info("Subagent Eino execution finished",
		zap.String("trace_id", e.traceID),
		zap.String("name", e.config.Name))
}

// ExecuteAsync 异步启动任务（后台执行）
func (e *SubagentExecutor) ExecuteAsync(task string, taskID string) string {
	initPools() // 确保线程池已初始化

	if taskID == "" {
		taskID = uuid.New().String()[:8]
	}

	// 创建初始结果
	result := NewSubagentResult(taskID, e.traceID)

	e.logger.Info("Subagent starting async execution",
		zap.String("trace_id", e.traceID),
		zap.String("name", e.config.Name),
		zap.String("task_id", taskID),
		zap.Int("timeout_seconds", e.config.TimeoutSeconds))

	// 存储任务
	backgroundTasksLock.Lock()
	backgroundTasks[taskID] = result
	backgroundTasksLock.Unlock()

	// 提交到调度器池
	schedulerPool.Submit(func() {
		// 更新状态为 RUNNING
		backgroundTasksLock.Lock()
		result := backgroundTasks[taskID]
		now := time.Now()
		result.Status = SubagentStatusRunning
		result.StartedAt = &now
		resultHolder := result
		backgroundTasksLock.Unlock()

		// 创建一个 done channel 用于超时控制
		done := make(chan *SubagentResult, 1)

		// 提交到执行池
		executionPool.Submit(func() {
			execResult := e.Execute(task, resultHolder)
			done <- execResult
		})

		// 等待执行完成或超时
		timeout := time.Duration(e.config.TimeoutSeconds) * time.Second
		select {
		case execResult := <-done:
			// 执行完成
			backgroundTasksLock.Lock()
			backgroundTasks[taskID].Status = execResult.Status
			backgroundTasks[taskID].Result = execResult.Result
			backgroundTasks[taskID].Error = execResult.Error
			backgroundTasks[taskID].CompletedAt = execResult.CompletedAt
			backgroundTasks[taskID].AIMessages = execResult.AIMessages
			backgroundTasksLock.Unlock()

			e.logger.Info("Subagent async execution completed",
				zap.String("trace_id", e.traceID),
				zap.String("task_id", taskID),
				zap.String("status", string(execResult.Status)))

		case <-time.After(timeout):
			// 超时
			backgroundTasksLock.Lock()
			backgroundTasks[taskID].Status = SubagentStatusTimedOut
			backgroundTasks[taskID].Error = fmt.Sprintf("Execution timed out after %d seconds", e.config.TimeoutSeconds)
			now := time.Now()
			backgroundTasks[taskID].CompletedAt = &now
			backgroundTasksLock.Unlock()

			e.logger.Error("Subagent async execution timed out",
				zap.String("trace_id", e.traceID),
				zap.String("task_id", taskID),
				zap.Int("timeout_seconds", e.config.TimeoutSeconds))
		}
	})

	return taskID
}

// ============================================
// 便捷函数
// ============================================

// ExecuteSubagent 便捷函数：同步执行子代理
func ExecuteSubagent(
	ctx context.Context,
	config *SubagentConfig,
	task string,
	tools []string,
	parentModel string,
	sandboxState any,
	threadData any,
	threadID string,
	traceID string,
	logger *zap.Logger,
) (*SubagentResult, error) {

	executor := NewSubagentExecutor(
		config,
		tools,
		parentModel,
		sandboxState,
		threadData,
		threadID,
		traceID,
		logger,
	)

	result := executor.Execute(task, nil)

	if result.Status == SubagentStatusFailed {
		return result, fmt.Errorf("subagent failed: %s", result.Error)
	}

	return result, nil
}

// ExecuteSubagentAsync 便捷函数：异步执行子代理
func ExecuteSubagentAsync(
	config *SubagentConfig,
	task string,
	tools []string,
	parentModel string,
	sandboxState any,
	threadData any,
	threadID string,
	traceID string,
	logger *zap.Logger,
) (string, error) {

	executor := NewSubagentExecutor(
		config,
		tools,
		parentModel,
		sandboxState,
		threadData,
		threadID,
		traceID,
		logger,
	)

	taskID := executor.ExecuteAsync(task, "")
	return taskID, nil
}

// ============================================
// 线程池管理
// ============================================

// StartPools 启动所有线程池
func StartPools() {
	initPools()
	schedulerPool.Start()
	executionPool.Start()
}

// StopPools 停止所有线程池
func StopPools() {
	if executionPool != nil {
		executionPool.Stop()
	}
	if schedulerPool != nil {
		schedulerPool.Stop()
	}
}

// SetPoolsLogger 设置线程池日志器
func SetPoolsLogger(logger *zap.Logger) {
	initPools()
	schedulerPool.SetLogger(logger)
	executionPool.SetLogger(logger)
}
