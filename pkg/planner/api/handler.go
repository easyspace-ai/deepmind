// Package api 提供规划系统的 API 层
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/nanobot-go/pkg/planner/model"
	"github.com/weibaohui/nanobot-go/pkg/planner/service"
	"go.uber.org/zap"
)

// PlannerHandler 规划处理器
type PlannerHandler struct {
	service *service.PlannerService
	logger  *zap.Logger
}

// NewPlannerHandler 创建规划处理器
func NewPlannerHandler(service *service.PlannerService, logger *zap.Logger) *PlannerHandler {
	if logger == nil {
		logger = zap.L()
	}
	return &PlannerHandler{
		service: service,
		logger:  logger,
	}
}

// AnalyzeIntentRequest 分析意图请求
type AnalyzeIntentRequest struct {
	Query string `json:"query" binding:"required"`
}

// AnalyzeIntentResponse 分析意图响应
type AnalyzeIntentResponse struct {
	Intent *model.IntentAnalysis `json:"intent"`
}

// HandleAnalyzeIntent 处理意图分析
func (h *PlannerHandler) HandleAnalyzeIntent(c *gin.Context) {
	var req AnalyzeIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intent, err := h.service.AnalyzeIntent(c.Request.Context(), req.Query)
	if err != nil {
		h.logger.Error("Failed to analyze intent", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AnalyzeIntentResponse{Intent: intent})
}

// DecomposeTasksRequest 分解任务请求
type DecomposeTasksRequest struct {
	Query  string                `json:"query" binding:"required"`
	Intent *model.IntentAnalysis `json:"intent" binding:"required"`
}

// DecomposeTasksResponse 分解任务响应
type DecomposeTasksResponse struct {
	Tasks []*model.SubTask `json:"tasks"`
}

// HandleDecomposeTasks 处理任务分解
func (h *PlannerHandler) HandleDecomposeTasks(c *gin.Context) {
	var req DecomposeTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tasks, err := h.service.DecomposeTasks(c.Request.Context(), req.Query, req.Intent)
	if err != nil {
		h.logger.Error("Failed to decompose tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DecomposeTasksResponse{Tasks: tasks})
}

// BuildWorkflowRequest 构建工作流请求
type BuildWorkflowRequest struct {
	Tasks []*model.SubTask `json:"tasks" binding:"required"`
}

// BuildWorkflowResponse 构建工作流响应
type BuildWorkflowResponse struct {
	Workflow *model.Workflow `json:"workflow"`
}

// HandleBuildWorkflow 处理构建工作流
func (h *PlannerHandler) HandleBuildWorkflow(c *gin.Context) {
	var req BuildWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.service.BuildWorkflow(c.Request.Context(), req.Tasks)
	if err != nil {
		h.logger.Error("Failed to build workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, BuildWorkflowResponse{Workflow: workflow})
}

// PlanRequest 完整规划请求
type PlanRequest struct {
	Query string `json:"query" binding:"required"`
}

// PlanResponse 完整规划响应
type PlanResponse struct {
	*service.PlanResult
}

// HandlePlan 处理完整规划
func (h *PlannerHandler) HandlePlan(c *gin.Context) {
	var req PlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Plan(c.Request.Context(), req.Query)
	if err != nil {
		h.logger.Error("Failed to plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, PlanResponse{PlanResult: result})
}

// RegisterRoutes 注册路由
func (h *PlannerHandler) RegisterRoutes(rg *gin.RouterGroup) {
	planner := rg.Group("/planner")
	{
		planner.POST("/analyze", h.HandleAnalyzeIntent)
		planner.POST("/decompose", h.HandleDecomposeTasks)
		planner.POST("/workflow", h.HandleBuildWorkflow)
		planner.POST("/plan", h.HandlePlan)
	}
}
