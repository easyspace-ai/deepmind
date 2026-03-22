package state

import (
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/require"
)

func TestApplyMiddlewareUpdates(t *testing.T) {
	s := NewThreadState()
	s.Messages = []*schema.Message{schema.UserMessage("a")}
	u := map[string]interface{}{
		"messages": []*schema.Message{schema.UserMessage("b")},
		"title":    "hello",
		"todos": []TodoItem{
			{ID: "1", Description: "x", Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		"artifacts": []string{"out/a.txt"},
		"viewed_images": map[string]ViewedImageData{
			"/p": {Base64: "QQ==", MimeType: "image/png"},
		},
	}
	ApplyMiddlewareUpdates(s, u)
	require.Len(t, s.Messages, 1)
	require.Equal(t, "b", s.Messages[0].Content)
	require.Equal(t, "hello", s.Title)
	require.Len(t, s.Todos, 1)
	require.Contains(t, s.Artifacts, "out/a.txt")
	require.Contains(t, s.ViewedImages, "/p")
}

func TestApplyMiddlewareUpdates_ClarificationAndToolError(t *testing.T) {
	s := NewThreadState()
	ApplyMiddlewareUpdates(s, map[string]interface{}{
		"interrupt":               true,
		"clarification_message":   "请选择方案",
		"tool_error":              map[string]interface{}{"tool_name": "bash", "tool_call_id": "c1", "error": "oops"},
	})
	require.True(t, s.ClarificationPending)
	require.Equal(t, "请选择方案", s.LastClarificationMessage)
	require.NotNil(t, s.LastToolError)
	require.Equal(t, "bash", s.LastToolError.ToolName)
	require.Equal(t, "c1", s.LastToolError.ToolCallID)
	require.Equal(t, "oops", s.LastToolError.Error)
}
