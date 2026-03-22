package prompts

import (
	"strings"
	"testing"
	"time"
)

func TestNewPrompt(t *testing.T) {
	p := NewPrompt()
	if p == nil {
		t.Error("NewPrompt() should not return nil")
	}
	if p.sections == nil {
		t.Error("sections should not be nil")
	}
	if len(p.sections) != 0 {
		t.Error("sections should be empty")
	}
}

func TestPrompt_AddSection(t *testing.T) {
	p := NewPrompt()
	section := NewBaseSection("test", "test content")
	p.AddSection(section)

	if len(p.sections) != 1 {
		t.Errorf("AddSection() len = %v, want 1", len(p.sections))
	}
}

func TestPrompt_AddSections(t *testing.T) {
	p := NewPrompt()
	section1 := NewBaseSection("test1", "content1")
	section2 := NewBaseSection("test2", "content2")
	p.AddSections(section1, section2)

	if len(p.sections) != 2 {
		t.Errorf("AddSections() len = %v, want 2", len(p.sections))
	}
}

func TestPrompt_GetSection(t *testing.T) {
	p := NewPrompt()
	section := NewBaseSection("test", "test content")
	p.AddSection(section)

	found := p.GetSection("test")
	if found == nil {
		t.Error("GetSection() should find section")
	}
	if found.Name() != "test" {
		t.Errorf("GetSection() name = %v, want 'test'", found.Name())
	}

	notFound := p.GetSection("not_found")
	if notFound != nil {
		t.Error("GetSection() should return nil for not found")
	}
}

func TestPrompt_RemoveSection(t *testing.T) {
	p := NewPrompt()
	p.AddSection(NewBaseSection("test1", "content1"))
	p.AddSection(NewBaseSection("test2", "content2"))
	p.RemoveSection("test1")

	if len(p.sections) != 1 {
		t.Errorf("RemoveSection() len = %v, want 1", len(p.sections))
	}
	if p.GetSection("test1") != nil {
		t.Error("RemoveSection() should remove section")
	}
}

func TestPrompt_Render(t *testing.T) {
	p := NewPrompt()
	p.AddSection(NewBaseSection("test1", "content1"))
	p.AddSection(NewBaseSection("test2", "content2"))

	result := p.Render()
	if !strings.Contains(result, "content1") {
		t.Error("Render() should contain content1")
	}
	if !strings.Contains(result, "content2") {
		t.Error("Render() should contain content2")
	}
}

func TestBaseSection(t *testing.T) {
	section := NewBaseSection("test", "test content")
	if section.Name() != "test" {
		t.Errorf("Name() = %v, want 'test'", section.Name())
	}
	if section.Render() != "test content" {
		t.Errorf("Render() = %v, want 'test content'", section.Render())
	}
}

func TestNamedSection(t *testing.T) {
	called := false
	section := NewNamedSection("test", func() string {
		called = true
		return "dynamic content"
	})

	if section.Name() != "test" {
		t.Errorf("Name() = %v, want 'test'", section.Name())
	}
	if section.Render() != "dynamic content" {
		t.Errorf("Render() = %v, want 'dynamic content'", section.Render())
	}
	if !called {
		t.Error("renderFn should be called")
	}
}

func TestRoleSection(t *testing.T) {
	section := NewRoleSection("TestAgent")
	if section.Name() != "role" {
		t.Errorf("Name() = %v, want 'role'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "TestAgent") {
		t.Error("Render() should contain agent name")
	}
}

func TestThinkingStyleSection(t *testing.T) {
	section := NewThinkingStyleSection(false, 3)
	if section.Name() != "thinking_style" {
		t.Errorf("Name() = %v, want 'thinking_style'", section.Name())
	}
	rendered := section.Render()
	if rendered == "" {
		t.Error("Render() should not be empty")
	}

	// With subagent
	sectionWithSubagent := NewThinkingStyleSection(true, 5)
	renderedWithSubagent := sectionWithSubagent.Render()
	if !strings.Contains(renderedWithSubagent, "DECOMPOSITION CHECK") {
		t.Error("Render() with subagent should contain DECOMPOSITION CHECK")
	}
}

func TestClarificationSection(t *testing.T) {
	section := NewClarificationSection()
	if section.Name() != "clarification" {
		t.Errorf("Name() = %v, want 'clarification'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "clarification_system") {
		t.Error("Render() should contain clarification_system")
	}
}

func TestSubagentSection(t *testing.T) {
	section := NewSubagentSection(3)
	if section.Name() != "subagent" {
		t.Errorf("Name() = %v, want 'subagent'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "subagent_system") {
		t.Error("Render() should contain subagent_system")
	}
	if !strings.Contains(rendered, "MAXIMUM 3") {
		t.Error("Render() should contain MAXIMUM 3")
	}
}

func TestWorkingDirSection(t *testing.T) {
	section := NewWorkingDirSection()
	if section.Name() != "working_dir" {
		t.Errorf("Name() = %v, want 'working_dir'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "working_directory") {
		t.Error("Render() should contain working_directory")
	}
}

func TestSkillsSection(t *testing.T) {
	section := NewSkillsSection("test skills content")
	if section.Name() != "skills" {
		t.Errorf("Name() = %v, want 'skills'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "test skills content") {
		t.Error("Render() should contain skills content")
	}

	// Empty content
	emptySection := NewSkillsSection("")
	if emptySection.Render() != "" {
		t.Error("Render() with empty content should return empty string")
	}
}

func TestEnvInfoSection(t *testing.T) {
	customTime := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	section := NewEnvInfoSectionWithTime(customTime)
	if section.Name() != "env_info" {
		t.Errorf("Name() = %v, want 'env_info'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "2026-03-22") {
		t.Error("Render() should contain custom date")
	}
}

func TestCustomSection(t *testing.T) {
	section := NewCustomSection("custom", "custom_tag", "custom content")
	if section.Name() != "custom" {
		t.Errorf("Name() = %v, want 'custom'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "<custom_tag>") {
		t.Error("Render() should contain custom_tag")
	}
	if !strings.Contains(rendered, "custom content") {
		t.Error("Render() should contain custom content")
	}

	// Empty tag
	noTagSection := NewCustomSection("custom", "", "content")
	if noTagSection.Render() != "content" {
		t.Error("Render() with empty tag should return just content")
	}

	// Empty content
	emptyContentSection := NewCustomSection("custom", "tag", "")
	if emptyContentSection.Render() != "" {
		t.Error("Render() with empty content should return empty string")
	}
}

func TestResponseStyleSection(t *testing.T) {
	section := NewResponseStyleSection()
	if section.Name() != "response_style" {
		t.Errorf("Name() = %v, want 'response_style'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "response_style") {
		t.Error("Render() should contain response_style")
	}
}

func TestCitationsSection(t *testing.T) {
	section := NewCitationsSection()
	if section.Name() != "citations" {
		t.Errorf("Name() = %v, want 'citations'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "citations") {
		t.Error("Render() should contain citations")
	}
}

func TestCriticalRemindersSection(t *testing.T) {
	section := NewCriticalRemindersSection(false, 3)
	if section.Name() != "critical_reminders" {
		t.Errorf("Name() = %v, want 'critical_reminders'", section.Name())
	}
	rendered := section.Render()
	if !strings.Contains(rendered, "critical_reminders") {
		t.Error("Render() should contain critical_reminders")
	}

	// With subagent
	withSubagent := NewCriticalRemindersSection(true, 5)
	renderedWithSubagent := withSubagent.Render()
	if !strings.Contains(renderedWithSubagent, "Orchestrator Mode") {
		t.Error("Render() with subagent should contain Orchestrator Mode")
	}
}

func TestBuilder(t *testing.T) {
	builder := NewBuilder()
	prompt := builder.
		WithRole("TestAgent").
		WithThinkingStyle().
		WithClarification().
		WithWorkingDir().
		WithEnvInfo().
		Build()

	if len(prompt.Sections()) != 5 {
		t.Errorf("Builder sections len = %v, want 5", len(prompt.Sections()))
	}

	rendered := prompt.Render()
	if rendered == "" {
		t.Error("Build() should render non-empty string")
	}
}

func TestBuilder_WithSubagent(t *testing.T) {
	builder := NewBuilder()
	prompt := builder.
		WithRole("TestAgent").
		WithThinkingStyleWithSubagent(5).
		WithSubagent(5).
		WithCriticalRemindersWithSubagent(5).
		Build()

	if len(prompt.Sections()) != 4 {
		t.Errorf("Builder sections len = %v, want 4", len(prompt.Sections()))
	}
}

func TestBuildLeadAgentPrompt(t *testing.T) {
	config := &LeadAgentConfig{
		AgentName:              "TestAgent",
		SoulContent:            "Test soul content",
		MemoryContent:          "Test memory content",
		SkillsContent:          "Test skills content",
		SubagentEnabled:        true,
		MaxConcurrentSubagents: 5,
	}

	prompt := BuildLeadAgentPrompt(config)
	if prompt == nil {
		t.Error("BuildLeadAgentPrompt() should not return nil")
	}

	rendered := prompt.Render()
	if !strings.Contains(rendered, "TestAgent") {
		t.Error("Render() should contain TestAgent")
	}
	if !strings.Contains(rendered, "Test soul content") {
		t.Error("Render() should contain soul content")
	}
}

func TestBuildLeadAgentPrompt_NilConfig(t *testing.T) {
	prompt := BuildLeadAgentPrompt(nil)
	if prompt == nil {
		t.Error("BuildLeadAgentPrompt(nil) should not return nil")
	}
}

func TestBuildGeneralPurposeSubagentPrompt(t *testing.T) {
	prompt := BuildGeneralPurposeSubagentPrompt()
	if prompt == nil {
		t.Error("BuildGeneralPurposeSubagentPrompt() should not return nil")
	}
	rendered := prompt.Render()
	if !strings.Contains(rendered, "General-Purpose Subagent") {
		t.Error("Render() should contain General-Purpose Subagent")
	}
}

func TestBuildBashSubagentPrompt(t *testing.T) {
	prompt := BuildBashSubagentPrompt()
	if prompt == nil {
		t.Error("BuildBashSubagentPrompt() should not return nil")
	}
	rendered := prompt.Render()
	if !strings.Contains(rendered, "Bash Specialist Subagent") {
		t.Error("Render() should contain Bash Specialist Subagent")
	}
}

func TestDefaultLeadAgentConfig(t *testing.T) {
	config := DefaultLeadAgentConfig()
	if config.AgentName != "DeerFlow 2.0" {
		t.Errorf("AgentName = %v, want 'DeerFlow 2.0'", config.AgentName)
	}
	if config.SubagentEnabled != false {
		t.Error("SubagentEnabled should be false")
	}
	if config.MaxConcurrentSubagents != 3 {
		t.Errorf("MaxConcurrentSubagents = %v, want 3", config.MaxConcurrentSubagents)
	}
}

func TestPrompt_Sections(t *testing.T) {
	p := NewPrompt()
	section1 := NewBaseSection("test1", "content1")
	section2 := NewBaseSection("test2", "content2")
	p.AddSections(section1, section2)

	sections := p.Sections()
	if len(sections) != 2 {
		t.Errorf("Sections() len = %v, want 2", len(sections))
	}
}
