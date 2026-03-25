package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SessionID 类型
type SessionID string

// String 返回字符串表示
func (id SessionID) String() string {
	return string(id)
}

// NewSessionID 生成新的 SessionID
func NewSessionID() SessionID {
	return SessionID(uuid.New().String())
}

// SessionState 会话状态类型
type SessionState string

const (
	StateActive   SessionState = "active"
	StateArchived SessionState = "archived"
	StateFailed   SessionState = "failed"
)

// IsValid 检查会话状态是否有效
func (s SessionState) IsValid() bool {
	return s == StateActive || s == StateArchived || s == StateFailed
}

// ConversationSession 聚合根：管理会话生命周期
type ConversationSession struct {
	// 标识
	id       SessionID
	userCode string
	agentID  string

	// 状态
	state SessionState

	// 消息历史（仅存储 ID）
	messageIDs []string

	// 线程数据（虚拟文件系统状态）
	threadData *ThreadData

	// 工具调用状态
	pendingToolCalls map[string]*PendingToolCall

	// 元数据
	version   int64
	createdAt time.Time
	updatedAt time.Time

	// 事件
	uncommittedEvents []interface{}
}

// NewConversationSession 创建新的 ConversationSession
func NewConversationSession(userCode, agentID string) *ConversationSession {
	sessionID := NewSessionID()
	session := &ConversationSession{
		id:               sessionID,
		userCode:         userCode,
		agentID:          agentID,
		version:          1,
		state:            StateActive,
		messageIDs:       make([]string, 0),
		threadData:       NewThreadData(userCode, sessionID.String()),
		pendingToolCalls: make(map[string]*PendingToolCall),
		createdAt:        time.Now(),
		updatedAt:        time.Now(),
		uncommittedEvents: make([]interface{}, 0),
	}

	session.RaiseDomainEvent(&SessionCreatedEvent{
		SessionID: session.id.String(),
		UserCode:  userCode,
		AgentID:   agentID,
		CreatedAt: session.createdAt,
	})

	return session
}

// ID 返回会话 ID
func (s *ConversationSession) ID() SessionID {
	return s.id
}

// UserCode 返回用户代码
func (s *ConversationSession) UserCode() string {
	return s.userCode
}

// AgentID 返回 Agent ID
func (s *ConversationSession) AgentID() string {
	return s.agentID
}

// State 返回会话状态
func (s *ConversationSession) State() SessionState {
	return s.state
}

// Version 返回版本
func (s *ConversationSession) Version() int64 {
	return s.version
}

// GetThreadData 返回线程数据
func (s *ConversationSession) GetThreadData() *ThreadData {
	return s.threadData
}

// GetMessageIDs 返回所有消息 ID
func (s *ConversationSession) GetMessageIDs() []string {
	return s.messageIDs
}

// AppendMessageID 业务规则：追加消息 ID
func (s *ConversationSession) AppendMessageID(msgID string) error {
	if s.state != StateActive {
		return ErrSessionNotActive
	}

	if msgID == "" {
		return ErrInvalidMessageID
	}

	s.messageIDs = append(s.messageIDs, msgID)
	s.version++
	s.updatedAt = time.Now()

	s.RaiseDomainEvent(&MessageAppendedEvent{
		SessionID:  s.id.String(),
		MessageID:  msgID,
		AppendedAt: s.updatedAt,
	})

	return nil
}

// RegisterToolCall 业务规则：注册工具调用
func (s *ConversationSession) RegisterToolCall(call *PendingToolCall) error {
	if s.state != StateActive {
		return ErrSessionNotActive
	}

	if call == nil || call.ID == "" || call.Name == "" {
		return ErrInvalidToolCall
	}

	s.pendingToolCalls[call.ID] = call
	s.version++
	s.updatedAt = time.Now()

	s.RaiseDomainEvent(&ToolCallRegisteredEvent{
		SessionID:    s.id.String(),
		ToolCallID:   call.ID,
		ToolName:     call.Name,
		RegisteredAt: s.updatedAt,
	})

	return nil
}

// ResolveToolCall 业务规则：解决工具调用
func (s *ConversationSession) ResolveToolCall(toolCallID string, result interface{}, isError bool) error {
	if s.state != StateActive {
		return ErrSessionNotActive
	}

	call, ok := s.pendingToolCalls[toolCallID]
	if !ok {
		return fmt.Errorf("tool call %s not found", toolCallID)
	}

	call.Resolve(result, isError)
	s.version++
	s.updatedAt = time.Now()

	s.RaiseDomainEvent(&ToolCallResolvedEvent{
		SessionID:   s.id.String(),
		ToolCallID:  toolCallID,
		Result:      result,
		IsError:     isError,
		ResolvedAt:  call.ResolvedAt,
	})

	return nil
}

// Archive 业务规则：存档会话
func (s *ConversationSession) Archive() error {
	if s.state == StateArchived {
		return ErrSessionAlreadyArchived
	}

	s.state = StateArchived
	s.version++
	s.updatedAt = time.Now()

	s.RaiseDomainEvent(&SessionArchivedEvent{
		SessionID:  s.id.String(),
		ArchivedAt: s.updatedAt,
	})

	return nil
}

// Close 业务规则：关闭会话
func (s *ConversationSession) Close(reason string) error {
	if s.state != StateActive {
		return ErrSessionNotActive
	}

	s.state = StateFailed
	s.version++
	s.updatedAt = time.Now()

	s.RaiseDomainEvent(&SessionClosedEvent{
		SessionID: s.id.String(),
		Reason:    reason,
		ClosedAt:  s.updatedAt,
	})

	return nil
}

// GetPendingToolCall 查询方法：获取单个工具调用
func (s *ConversationSession) GetPendingToolCall(id string) *PendingToolCall {
	return s.pendingToolCalls[id]
}

// GetAllPendingToolCalls 查询方法：获取所有工具调用
func (s *ConversationSession) GetAllPendingToolCalls() []*PendingToolCall {
	result := make([]*PendingToolCall, 0, len(s.pendingToolCalls))
	for _, call := range s.pendingToolCalls {
		result = append(result, call)
	}
	return result
}

// RaiseDomainEvent 发布领域事件
func (s *ConversationSession) RaiseDomainEvent(event interface{}) {
	s.uncommittedEvents = append(s.uncommittedEvents, event)
}

// GetUncommittedEvents 获取未提交的事件
func (s *ConversationSession) GetUncommittedEvents() []interface{} {
	return s.uncommittedEvents
}

// ClearUncommittedEvents 清空未提交的事件
func (s *ConversationSession) ClearUncommittedEvents() {
	s.uncommittedEvents = make([]interface{}, 0)
}

// Validate 验证聚合根
func (s *ConversationSession) Validate() error {
	if s.id == "" {
		return errors.New("session id cannot be empty")
	}
	if s.userCode == "" {
		return errors.New("user code cannot be empty")
	}
	if s.agentID == "" {
		return errors.New("agent id cannot be empty")
	}
	if !s.state.IsValid() {
		return fmt.Errorf("invalid session state: %s", s.state)
	}
	return nil
}

// 错误定义
var (
	ErrSessionNotActive      = errors.New("session is not active")
	ErrInvalidMessageID      = errors.New("invalid message id")
	ErrInvalidToolCall       = errors.New("invalid tool call")
	ErrSessionAlreadyArchived = errors.New("session already archived")
)
