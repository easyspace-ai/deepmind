package session

import "context"

// ConversationSessionRepository 会话仓储接口
type ConversationSessionRepository interface {
	// FindByID 根据 ID 查找 Session
	FindByID(ctx context.Context, id SessionID) (*ConversationSession, error)

	// Save 保存 Session
	Save(ctx context.Context, session *ConversationSession) error

	// Update 更新 Session
	Update(ctx context.Context, session *ConversationSession) error

	// Delete 删除 Session
	Delete(ctx context.Context, id SessionID) error

	// FindByUserCode 根据用户代码查找所有 Session
	FindByUserCode(ctx context.Context, userCode string) ([]*ConversationSession, error)

	// FindActiveByUserCode 根据用户代码查找所有活跃 Session
	FindActiveByUserCode(ctx context.Context, userCode string) ([]*ConversationSession, error)

	// FindAll 查找所有 Session
	FindAll(ctx context.Context) ([]*ConversationSession, error)
}
