package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ============================================
// LangGraph 兼容 API 处理器
// ============================================

// LangGraphHandler LangGraph 兼容 API 处理器
type LangGraphHandler struct {
	logger      *zap.Logger
	threadStore *ThreadStore
	runStore    *RunStore
}

// NewLangGraphHandler 创建 LangGraph 处理器
func NewLangGraphHandler(logger *zap.Logger) *LangGraphHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LangGraphHandler{
		logger:      logger,
		threadStore: NewThreadStore(),
		runStore:    NewRunStore(),
	}
}

// RegisterRoutes 注册 LangGraph 路由
func (h *LangGraphHandler) RegisterRoutes(router *gin.Engine) {
	lg := router.Group("/api/langgraph")
	{
		// Threads API
		lg.POST("/threads", h.createThread)
		lg.GET("/threads/:threadId", h.getThread)
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
			"messages": []interface{}{},
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

// getThreadHistory 获取线程历史
func (h *LangGraphHandler) getThreadHistory(c *gin.Context) {
	threadID := c.Param("threadId")
	thread, ok := h.threadStore.Get(threadID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread not found"})
		return
	}

	// 返回消息历史
	c.JSON(http.StatusOK, gin.H{
		"history": thread.Values["messages"],
	})
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
func (h *LangGraphHandler) streamRun(c *gin.Context) {
	threadID := c.Param("threadId")

	// 检查线程是否存在
	thread, ok := h.threadStore.Get(threadID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "thread not found"})
		return
	}

	var req StreamRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	runID := uuid.NewString()
	h.runStore.Create(runID, threadID)

	h.logger.Info("Stream run started",
		zap.String("thread_id", threadID),
		zap.String("run_id", runID),
		zap.String("assistant_id", req.AssistantID))

	// 设置响应头为 Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 获取 Gin 的 ResponseWriter
	w := c.Writer

	// 发送 metadata 事件
	sendSSE(w, "metadata", map[string]interface{}{
		"run_id":    runID,
		"thread_id": threadID,
	})

	// 发送 created 事件
	sendSSE(w, "created", map[string]interface{}{
		"run_id":    runID,
		"thread_id": threadID,
	})

	// TODO: 这里需要集成真实的 Lead Agent 执行逻辑
	// 目前返回模拟响应

	// 模拟 AI 消息
	aiMessage := map[string]interface{}{
		"type":    "ai",
		"id":      "msg-" + uuid.NewString(),
		"content": "你好！我是 DeerFlow AI 助手。目前正在实现中，稍后将提供完整功能。",
	}

	// 更新线程状态
	messages, _ := thread.Values["messages"].([]interface{})
	if req.Messages != nil {
		messages = append(messages, req.Messages...)
	}
	messages = append(messages, aiMessage)

	// 更新线程
	h.threadStore.Update(threadID, map[string]interface{}{
		"messages": messages,
	})

	// 发送 updates 事件
	sendSSE(w, "updates", map[string]interface{}{
		"__root__": map[string]interface{}{
			"messages": []interface{}{aiMessage},
		},
	})

	// 发送完整事件
	sendSSE(w, "events", map[string]interface{}{
		"event": "on_chain_start",
		"name":  "LeadAgent",
	})

	// 发送 finish 事件
	sendSSE(w, "finish", map[string]interface{}{
		"run_id": runID,
		"status": "success",
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

// ============================================
// SSE 辅助函数
// ============================================

func sendSSE(w http.ResponseWriter, event string, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", dataBytes)
	w.(http.Flusher).Flush()
}
