package memory

import (
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/require"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
)

func TestManager_EnqueueFromThreadState(t *testing.T) {
	m := &Manager{pending: make(map[string]string), logger: nil}
	ts := state.NewThreadState()
	ts.ThreadData = &state.ThreadDataState{WorkspacePath: "/tmp/ws"}
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
		{Role: schema.Assistant, Content: "world"},
	}
	m.EnqueueFromThreadState(ts)
	snap := m.PendingSnapshot()
	require.Contains(t, snap["/tmp/ws"], "user: hello")
	require.Contains(t, snap["/tmp/ws"], "assistant: world")
}
