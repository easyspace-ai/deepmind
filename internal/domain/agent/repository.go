package agent

import "context"

// AgentRepository Agent 仓储接口
type AgentRepository interface {
	// FindByID 根据 ID 查找 Agent
	FindByID(ctx context.Context, id AgentID) (*Agent, error)

	// Save 保存 Agent
	Save(ctx context.Context, agent *Agent) error

	// Update 更新 Agent
	Update(ctx context.Context, agent *Agent) error

	// Delete 删除 Agent
	Delete(ctx context.Context, id AgentID) error

	// FindByUserCode 根据用户代码查找所有 Agent
	FindByUserCode(ctx context.Context, userCode string) ([]*Agent, error)

	// FindAll 查找所有 Agent
	FindAll(ctx context.Context) ([]*Agent, error)
}
