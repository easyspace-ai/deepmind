package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ============================================
// DeerFlow 前端专用 API 处理器
// ============================================

// Model 模型信息（DeerFlow 前端格式）
type Model struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Model            string `json:"model"`
	DisplayName      string `json:"display_name"`
	SupportsThinking bool   `json:"supports_thinking"`
}

// ListModelsResponse 模型列表响应
type ListModelsResponse struct {
	Models []Model `json:"models"`
}

// listModelsDeerFlow 获取模型列表（DeerFlow 格式）
func (h *Handler) listModelsDeerFlow(c *gin.Context) {
	// TODO: 从 provider service 获取真实模型
	models := []Model{
		{
			ID:               "doubao-seed-2-0-pro",
			Name:             "doubao-seed-2-0-pro",
			Model:            "doubao-seed-2-0-pro-260215",
			DisplayName:      "Doubao Seed 2.0 Pro",
			SupportsThinking: true,
		},
		{
			ID:               "doubao-seed-1-8",
			Name:             "doubao-seed-1-8",
			Model:            "doubao-seed-1-8",
			DisplayName:      "Doubao Seed 1.8",
			SupportsThinking: true,
		},
		{
			ID:               "gpt-4o",
			Name:             "gpt-4o",
			Model:            "gpt-4o",
			DisplayName:      "GPT-4o",
			SupportsThinking: false,
		},
	}

	c.JSON(http.StatusOK, ListModelsResponse{
		Models: models,
	})
}

// ============================================
// Skills API (DeerFlow 格式)
// ============================================

// Skill 技能信息（DeerFlow 前端格式）
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	License     *string `json:"license"`
	Category    string `json:"category"`
	Enabled     bool   `json:"enabled"`
}

// ListSkillsResponseDeerFlow 技能列表响应（DeerFlow 格式）
type ListSkillsResponseDeerFlow struct {
	Skills []Skill `json:"skills"`
}

// listSkillsDeerFlow 获取技能列表（DeerFlow 格式）
func (h *Handler) listSkillsDeerFlow(c *gin.Context) {
	// TODO: 从 skill service 获取真实技能
	skills := []Skill{
		{
			Name:        "deep-research",
			Description: "Use this skill BEFORE any content generation task. Provides a systematic methodology for conducting thorough, multi-angle web research.",
			License:     nil,
			Category:    "public",
			Enabled:     true,
		},
		{
			Name:        "frontend-design",
			Description: "Create distinctive, production-grade frontend interfaces with high design quality.",
			License:     stringPtr("Complete terms in LICENSE.txt"),
			Category:    "public",
			Enabled:     true,
		},
		{
			Name:        "github-deep-research",
			Description: "Conduct multi-round deep research on any GitHub Repo.",
			License:     nil,
			Category:    "public",
			Enabled:     true,
		},
	}

	c.JSON(http.StatusOK, ListSkillsResponseDeerFlow{
		Skills: skills,
	})
}

// ============================================
// Agents API (DeerFlow 格式)
// ============================================

// Agent Agent 信息（DeerFlow 前端格式）
type AgentDeerFlow struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListAgentsResponseDeerFlow Agent 列表响应（DeerFlow 格式）
type ListAgentsResponseDeerFlow struct {
	Agents []AgentDeerFlow `json:"agents"`
}

// listAgentsDeerFlow 获取 Agent 列表（DeerFlow 格式）
func (h *Handler) listAgentsDeerFlow(c *gin.Context) {
	// TODO: 从 agent service 获取真实 Agent
	agents := []AgentDeerFlow{
		{
			Name:        "lead_agent",
			Description: "DeerFlow Lead Agent - 主智能体",
			Model:       "doubao-seed-2-0-pro",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
	}

	c.JSON(http.StatusOK, ListAgentsResponseDeerFlow{
		Agents: agents,
	})
}

// getAgentDeerFlow 获取单个 Agent（DeerFlow 格式）
func (h *Handler) getAgentDeerFlow(c *gin.Context) {
	name := c.Param("name")

	// TODO: 从 agent service 获取真实 Agent
	agent := AgentDeerFlow{
		Name:        name,
		Description: "DeerFlow Lead Agent - 主智能体",
		Model:       "doubao-seed-2-0-pro",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}

	c.JSON(http.StatusOK, agent)
}

// ============================================
// MCP Config API
// ============================================

// MCPConfig MCP 配置
type MCPConfig struct {
	Servers []map[string]interface{} `json:"servers"`
}

// getMCPConfig 获取 MCP 配置
func (h *Handler) getMCPConfig(c *gin.Context) {
	config := MCPConfig{
		Servers: []map[string]interface{}{},
	}
	c.JSON(http.StatusOK, config)
}

// updateMCPConfig 更新 MCP 配置
func (h *Handler) updateMCPConfig(c *gin.Context) {
	var config MCPConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	// TODO: 保存配置
	c.JSON(http.StatusOK, config)
}

// ============================================
// Memory API
// ============================================

// UserMemory 用户记忆
type UserMemory struct {
	Memories []map[string]interface{} `json:"memories"`
}

// getMemory 获取用户记忆
func (h *Handler) getMemory(c *gin.Context) {
	memory := UserMemory{
		Memories: []map[string]interface{}{},
	}
	c.JSON(http.StatusOK, memory)
}

// ============================================
// Helper functions
// ============================================

func stringPtr(s string) *string {
	return &s
}
