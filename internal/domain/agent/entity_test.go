package agent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weibaohui/nanobot-go/internal/domain/agent"
)

func TestNewAgent(t *testing.T) {
	identity := agent.NewIdentity(
		"TestAgent",
		"A test AI agent",
		"assistant",
	)

	ag := agent.NewAgent("user-1", identity)

	assert.NotNil(t, ag)
	assert.Equal(t, "user-1", ag.UserCode())
	assert.NotEmpty(t, ag.ID().String())
	assert.Equal(t, int64(1), ag.Version())

	// 验证事件被发布
	events := ag.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*agent.AgentCreatedEvent)(nil), events[0])
}

func TestAgentAddCapability(t *testing.T) {
	identity := agent.NewIdentity("TestAgent", "Test", "assistant")
	ag := agent.NewAgent("user-1", identity)

	cap := &agent.Capability{
		Name:        "search",
		Type:        agent.ToolCapabilityType,
		Description: "Search capability",
		Config:      make(map[string]interface{}),
	}

	err := ag.AddCapability(cap)
	assert.NoError(t, err)
	assert.True(t, ag.HasCapability("search"))

	// 验证版本增加
	assert.Equal(t, int64(2), ag.Version())

	// 验证事件
	events := ag.GetUncommittedEvents()
	assert.Equal(t, 2, len(events))
	assert.IsType(t, (*agent.CapabilityAddedEvent)(nil), events[1])
}

func TestAgentAddDuplicateCapability(t *testing.T) {
	identity := agent.NewIdentity("TestAgent", "Test", "assistant")
	ag := agent.NewAgent("user-1", identity)

	cap := &agent.Capability{
		Name:        "search",
		Type:        agent.ToolCapabilityType,
		Description: "Search capability",
		Config:      make(map[string]interface{}),
	}

	ag.AddCapability(cap)
	err := ag.AddCapability(cap)

	assert.Error(t, err)
	assert.Equal(t, int64(2), ag.Version()) // 版本未增加
}

func TestAgentRemoveCapability(t *testing.T) {
	identity := agent.NewIdentity("TestAgent", "Test", "assistant")
	ag := agent.NewAgent("user-1", identity)

	cap := &agent.Capability{
		Name:        "search",
		Type:        agent.ToolCapabilityType,
		Description: "Search capability",
		Config:      make(map[string]interface{}),
	}

	ag.AddCapability(cap)
	assert.True(t, ag.HasCapability("search"))

	// 清空未提交的事件以便测试
	ag.ClearUncommittedEvents()

	err := ag.RemoveCapability("search")
	assert.NoError(t, err)
	assert.False(t, ag.HasCapability("search"))

	// 验证事件
	events := ag.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*agent.CapabilityRemovedEvent)(nil), events[0])
}

func TestAgentRemoveNonexistentCapability(t *testing.T) {
	identity := agent.NewIdentity("TestAgent", "Test", "assistant")
	ag := agent.NewAgent("user-1", identity)

	err := ag.RemoveCapability("nonexistent")
	assert.Error(t, err)
}

func TestAgentUpdatePersonality(t *testing.T) {
	identity := agent.NewIdentity("TestAgent", "Test", "assistant")
	ag := agent.NewAgent("user-1", identity)

	newConfig := agent.NewPersonalityConfig(
		"creative", "detailed", "casual", 3000, 0.9,
	)

	ag.ClearUncommittedEvents()
	err := ag.UpdatePersonality(newConfig)
	assert.NoError(t, err)

	// 验证配置已更新
	assert.Equal(t, "creative", ag.GetPersonality().ThinkingStyle)

	// 验证事件
	events := ag.GetUncommittedEvents()
	assert.Equal(t, 1, len(events))
	assert.IsType(t, (*agent.PersonalityUpdatedEvent)(nil), events[0])
}

func TestAgentUpdatePersonalityInvalidTemp(t *testing.T) {
	identity := agent.NewIdentity("TestAgent", "Test", "assistant")
	ag := agent.NewAgent("user-1", identity)

	invalidConfig := agent.NewPersonalityConfig(
		"creative", "detailed", "casual", 3000, 2.5, // 温度超出范围
	)

	err := ag.UpdatePersonality(invalidConfig)
	assert.Error(t, err)
}

func TestAgentValidate(t *testing.T) {
	identity := agent.NewIdentity("TestAgent", "Test", "assistant")
	ag := agent.NewAgent("user-1", identity)

	err := ag.Validate()
	assert.NoError(t, err)
}

func TestAgentGetCapabilities(t *testing.T) {
	identity := agent.NewIdentity("TestAgent", "Test", "assistant")
	ag := agent.NewAgent("user-1", identity)

	cap1 := &agent.Capability{
		Name:        "search",
		Type:        agent.ToolCapabilityType,
		Description: "Search capability",
		Config:      make(map[string]interface{}),
	}

	cap2 := &agent.Capability{
		Name:        "code_exec",
		Type:        agent.ToolCapabilityType,
		Description: "Code execution capability",
		Config:      make(map[string]interface{}),
	}

	ag.AddCapability(cap1)
	ag.AddCapability(cap2)

	capabilities := ag.GetCapabilities()
	assert.Equal(t, 2, len(capabilities))
}

func TestIdentityEquals(t *testing.T) {
	id1 := agent.NewIdentity("TestAgent", "Test", "assistant")
	id2 := agent.NewIdentity("TestAgent", "Test", "assistant")
	id3 := agent.NewIdentity("OtherAgent", "Other", "user")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
	assert.False(t, id1.Equals(nil))
}

func TestCapabilityEquals(t *testing.T) {
	cap1 := &agent.Capability{
		Name:        "search",
		Type:        agent.ToolCapabilityType,
		Description: "Search capability",
	}
	cap2 := &agent.Capability{
		Name:        "search",
		Type:        agent.ToolCapabilityType,
		Description: "Different description",
	}
	cap3 := &agent.Capability{
		Name:        "other",
		Type:        agent.SkillCapabilityType,
		Description: "Other capability",
	}

	assert.True(t, cap1.Equals(cap2))
	assert.False(t, cap1.Equals(cap3))
	assert.False(t, cap1.Equals(nil))
}

func TestPersonalityConfigEquals(t *testing.T) {
	pc1 := agent.NewPersonalityConfig("analytical", "concise", "professional", 2000, 0.7)
	pc2 := agent.NewPersonalityConfig("analytical", "concise", "professional", 2000, 0.7)
	pc3 := agent.NewPersonalityConfig("creative", "detailed", "casual", 3000, 0.9)

	assert.True(t, pc1.Equals(pc2))
	assert.False(t, pc1.Equals(pc3))
	assert.False(t, pc1.Equals(nil))
}
