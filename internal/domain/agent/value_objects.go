package agent

import (
	"errors"
	"fmt"
)

// 值对象：Identity - Agent 的身份标识
type Identity struct {
	name        string
	description string
	role        string
}

// NewIdentity 创建新的 Identity
func NewIdentity(name, description, role string) *Identity {
	return &Identity{
		name:        name,
		description: description,
		role:        role,
	}
}

// Name 返回名称
func (i *Identity) Name() string {
	return i.name
}

// Description 返回描述
func (i *Identity) Description() string {
	return i.description
}

// Role 返回角色
func (i *Identity) Role() string {
	return i.role
}

// Validate 验证 Identity
func (i *Identity) Validate() error {
	if i.name == "" {
		return errors.New("identity name cannot be empty")
	}
	return nil
}

// Equals 值对象相等性检查
func (i *Identity) Equals(other *Identity) bool {
	if other == nil {
		return false
	}
	return i.name == other.name &&
		i.description == other.description &&
		i.role == other.role
}

// CapabilityType 能力类型
type CapabilityType string

const (
	ToolCapabilityType   CapabilityType = "tool"
	SkillCapabilityType  CapabilityType = "skill"
	ModelCapabilityType  CapabilityType = "model"
)

// Capability 值对象：Agent 的能力
type Capability struct {
	Name        string
	Type        CapabilityType
	Description string
	Config      map[string]interface{}
}

// Validate 验证 Capability
func (c *Capability) Validate() error {
	if c.Name == "" {
		return errors.New("capability name cannot be empty")
	}
	if c.Type == "" {
		return fmt.Errorf("invalid capability type: %s", c.Type)
	}
	return nil
}

// Equals 检查两个 Capability 是否相等
func (c *Capability) Equals(other *Capability) bool {
	if other == nil {
		return false
	}
	return c.Name == other.Name && c.Type == other.Type
}

// CapabilitiesSet 集合值对象：Capability 的集合
type CapabilitiesSet struct {
	items map[string]*Capability
}

// NewCapabilitiesSet 创建新的 CapabilitiesSet
func NewCapabilitiesSet() *CapabilitiesSet {
	return &CapabilitiesSet{
		items: make(map[string]*Capability),
	}
}

// Add 添加 Capability
func (cs *CapabilitiesSet) Add(cap *Capability) {
	if cap != nil {
		cs.items[cap.Name] = cap
	}
}

// Remove 移除 Capability
func (cs *CapabilitiesSet) Remove(name string) {
	delete(cs.items, name)
}

// Contains 检查是否包含某个 Capability
func (cs *CapabilitiesSet) Contains(name string) bool {
	_, exists := cs.items[name]
	return exists
}

// All 返回所有 Capability
func (cs *CapabilitiesSet) All() []*Capability {
	result := make([]*Capability, 0, len(cs.items))
	for _, cap := range cs.items {
		result = append(result, cap)
	}
	return result
}

// PersonalityConfig 值对象：Agent 的个性配置
type PersonalityConfig struct {
	ThinkingStyle  string
	ResponseStyle  string
	Tone           string
	MaxTokens      int
	Temperature    float64
	CustomSettings map[string]interface{}
}

// NewDefaultPersonalityConfig 创建默认的 PersonalityConfig
func NewDefaultPersonalityConfig() *PersonalityConfig {
	return &PersonalityConfig{
		ThinkingStyle: "analytical",
		ResponseStyle: "concise",
		Tone:          "professional",
		MaxTokens:     2000,
		Temperature:   0.7,
		CustomSettings: make(map[string]interface{}),
	}
}

// NewPersonalityConfig 创建自定义的 PersonalityConfig
func NewPersonalityConfig(thinkingStyle, responseStyle, tone string, maxTokens int, temperature float64) *PersonalityConfig {
	return &PersonalityConfig{
		ThinkingStyle:  thinkingStyle,
		ResponseStyle:  responseStyle,
		Tone:           tone,
		MaxTokens:      maxTokens,
		Temperature:    temperature,
		CustomSettings: make(map[string]interface{}),
	}
}

// Validate 验证 PersonalityConfig
func (pc *PersonalityConfig) Validate() error {
	if pc.Temperature < 0 || pc.Temperature > 2 {
		return errors.New("temperature must be between 0 and 2")
	}
	if pc.MaxTokens <= 0 {
		return errors.New("max tokens must be positive")
	}
	return nil
}

// Equals 检查两个 PersonalityConfig 是否相等
func (pc *PersonalityConfig) Equals(other *PersonalityConfig) bool {
	if other == nil {
		return false
	}
	return pc.ThinkingStyle == other.ThinkingStyle &&
		pc.ResponseStyle == other.ResponseStyle &&
		pc.Tone == other.Tone &&
		pc.Temperature == other.Temperature &&
		pc.MaxTokens == other.MaxTokens
}
