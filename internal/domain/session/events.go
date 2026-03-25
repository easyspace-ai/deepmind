package session

import "time"

// SessionCreatedEvent 会话创建事件
type SessionCreatedEvent struct {
	SessionID string
	UserCode  string
	AgentID   string
	CreatedAt time.Time
}

// EventType 返回事件类型
func (e *SessionCreatedEvent) EventType() string {
	return "session.created"
}

// OccurredAt 返回事件发生时间
func (e *SessionCreatedEvent) OccurredAt() time.Time {
	return e.CreatedAt
}

// AggregateID 返回聚合根 ID
func (e *SessionCreatedEvent) AggregateID() string {
	return e.SessionID
}

// AggregateType 返回聚合根类型
func (e *SessionCreatedEvent) AggregateType() string {
	return "ConversationSession"
}

// MessageAppendedEvent 消息追加事件
type MessageAppendedEvent struct {
	SessionID  string
	MessageID  string
	AppendedAt time.Time
}

// EventType 返回事件类型
func (e *MessageAppendedEvent) EventType() string {
	return "session.message_appended"
}

// OccurredAt 返回事件发生时间
func (e *MessageAppendedEvent) OccurredAt() time.Time {
	return e.AppendedAt
}

// AggregateID 返回聚合根 ID
func (e *MessageAppendedEvent) AggregateID() string {
	return e.SessionID
}

// AggregateType 返回聚合根类型
func (e *MessageAppendedEvent) AggregateType() string {
	return "ConversationSession"
}

// ToolCallRegisteredEvent 工具调用注册事件
type ToolCallRegisteredEvent struct {
	SessionID    string
	ToolCallID   string
	ToolName     string
	RegisteredAt time.Time
}

// EventType 返回事件类型
func (e *ToolCallRegisteredEvent) EventType() string {
	return "session.tool_call_registered"
}

// OccurredAt 返回事件发生时间
func (e *ToolCallRegisteredEvent) OccurredAt() time.Time {
	return e.RegisteredAt
}

// AggregateID 返回聚合根 ID
func (e *ToolCallRegisteredEvent) AggregateID() string {
	return e.SessionID
}

// AggregateType 返回聚合根类型
func (e *ToolCallRegisteredEvent) AggregateType() string {
	return "ConversationSession"
}

// ToolCallResolvedEvent 工具调用解决事件
type ToolCallResolvedEvent struct {
	SessionID   string
	ToolCallID  string
	Result      interface{}
	IsError     bool
	ResolvedAt  time.Time
}

// EventType 返回事件类型
func (e *ToolCallResolvedEvent) EventType() string {
	return "session.tool_call_resolved"
}

// OccurredAt 返回事件发生时间
func (e *ToolCallResolvedEvent) OccurredAt() time.Time {
	return e.ResolvedAt
}

// AggregateID 返回聚合根 ID
func (e *ToolCallResolvedEvent) AggregateID() string {
	return e.SessionID
}

// AggregateType 返回聚合根类型
func (e *ToolCallResolvedEvent) AggregateType() string {
	return "ConversationSession"
}

// SessionArchivedEvent 会话存档事件
type SessionArchivedEvent struct {
	SessionID  string
	ArchivedAt time.Time
}

// EventType 返回事件类型
func (e *SessionArchivedEvent) EventType() string {
	return "session.archived"
}

// OccurredAt 返回事件发生时间
func (e *SessionArchivedEvent) OccurredAt() time.Time {
	return e.ArchivedAt
}

// AggregateID 返回聚合根 ID
func (e *SessionArchivedEvent) AggregateID() string {
	return e.SessionID
}

// AggregateType 返回聚合根类型
func (e *SessionArchivedEvent) AggregateType() string {
	return "ConversationSession"
}

// SessionClosedEvent 会话关闭事件
type SessionClosedEvent struct {
	SessionID string
	Reason    string
	ClosedAt  time.Time
}

// EventType 返回事件类型
func (e *SessionClosedEvent) EventType() string {
	return "session.closed"
}

// OccurredAt 返回事件发生时间
func (e *SessionClosedEvent) OccurredAt() time.Time {
	return e.ClosedAt
}

// AggregateID 返回聚合根 ID
func (e *SessionClosedEvent) AggregateID() string {
	return e.SessionID
}

// AggregateType 返回聚合根类型
func (e *SessionClosedEvent) AggregateType() string {
	return "ConversationSession"
}
