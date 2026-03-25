package agent

import "time"

// AgentCreatedEvent Agent 创建事件
type AgentCreatedEvent struct {
	AgentID   string
	UserCode  string
	Identity  *Identity
	CreatedAt time.Time
}

// EventType 返回事件类型
func (e *AgentCreatedEvent) EventType() string {
	return "agent.created"
}

// OccurredAt 返回事件发生时间
func (e *AgentCreatedEvent) OccurredAt() time.Time {
	return e.CreatedAt
}

// AggregateID 返回聚合根 ID
func (e *AgentCreatedEvent) AggregateID() string {
	return e.AgentID
}

// AggregateType 返回聚合根类型
func (e *AgentCreatedEvent) AggregateType() string {
	return "Agent"
}

// CapabilityAddedEvent 能力添加事件
type CapabilityAddedEvent struct {
	AgentID    string
	Capability *Capability
	AddedAt    time.Time
}

// EventType 返回事件类型
func (e *CapabilityAddedEvent) EventType() string {
	return "agent.capability_added"
}

// OccurredAt 返回事件发生时间
func (e *CapabilityAddedEvent) OccurredAt() time.Time {
	return e.AddedAt
}

// AggregateID 返回聚合根 ID
func (e *CapabilityAddedEvent) AggregateID() string {
	return e.AgentID
}

// AggregateType 返回聚合根类型
func (e *CapabilityAddedEvent) AggregateType() string {
	return "Agent"
}

// CapabilityRemovedEvent 能力移除事件
type CapabilityRemovedEvent struct {
	AgentID    string
	Capability string
	RemovedAt  time.Time
}

// EventType 返回事件类型
func (e *CapabilityRemovedEvent) EventType() string {
	return "agent.capability_removed"
}

// OccurredAt 返回事件发生时间
func (e *CapabilityRemovedEvent) OccurredAt() time.Time {
	return e.RemovedAt
}

// AggregateID 返回聚合根 ID
func (e *CapabilityRemovedEvent) AggregateID() string {
	return e.AgentID
}

// AggregateType 返回聚合根类型
func (e *CapabilityRemovedEvent) AggregateType() string {
	return "Agent"
}

// PersonalityUpdatedEvent 个性配置更新事件
type PersonalityUpdatedEvent struct {
	AgentID   string
	OldConfig *PersonalityConfig
	NewConfig *PersonalityConfig
	UpdatedAt time.Time
}

// EventType 返回事件类型
func (e *PersonalityUpdatedEvent) EventType() string {
	return "agent.personality_updated"
}

// OccurredAt 返回事件发生时间
func (e *PersonalityUpdatedEvent) OccurredAt() time.Time {
	return e.UpdatedAt
}

// AggregateID 返回聚合根 ID
func (e *PersonalityUpdatedEvent) AggregateID() string {
	return e.AgentID
}

// AggregateType 返回聚合根类型
func (e *PersonalityUpdatedEvent) AggregateType() string {
	return "Agent"
}
