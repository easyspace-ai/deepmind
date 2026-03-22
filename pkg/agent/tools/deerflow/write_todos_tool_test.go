package deerflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
)

func TestWriteTodosTool_Invoke(t *testing.T) {
	ts := state.NewThreadState()
	cfg := &ToolConfig{ThreadState: ts}
	tool := NewWriteTodosTool(cfg).(*WriteTodosTool)
	out, err := tool.Invoke(context.Background(), map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{"id": "a", "description": "one", "status": "pending"},
		},
	})
	require.NoError(t, err)
	m, ok := out.(map[string]interface{})
	require.True(t, ok)
	require.Len(t, ts.Todos, 1)
	require.Equal(t, "a", ts.Todos[0].ID)
	require.Equal(t, "one", ts.Todos[0].Description)
	require.NotNil(t, m["todos"])
}
