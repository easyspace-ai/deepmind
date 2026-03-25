package session_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weibaohui/nanobot-go/internal/domain/session"
)

func TestNewConversationSession(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")

	assert.NotNil(t, sess)
	assert.Equal(t, "user-1", sess.UserCode())
	assert.Equal(t, "agent-1", sess.AgentID())
	assert.NotEmpty(t, sess.ID().String())
	assert.Equal(t, session.StateActive, sess.State())
	assert.Equal(t, int64(1), sess.Version())

	// 验证事件被发布
	events := sess.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*session.SessionCreatedEvent)(nil), events[0])
}

func TestAppendMessageID(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")
	sess.ClearUncommittedEvents()

	err := sess.AppendMessageID("msg-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), sess.Version())

	messageIDs := sess.GetMessageIDs()
	assert.Equal(t, 1, len(messageIDs))
	assert.Equal(t, "msg-1", messageIDs[0])

	// 验证事件
	events := sess.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*session.MessageAppendedEvent)(nil), events[0])
}

func TestAppendMessageIDToClosedSession(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")
	sess.Close("test close")

	err := sess.AppendMessageID("msg-1")
	assert.Error(t, err)
	assert.Equal(t, session.ErrSessionNotActive, err)
}

func TestRegisterToolCall(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")
	sess.ClearUncommittedEvents()

	toolCall := session.NewPendingToolCall("search", map[string]interface{}{
		"query": "test query",
	})

	err := sess.RegisterToolCall(toolCall)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), sess.Version())

	// 验证事件
	events := sess.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*session.ToolCallRegisteredEvent)(nil), events[0])
}

func TestResolveToolCall(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")

	toolCall := session.NewPendingToolCall("search", map[string]interface{}{
		"query": "test query",
	})

	sess.RegisterToolCall(toolCall)
	sess.ClearUncommittedEvents()

	result := map[string]interface{}{
		"results": []string{"result1", "result2"},
	}

	err := sess.ResolveToolCall(toolCall.ID, result, false)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), sess.Version())

	// 验证结果
	resolved := sess.GetPendingToolCall(toolCall.ID)
	assert.NotNil(t, resolved)
	assert.True(t, resolved.IsResolved())
	assert.False(t, resolved.Error)

	// 验证事件
	events := sess.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*session.ToolCallResolvedEvent)(nil), events[0])
}

func TestResolveNonexistentToolCall(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")

	err := sess.ResolveToolCall("nonexistent", nil, false)
	assert.Error(t, err)
}

func TestArchiveSession(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")
	sess.ClearUncommittedEvents()

	err := sess.Archive()
	assert.NoError(t, err)
	assert.Equal(t, session.StateArchived, sess.State())
	assert.Equal(t, int64(2), sess.Version())

	// 验证事件
	events := sess.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*session.SessionArchivedEvent)(nil), events[0])
}

func TestArchiveAlreadyArchivedSession(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")
	sess.Archive()

	err := sess.Archive()
	assert.Equal(t, session.ErrSessionAlreadyArchived, err)
}

func TestCloseSession(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")
	sess.ClearUncommittedEvents()

	err := sess.Close("test close")
	assert.NoError(t, err)
	assert.Equal(t, session.StateFailed, sess.State())
	assert.Equal(t, int64(2), sess.Version())

	// 验证事件
	events := sess.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*session.SessionClosedEvent)(nil), events[0])
}

func TestGetThreadData(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")
	threadData := sess.GetThreadData()

	assert.NotNil(t, threadData)
	assert.NotEmpty(t, threadData.GetWorkspacePath())
	assert.NotEmpty(t, threadData.GetUploadsPath())
	assert.NotEmpty(t, threadData.GetOutputsPath())
}

func TestGetAllPendingToolCalls(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")

	toolCall1 := session.NewPendingToolCall("search", map[string]interface{}{})
	toolCall2 := session.NewPendingToolCall("execute", map[string]interface{}{})

	sess.RegisterToolCall(toolCall1)
	sess.RegisterToolCall(toolCall2)

	calls := sess.GetAllPendingToolCalls()
	assert.Equal(t, 2, len(calls))
}

func TestValidateSession(t *testing.T) {
	sess := session.NewConversationSession("user-1", "agent-1")
	err := sess.Validate()
	assert.NoError(t, err)
}

func TestThreadDataMetadata(t *testing.T) {
	threadData := session.NewThreadData("user-1", "session-1")

	threadData.SetMetadata("key1", "value1")
	metadata := threadData.GetMetadata()

	assert.Equal(t, "value1", metadata["key1"])
}

func TestPendingToolCallResolve(t *testing.T) {
	call := session.NewPendingToolCall("search", map[string]interface{}{})

	assert.False(t, call.IsResolved())

	call.Resolve("result data", false)
	assert.True(t, call.IsResolved())
	assert.Equal(t, "result data", call.Result)
	assert.False(t, call.Error)
}
