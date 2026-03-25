package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	tasksvc "github.com/weibaohui/nanobot-go/internal/service/task"
	"go.uber.org/zap"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status string `json:"status"`
}

// WebSocketHandler WebSocket 处理器接口
type WebSocketHandler interface {
	Handle(c *gin.Context)
}

// Server API 服务器
type Server struct {
	handler       *Handler
	server        *http.Server
	logger        *zap.Logger
	router        *gin.Engine
	providers     *Providers
	wsHandler     WebSocketHandler
	taskWSHandler *TaskWebSocketHandler
}

// NewServer 创建 API 服务器
func NewServer(addr string, providers *Providers, logger *zap.Logger) *Server {
	// 创建 TaskService（优先复用已注入的 TaskService，否则从 TaskManager 创建）
	var taskService TaskService
	if providers.TaskService != nil {
		taskService = providers.TaskService
	} else if providers.TaskManager != nil {
		taskService = tasksvc.NewService(providers.TaskManager)
		providers.TaskService = taskService
	}

	handler := NewHandler(
		providers.UserService,
		providers.AgentService,
		providers.ChannelService,
		providers.SessionService,
		providers.ProviderService,
		providers.CronJobService,
		providers.ConversationRecordService,
		providers.ConversationService,
		providers.SessionManager,
		providers.MCPService,
		providers.SkillService,
		taskService,
		providers.CodeLookupService,
	)

	// 创建 Gin 路由
	router := gin.Default()

	// 添加 CORS 中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:5174",
			"http://127.0.0.1:5174",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	// 注册 API 路由
	handler.RegisterRoutes(router)

	// 注册 DeerFlow 前端专用 API (无认证前缀 /api/*)
	handler.RegisterDeerFlowRoutes(router)

	// 注册 LangGraph API 路由
	if providers.LangGraphHandler != nil {
		providers.LangGraphHandler.RegisterRoutes(router)
	}

	// 注册 Planner API 路由
	apiV1 := router.Group("/api/v1")
	apiV1.Use(AuthMiddleware())
	if providers.PlannerHandler != nil {
		providers.PlannerHandler.RegisterRoutes(apiV1)
	}

	// 添加健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
	})

	server := &http.Server{
		Addr:        addr,
		Handler:     router,
		ReadTimeout: 15 * time.Second,
		// SSE uses long-lived responses, so a fixed write timeout will sever healthy streams.
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		handler:   handler,
		server:    server,
		logger:    logger,
		router:    router,
		providers: providers,
	}
}

// Start 启动 API 服务器
func (s *Server) Start() error {
	s.logger.Info("API 服务器启动", zap.String("addr", s.server.Addr))
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("API 服务器错误", zap.Error(err))
		}
	}()
	return nil
}

// Stop 停止 API 服务器
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// SetWebSocketHandler 设置 WebSocket 处理器
func (s *Server) SetWebSocketHandler(handler WebSocketHandler) {
	s.wsHandler = handler
	// 注册 WebSocket 路由
	s.router.GET("/ws/chat", handler.Handle)
}

// SetTaskWebSocketHandler 设置 Task WebSocket 处理器
func (s *Server) SetTaskWebSocketHandler(handler *TaskWebSocketHandler) {
	s.taskWSHandler = handler
	// 注册 Task WebSocket 路由
	s.router.GET("/ws/tasks", handler.Handle)
}

// GetTaskWebSocketHandler 获取 Task WebSocket 处理器
// 用于从其他模块广播消息
func (s *Server) GetTaskWebSocketHandler() *TaskWebSocketHandler {
	return s.taskWSHandler
}
