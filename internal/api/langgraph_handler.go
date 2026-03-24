package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/weibaohui/nanobot-go/pkg/agent"
	"github.com/weibaohui/nanobot-go/pkg/agent/provider"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/sse/deerflow"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ============================================
// LangGraph 兼容 API 处理器
// 一比一复刻 DeerFlow 的 SSE 事件格式
// ============================================

// LangGraphHandler LangGraph 兼容 API 处理器
type LangGraphHandler struct {
	logger       *zap.Logger
	threadStore  *ThreadStore
	runStore     *RunStore
	loop         *agent.Loop
	db           *gorm.DB
	configLoader provider.LLMConfigLoader
}

// NewLangGraphHandler 创建 LangGraph 处理器
func NewLangGraphHandler(logger *zap.Logger) *LangGraphHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	h := &LangGraphHandler{
		logger:      logger,
		threadStore: NewThreadStore(),
		runStore:    NewRunStore(),
	}

	// 直接设置测试用的 ConfigLoader（简化配置，不从数据库读取）
	h.configLoader = func(ctx context.Context) (*provider.LLMConfig, error) {
		return &provider.LLMConfig{
			APIKey:       "82c9ade2-b73a-4c5f-8ec6-5c507e0b6036",
			APIBase:      "https://ark.cn-beijing.volces.com/api/coding/v3",
			DefaultModel: "doubao-seed-2-0-pro-260215",
		}, nil
	}

	logger.Info("LangGraphHandler 已初始化，使用测试配置")
	return h
}

// SetLoop 注入 Agent Loop（用于真实 Lead Agent 集成）
func (h *LangGraphHandler) SetLoop(loop *agent.Loop) {
	h.loop = loop
	h.logger.Info("Agent Loop 已注入到 LangGraphHandler")
}

// SetDB 注入数据库连接（用于直接加载 LLM 配置）
func (h *LangGraphHandler) SetDB(db *gorm.DB) {
	h.db = db
	h.logger.Info("数据库连接已注入到 LangGraphHandler")
}

// SetConfigLoader 注入配置加载器（用于直接加载 LLM 配置）
func (h *LangGraphHandler) SetConfigLoader(loader provider.LLMConfigLoader) {
	h.configLoader = loader
	h.logger.Info("ConfigLoader 已注入到 LangGraphHandler")
}

// RegisterRoutes 注册 LangGraph 路由
func (h *LangGraphHandler) RegisterRoutes(router *gin.Engine) {
	lg := router.Group("/api/langgraph")
	{
		// Threads API
		lg.POST("/threads", h.createThread)
		lg.GET("/threads/:threadId", h.getThread)
		lg.GET("/threads/:threadId/state", h.getThreadState)
		lg.POST("/threads/search", h.searchThreads)
		lg.DELETE("/threads/:threadId", h.deleteThread)
		lg.POST("/threads/:threadId/state", h.updateThreadState)
		lg.POST("/threads/:threadId/history", h.getThreadHistory)

		// Runs API
		lg.POST("/threads/:threadId/runs/stream", h.streamRun)
		lg.POST("/threads/:threadId/runs/:runId/join", h.joinStream)
	}
}

// ============================================
// 辅助函数 - 消息序列化（对齐原版 DeerFlow）
// ============================================

// serializeMessage 序列化消息为 values 事件使用的格式
func serializeMessage(msg interface{}) map[string]interface{} {
	if msgMap, ok := msg.(map[string]interface{}); ok {
		return msgMap
	}
	// 默认返回空消息
	return map[string]interface{}{
		"type":    "unknown",
		"content": fmt.Sprintf("%v", msg),
	}
}

// serializeMessageForTuple 序列化消息为 messages-tuple 事件使用的格式
func serializeMessageForTuple(msg interface{}) map[string]interface{} {
	if msgMap, ok := msg.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		// 复制基本字段
		for k, v := range msgMap {
			result[k] = v
		}
		return result
	}
	return map[string]interface{}{
		"type":    "unknown",
		"content": fmt.Sprintf("%v", msg),
	}
}

// buildValuesEvent 构建 values 事件数据（完整状态快照）
// 一比一复刻原版 DeerFlow 的 values 事件格式
func buildValuesEvent(thread *Thread) map[string]interface{} {
	values := thread.Values

	// 确保包含所有必需字段
	result := make(map[string]interface{})

	// 复制现有字段
	for k, v := range values {
		result[k] = v
	}

	// 确保 messages 存在
	if _, ok := result["messages"]; !ok {
		result["messages"] = []interface{}{}
	}

	// 确保 artifacts 存在
	if _, ok := result["artifacts"]; !ok {
		result["artifacts"] = []string{}
	}

	// 确保 todos 存在
	if _, ok := result["todos"]; !ok {
		result["todos"] = []state.TodoItem{}
	}

	// 确保 sandbox 存在（可为 nil）
	if _, ok := result["sandbox"]; !ok {
		result["sandbox"] = nil
	}

	// 确保 thread_data 存在（可为 nil）
	if _, ok := result["thread_data"]; !ok {
		result["thread_data"] = nil
	}

	// 确保 uploaded_files 存在
	if _, ok := result["uploaded_files"]; !ok {
		result["uploaded_files"] = []state.UploadedFile{}
	}

	// 确保 viewed_images 存在
	if _, ok := result["viewed_images"]; !ok {
		result["viewed_images"] = map[string]state.ViewedImageData{}
	}

	// 确保 clarification_pending 存在
	if _, ok := result["clarification_pending"]; !ok {
		result["clarification_pending"] = false
	}

	// title 可能为 nil 或字符串
	if title, ok := result["title"]; !ok || title == nil {
		result["title"] = nil
	}

	return result
}

// ============================================
// 辅助函数 - 其他
// ============================================

// extractMessageText 从消息数组中提取第一条人类消息的文本
func extractMessageText(messages []interface{}) string {
	for _, msg := range messages {
		if msgMap, ok := msg.(map[string]interface{}); ok {
			// 检查是否是人类消息
			if msgType, _ := msgMap["type"].(string); msgType == "human" || msgType == "user" {
				// 提取文本内容
				if content, ok := msgMap["content"]; ok {
					// content 可能是字符串
					if text, ok := content.(string); ok && text != "" {
						return text
					}
					// content 也可能是数组
					if parts, ok := content.([]interface{}); ok {
						for _, part := range parts {
							if partMap, ok := part.(map[string]interface{}); ok {
								if partType, _ := partMap["type"].(string); partType == "text" {
									if text, _ := partMap["text"].(string); text != "" {
										return text
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}

// normalizeContent 标准化内容（类似原版 TitleMiddleware）
func normalizeContent(content interface{}) string {
	if content == nil {
		return ""
	}
	if str, ok := content.(string); ok {
		return str
	}
	if list, ok := content.([]interface{}); ok {
		var parts []string
		for _, item := range list {
			parts = append(parts, normalizeContent(item))
		}
		return strings.Join(parts, "\n")
	}
	if mp, ok := content.(map[string]interface{}); ok {
		if text, ok := mp["text"].(string); ok {
			return text
		}
		if nested, ok := mp["content"]; ok {
			return normalizeContent(nested)
		}
	}
	return fmt.Sprintf("%v", content)
}

// shouldGenerateTitle 检查是否应该生成标题（类似原版 TitleMiddleware）
func shouldGenerateTitle(values map[string]interface{}) bool {
	// 检查是否已经有标题
	if title, hasTitle := values["title"]; hasTitle && title != "" && title != nil {
		return false
	}

	// 检查 messages
	messages, ok := values["messages"].([]interface{})
	if !ok {
		return false
	}

	// 检查是否是第一次完整对话（至少 1 条用户消息 + 1 条 AI 响应）
	if len(messages) < 2 {
		return false
	}

	// 统计用户和 AI 消息
	userCount := 0
	assistantCount := 0
	for _, msg := range messages {
		if msgMap, ok := msg.(map[string]interface{}); ok {
			msgType, _ := msgMap["type"].(string)
			if msgType == "human" || msgType == "user" {
				userCount++
			} else if msgType == "ai" || msgType == "assistant" {
				assistantCount++
			}
		}
	}

	// 第一次完整对话后生成标题
	return userCount == 1 && assistantCount >= 1
}

// extractFirstMessages 提取第一条用户消息和第一条 AI 消息
func extractFirstMessages(values map[string]interface{}) (userMsg string, assistantMsg string) {
	messages, ok := values["messages"].([]interface{})
	if !ok {
		return "", ""
	}

	for _, msg := range messages {
		if msgMap, ok := msg.(map[string]interface{}); ok {
			msgType, _ := msgMap["type"].(string)
			content := msgMap["content"]
			if (msgType == "human" || msgType == "user") && userMsg == "" {
				userMsg = normalizeContent(content)
			} else if (msgType == "ai" || msgType == "assistant") && assistantMsg == "" {
				assistantMsg = normalizeContent(content)
			}
		}
	}
	return userMsg, assistantMsg
}

// generateFallbackTitle 生成回退标题（类似原版 TitleMiddleware）
func generateFallbackTitle(userMsg string) string {
	maxChars := 50
	runes := []rune(userMsg)
	if len(runes) > maxChars {
		return string(runes[:maxChars]) + "..."
	}
	if userMsg == "" {
		return "New Conversation"
	}
	return userMsg
}

// ============================================
// Thread 数据结构
// ============================================

// Thread 线程
type Thread struct {
	ThreadID  string                 `json:"thread_id"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
	Values    map[string]interface{} `json:"values"`
}

// ThreadStore 线程存储（内存实现）
type ThreadStore struct {
	mu      sync.RWMutex
	threads map[string]*Thread
}

func NewThreadStore() *ThreadStore {
	return &ThreadStore{
		threads: make(map[string]*Thread),
	}
}

func (s *ThreadStore) Create(threadID string) *Thread {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	thread := &Thread{
		ThreadID:  threadID,
		CreatedAt: now,
		UpdatedAt: now,
		Values: map[string]interface{}{
			"messages":  []interface{}{},
			"artifacts": []string{},
			"todos":     []state.TodoItem{},
		},
	}
	s.threads[threadID] = thread
	return thread
}

func (s *ThreadStore) Get(threadID string) (*Thread, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.threads[threadID]
	return t, ok
}

func (s *ThreadStore) List() []*Thread {
	s.mu.RLock()
	defer s.mu.RUnlock()
	threads := make([]*Thread, 0, len(s.threads))
	for _, t := range s.threads {
		threads = append(threads, t)
	}
	return threads
}

func (s *ThreadStore) Update(threadID string, values map[string]interface{}) *Thread {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.threads[threadID]
	if !ok {
		return nil
	}
	t.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if values != nil {
		for k, v := range values {
			t.Values[k] = v
		}
	}
	return t
}

func (s *ThreadStore) Delete(threadID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.threads[threadID]
	if ok {
		delete(s.threads, threadID)
	}
	return ok
}

// ============================================
// Run 数据结构
// ============================================

// Run 运行
type Run struct {
	RunID     string
	ThreadID  string
	Status    string
	CreatedAt string
}

// RunStore 运行存储
type RunStore struct {
	mu   sync.RWMutex
	runs map[string]*Run
}

func NewRunStore() *RunStore {
	return &RunStore{
		runs: make(map[string]*Run),
	}
}

func (s *RunStore) Create(runID, threadID string) *Run {
	s.mu.Lock()
	defer s.mu.Unlock()
	run := &Run{
		RunID:     runID,
		ThreadID:  threadID,
		Status:    "running",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	s.runs[runID] = run
	return run
}

// ============================================
// Threads API Handlers
// ============================================

// CreateThreadRequest 创建线程请求
type CreateThreadRequest struct {
	Values map[string]interface{} `json:"values,omitempty"`
}

// createThread 创建线程
func (h *LangGraphHandler) createThread(c *gin.Context) {
	var req CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空请求体
		req = CreateThreadRequest{}
	}

	threadID := uuid.NewString()
	thread := h.threadStore.Create(threadID)

	if req.Values != nil {
		thread = h.threadStore.Update(threadID, req.Values)
	}

	h.logger.Info("Thread created", zap.String("thread_id", threadID))
	c.JSON(http.StatusOK, thread)
}

// getThread 获取线程
func (h *LangGraphHandler) getThread(c *gin.Context) {
	threadID := c.Param("threadId")
	thread, ok := h.threadStore.Get(threadID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread not found"})
		return
	}
	c.JSON(http.StatusOK, thread)
}

// getThreadState 获取线程状态 (GET 方法)
func (h *LangGraphHandler) getThreadState(c *gin.Context) {
	threadID := c.Param("threadId")
	thread, ok := h.threadStore.Get(threadID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread not found"})
		return
	}

	// 返回带 checkpoint 的状态格式
	c.JSON(http.StatusOK, gin.H{
		"values":     thread.Values,
		"checkpoint": nil,
		"next":       []string{},
	})
}

// SearchThreadsRequest 搜索线程请求
type SearchThreadsRequest struct {
	Limit     *int    `json:"limit,omitempty"`
	Offset    *int    `json:"offset,omitempty"`
	SortBy    *string `json:"sortBy,omitempty"`
	SortOrder *string `json:"sortOrder,omitempty"`
}

// searchThreads 搜索线程
func (h *LangGraphHandler) searchThreads(c *gin.Context) {
	var req SearchThreadsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = SearchThreadsRequest{}
	}

	threads := h.threadStore.List()

	// 应用分页
	limit := 50
	if req.Limit != nil && *req.Limit > 0 {
		limit = *req.Limit
	}
	offset := 0
	if req.Offset != nil && *req.Offset > 0 {
		offset = *req.Offset
	}

	start := offset
	end := offset + limit
	if start > len(threads) {
		start = len(threads)
	}
	if end > len(threads) {
		end = len(threads)
	}

	pagedThreads := threads[start:end]

	h.logger.Debug("Threads searched",
		zap.Int("total", len(threads)),
		zap.Int("returned", len(pagedThreads)))

	c.JSON(http.StatusOK, pagedThreads)
}

// deleteThread 删除线程
func (h *LangGraphHandler) deleteThread(c *gin.Context) {
	threadID := c.Param("threadId")
	ok := h.threadStore.Delete(threadID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread not found"})
		return
	}
	h.logger.Info("Thread deleted", zap.String("thread_id", threadID))
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// UpdateThreadStateRequest 更新线程状态请求
type UpdateThreadStateRequest struct {
	Values map[string]interface{} `json:"values"`
}

// updateThreadState 更新线程状态
func (h *LangGraphHandler) updateThreadState(c *gin.Context) {
	threadID := c.Param("threadId")

	var req UpdateThreadStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	thread := h.threadStore.Update(threadID, req.Values)
	if thread == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ThreadHistoryState 线程历史状态项
type ThreadHistoryState struct {
	Values     map[string]interface{} `json:"values"`
	Checkpoint map[string]interface{} `json:"checkpoint,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// getThreadHistory 获取线程历史
func (h *LangGraphHandler) getThreadHistory(c *gin.Context) {
	threadID := c.Param("threadId")
	thread, ok := h.threadStore.Get(threadID)
	if !ok {
		// 返回空数组而不是 404，避免前端报错
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	// 解析 limit 参数
	var limit int = 10
	if limitStr := c.Query("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	// 返回状态历史数组格式（LangGraph SDK 期望的格式）
	history := []ThreadHistoryState{
		{
			Values: thread.Values,
		},
	}

	c.JSON(http.StatusOK, history)
}

// ============================================
// Runs API Handlers
// ============================================

// StreamRunRequest 流式运行请求
type StreamRunRequest struct {
	AssistantID string                 `json:"assistant_id"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Messages    []interface{}          `json:"messages,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// streamRun 流式运行（核心对话接口）
// 一比一复刻 DeerFlow 的 SSE 事件格式：values, messages-tuple, end
// 使用 pkg/sse/deerflow.Writer 替代手动 sendEvent
func (h *LangGraphHandler) streamRun(c *gin.Context) {
	threadID := c.Param("threadId")

	// 检查线程是否存在，如果不存在则创建
	thread, ok := h.threadStore.Get(threadID)
	if !ok {
		h.logger.Info("Thread not found, creating new one", zap.String("thread_id", threadID))
		thread = h.threadStore.Create(threadID)
	}

	var req StreamRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	runID := uuid.NewString()
	h.runStore.Create(runID, threadID)

	h.logger.Info("Stream run started (DeerFlow format - using deerflow.Writer)",
		zap.String("thread_id", threadID),
		zap.String("run_id", runID),
		zap.String("assistant_id", req.AssistantID))

	// 提取用户消息
	messages, _ := thread.Values["messages"].([]interface{})
	if req.Input != nil {
		if inputMsgs, ok := req.Input["messages"].([]interface{}); ok {
			messages = append(messages, inputMsgs...)
		}
	}
	if req.Messages != nil {
		messages = append(messages, req.Messages...)
	}

	userMessageText := extractMessageText(messages)
	if userMessageText == "" {
		userMessageText = "你好"
	}

	// 创建 DeerFlow SSE Writer
	sseWriter := deerflow.NewWriter(runID, c.Writer)
	if sseWriter == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "当前响应不支持 SSE"})
		return
	}
	defer sseWriter.Close()

	// 添加用户消息
	userMessageID := "msg-" + uuid.NewString()
	userMessage := map[string]interface{}{
		"type":    "human",
		"id":      userMessageID,
		"content": userMessageText,
	}
	messages = append(messages, userMessage)

	// 1. 发送用户消息
	sseWriter.WriteMessagesTuple(serializeMessageForTuple(userMessage))

	// 2. 调用真实 LLM 流式获取响应
	aiMessageID := "msg-" + runID
	var aiResponse strings.Builder

	h.logger.Debug("检查配置状态",
		zap.Bool("has_db", h.db != nil),
		zap.Bool("has_configLoader", h.configLoader != nil))

	if h.db != nil || h.configLoader != nil {
		// 尝试创建 ChatModelAdapter 并使用真实流式
		var chatModel *provider.ChatModelAdapter
		var err error

		if h.db != nil {
			h.logger.Debug("使用数据库创建 ChatModelAdapter")
			chatModel, err = provider.NewChatModelAdapterFromDB(h.db, h.logger, nil)
		} else if h.configLoader != nil {
			h.logger.Debug("使用 ConfigLoader 创建 ChatModelAdapter")
			chatModel, err = provider.NewChatModelAdapter(h.logger, h.configLoader, nil)
		}

		h.logger.Debug("ChatModelAdapter 创建结果",
			zap.Bool("has_chatModel", chatModel != nil),
			zap.Error(err))

		if chatModel != nil && err == nil {
			h.logger.Info("调用真实 LLM 流式...")
			einoMessages := []*schema.Message{
				schema.UserMessage(userMessageText),
			}

			// 创建流式 context
			streamCtx, streamCancel := context.WithCancel(context.Background())
			defer streamCancel()

			// 监听客户端断开
			go func() {
				<-c.Request.Context().Done()
				h.logger.Debug("客户端断开连接，取消流式调用")
				streamCancel()
			}()

			// 调用 Stream 方法
			streamReader, streamErr := chatModel.Stream(streamCtx, einoMessages)
			if streamErr != nil {
				h.logger.Error("LLM 流式调用失败", zap.Error(streamErr))
				aiResponse.WriteString(fmt.Sprintf("抱歉，LLM 流式调用失败：%v", streamErr))
			} else {
				defer streamReader.Close()

				h.logger.Info("开始读取流式响应...")

				// 参考 test/Agent-Eino 的实现：使用 buffer 累积 delta
				buffer := ""
				flushThreshold := 3 // 每 3 字符就发送，更实时
				chunkCount := 0

				for {
					// 检查客户端是否断开
					select {
					case <-c.Request.Context().Done():
						h.logger.Debug("客户端断开连接")
						return
					default:
					}

					chunk, recvErr := streamReader.Recv()
					chunkCount++

					if recvErr != nil {
						// 检查是否是流结束
						if recvErr == io.EOF || strings.Contains(recvErr.Error(), "EOF") || strings.Contains(recvErr.Error(), "end of stream") {
							h.logger.Debug("流式响应读取完毕", zap.Int("totalChunks", chunkCount))
							// 发送剩余的 buffer
							if buffer != "" {
								aiResponse.WriteString(buffer)
								partialMsg := map[string]interface{}{
									"type":    "ai",
									"id":      aiMessageID,
									"content": aiResponse.String(),
								}
								sseWriter.WriteMessagesTuple(serializeMessageForTuple(partialMsg))
							}
						} else {
							h.logger.Error("流式响应读取错误", zap.Error(recvErr), zap.Int("chunkCount", chunkCount))
						}
						break
					}

					if chunk == nil {
						continue
					}

					h.logger.Debug("[StreamReader] received chunk",
						zap.Int("chunkNum", chunkCount),
						zap.String("role", string(chunk.Role)),
						zap.Int("contentLen", len(chunk.Content)),
						zap.Int("toolCallsCount", len(chunk.ToolCalls)),
						zap.String("content", chunk.Content))

					// 处理工具调用（暂略）
					if len(chunk.ToolCalls) > 0 {
						continue
					}

					if chunk.Content == "" {
						continue
					}

					// 关键：delta 追加，不是替换！
					buffer += chunk.Content

					// 更实时地发送：每 3 字符或遇到标点就发送
					if len(buffer) >= flushThreshold || strings.Contains(buffer, "。") || strings.Contains(buffer, "\n") || strings.Contains(buffer, "！") || strings.Contains(buffer, "？") {
						aiResponse.WriteString(buffer)

						partialMsg := map[string]interface{}{
							"type":    "ai",
							"id":      aiMessageID,
							"content": aiResponse.String(),
						}
						sseWriter.WriteMessagesTuple(serializeMessageForTuple(partialMsg))

						buffer = ""
					}
				}

				h.logger.Info("LLM 流式调用完成", zap.Int("len", aiResponse.Len()))
			}
		}
	}

	// 如果没有 AI 响应，使用模拟消息
	if aiResponse.Len() == 0 {
		h.logger.Warn("使用硬编码模拟消息")
		aiResponse.WriteString(fmt.Sprintf("你好！我是 DeerFlow AI 助手。你说的是：%s", userMessageText))

		// 发送完整消息
		fullMsg := map[string]interface{}{
			"type":    "ai",
			"id":      aiMessageID,
			"content": aiResponse.String(),
		}
		sseWriter.WriteMessagesTuple(serializeMessageForTuple(fullMsg))
	}

	// 3. 更新线程状态
	aiMessage := map[string]interface{}{
		"type":    "ai",
		"id":      aiMessageID,
		"content": aiResponse.String(),
	}
	messages = append(messages, aiMessage)

	threadUpdates := map[string]interface{}{
		"messages": messages,
	}

	// 检查是否生成标题
	tempValues := make(map[string]interface{})
	for k, v := range thread.Values {
		tempValues[k] = v
	}
	tempValues["messages"] = messages

	if shouldGenerateTitle(tempValues) {
		userMsg, _ := extractFirstMessages(tempValues)
		title := generateFallbackTitle(userMsg)
		threadUpdates["title"] = title
		h.logger.Info("Generated title", zap.String("title", title))
	}

	h.threadStore.Update(threadID, threadUpdates)

	// 4. 发送 values 事件
	thread, _ = h.threadStore.Get(threadID)
	valuesEvent := buildValuesEvent(thread)
	sseWriter.WriteValues(valuesEvent)

	// 5. 发送 end 事件
	sseWriter.WriteEnd(map[string]interface{}{
		"input_tokens":  0,
		"output_tokens": 0,
		"total_tokens":  0,
	})
}

// joinStream 加入流式运行
func (h *LangGraphHandler) joinStream(c *gin.Context) {
	threadID := c.Param("threadId")
	runID := c.Param("runId")

	h.logger.Info("Join stream requested",
		zap.String("thread_id", threadID),
		zap.String("run_id", runID))

	// TODO: 实现加入已有流的逻辑
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
