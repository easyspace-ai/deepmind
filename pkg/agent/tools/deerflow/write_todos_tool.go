package deerflow

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
)

// WriteTodosTool write_todos：用模型提供的列表覆盖线程 Todos（与 TodoListMiddleware 配合）。
type WriteTodosTool struct {
	*BaseDeerFlowTool
	cfg *ToolConfig
}

// NewWriteTodosTool 创建 write_todos 工具。
func NewWriteTodosTool(cfg *ToolConfig) tool.BaseTool {
	return &WriteTodosTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"write_todos",
			"Replace the thread todo list. Pass the full list each time; statuses: pending, in_progress, completed.",
			map[string]interface{}{
				"todos": map[string]interface{}{
					"type":        "array",
					"description": "Complete todo list for this thread",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type":        "string",
								"description": "Stable id",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "Task text",
							},
							"status": map[string]interface{}{
								"type":        "string",
								"description": "pending | in_progress | completed",
								"enum":        []string{"pending", "in_progress", "completed"},
							},
						},
					},
				},
			},
		),
		cfg: cfg,
	}
}

// Invoke 解析 todos 并写回 ToolConfig.ThreadState（若已绑定）。
func (t *WriteTodosTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	raw, _ := args["todos"].([]interface{})
	now := time.Now()
	var items []state.TodoItem
	for i, it := range raw {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := m["id"].(string)
		if id == "" {
			id = fmt.Sprintf("todo-%d", i)
		}
		desc, _ := m["description"].(string)
		st, _ := m["status"].(string)
		if st == "" {
			st = "pending"
		}
		items = append(items, state.TodoItem{
			ID:          id,
			Description: desc,
			Status:      st,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	if t.cfg != nil && t.cfg.ThreadState != nil {
		t.cfg.ThreadState.Todos = items
	}
	return map[string]interface{}{
		"todos":   items,
		"content": fmt.Sprintf("Updated %d todo(s)", len(items)),
	}, nil
}
