package agent

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AgentID 类型
type AgentID string

// String 返回字符串表示
func (id AgentID) String() string {
	return string(id)
}

// NewAgentID 生成新的 AgentID
func NewAgentID() AgentID {
	return AgentID(uuid.New().String())
}

// Agent 聚合根：管理 AI Agent 的身份、能力、配置
type Agent struct {
	// 标识
	id       AgentID
	userCode string // 多租户隔离键

	// 核心属性
	identity     *Identity
	capabilities *CapabilitiesSet
	personality  *PersonalityConfig

	// 元数据
	version   int64
	createdAt time.Time
	updatedAt time.Time

	// 事件
	uncommittedEvents []interface{}
}

// NewAgent 创建新的 Agent
func NewAgent(userCode string, identity *Identity) *Agent {
	agent := &Agent{
		id:                NewAgentID(),
		userCode:          userCode,
		version:           1,
		identity:          identity,
		capabilities:      NewCapabilitiesSet(),
		personality:       NewDefaultPersonalityConfig(),
		createdAt:         time.Now(),
		updatedAt:         time.Now(),
		uncommittedEvents: make([]interface{}, 0),
	}

	agent.RaiseDomainEvent(&AgentCreatedEvent{
		AgentID:   agent.id.String(),
		UserCode:  userCode,
		Identity:  identity,
		CreatedAt: agent.createdAt,
	})

	return agent
}

// ID 返回 Agent ID
func (a *Agent) ID() AgentID {
	return a.id
}

// UserCode 返回用户代码
func (a *Agent) UserCode() string {
	return a.userCode
}

// Version 返回版本（用于乐观锁）
func (a *Agent) Version() int64 {
	return a.version
}

// GetIdentity 返回 Agent 身份
func (a *Agent) GetIdentity() *Identity {
	return a.identity
}

// GetPersonality 返回 Agent 个性配置
func (a *Agent) GetPersonality() *PersonalityConfig {
	return a.personality
}

// AddCapability 业务规则：添加能力
func (a *Agent) AddCapability(cap *Capability) error {
	if cap == nil {
		return ErrInvalidCapability
	}

	if err := cap.Validate(); err != nil {
		return fmt.Errorf("invalid capability: %w", err)
	}

	if a.capabilities.Contains(cap.Name) {
		return fmt.Errorf("capability %s already exists", cap.Name)
	}

	a.capabilities.Add(cap)
	a.version++
	a.updatedAt = time.Now()

	a.RaiseDomainEvent(&CapabilityAddedEvent{
		AgentID:    a.id.String(),
		Capability: cap,
		AddedAt:    a.updatedAt,
	})

	return nil
}

// RemoveCapability 业务规则：移除能力
func (a *Agent) RemoveCapability(capName string) error {
	if capName == "" {
		return errors.New("capability name cannot be empty")
	}

	if !a.capabilities.Contains(capName) {
		return fmt.Errorf("capability %s not found", capName)
	}

	a.capabilities.Remove(capName)
	a.version++
	a.updatedAt = time.Now()

	a.RaiseDomainEvent(&CapabilityRemovedEvent{
		AgentID:    a.id.String(),
		Capability: capName,
		RemovedAt:  a.updatedAt,
	})

	return nil
}

// UpdatePersonality 业务规则：更新个性配置
func (a *Agent) UpdatePersonality(config *PersonalityConfig) error {
	if config == nil {
		return ErrInvalidConfig
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid personality config: %w", err)
	}

	if a.personality.Equals(config) {
		// 无需更新
		return nil
	}

	oldConfig := a.personality
	a.personality = config
	a.version++
	a.updatedAt = time.Now()

	a.RaiseDomainEvent(&PersonalityUpdatedEvent{
		AgentID:   a.id.String(),
		OldConfig: oldConfig,
		NewConfig: config,
		UpdatedAt: a.updatedAt,
	})

	return nil
}

// GetCapabilities 查询方法：获取所有能力
func (a *Agent) GetCapabilities() []*Capability {
	return a.capabilities.All()
}

// HasCapability 查询方法：检查是否有某个能力
func (a *Agent) HasCapability(name string) bool {
	return a.capabilities.Contains(name)
}

// RaiseDomainEvent 发布领域事件
func (a *Agent) RaiseDomainEvent(event interface{}) {
	a.uncommittedEvents = append(a.uncommittedEvents, event)
}

// GetUncommittedEvents 获取未提交的事件
func (a *Agent) GetUncommittedEvents() []interface{} {
	return a.uncommittedEvents
}

// ClearUncommittedEvents 清空未提交的事件
func (a *Agent) ClearUncommittedEvents() {
	a.uncommittedEvents = make([]interface{}, 0)
}

// Validate 验证聚合根
func (a *Agent) Validate() error {
	if a.id == "" {
		return errors.New("agent id cannot be empty")
	}
	if a.userCode == "" {
		return errors.New("user code cannot be empty")
	}
	if a.identity == nil {
		return errors.New("agent identity cannot be nil")
	}
	if err := a.identity.Validate(); err != nil {
		return err
	}
	if err := a.personality.Validate(); err != nil {
		return err
	}
	return nil
}

// 错误定义
var (
	ErrInvalidCapability = errors.New("invalid capability")
	ErrInvalidConfig     = errors.New("invalid config")
)
