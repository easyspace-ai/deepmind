package prompts

// PromptSection 提示词分段接口
// 一比一复刻 DeerFlow 的模块化提示词设计
type PromptSection interface {
	// Name 返回分段名称
	Name() string
	// Render 渲染分段内容
	Render() string
}

// BaseSection 基础分段实现
type BaseSection struct {
	name    string
	content string
}

// NewBaseSection 创建基础分段
func NewBaseSection(name, content string) *BaseSection {
	return &BaseSection{
		name:    name,
		content: content,
	}
}

// Name 实现 PromptSection 接口
func (s *BaseSection) Name() string {
	return s.name
}

// Render 实现 PromptSection 接口
func (s *BaseSection) Render() string {
	return s.content
}

// NamedSection 函数式分段
type NamedSection struct {
	name     string
	renderFn func() string
}

// NewNamedSection 创建函数式分段
func NewNamedSection(name string, renderFn func() string) *NamedSection {
	return &NamedSection{
		name:     name,
		renderFn: renderFn,
	}
}

// Name 实现 PromptSection 接口
func (s *NamedSection) Name() string {
	return s.name
}

// Render 实现 PromptSection 接口
func (s *NamedSection) Render() string {
	if s.renderFn == nil {
		return ""
	}
	return s.renderFn()
}

// Prompt 完整提示词
type Prompt struct {
	sections []PromptSection
}

// NewPrompt 创建空提示词
func NewPrompt() *Prompt {
	return &Prompt{
		sections: make([]PromptSection, 0),
	}
}

// AddSection 添加分段
func (p *Prompt) AddSection(section PromptSection) *Prompt {
	p.sections = append(p.sections, section)
	return p
}

// AddSections 添加多个分段
func (p *Prompt) AddSections(sections ...PromptSection) *Prompt {
	p.sections = append(p.sections, sections...)
	return p
}

// GetSection 获取指定名称的分段
func (p *Prompt) GetSection(name string) PromptSection {
	for _, section := range p.sections {
		if section.Name() == name {
			return section
		}
	}
	return nil
}

// RemoveSection 移除指定名称的分段
func (p *Prompt) RemoveSection(name string) *Prompt {
	var newSections []PromptSection
	for _, section := range p.sections {
		if section.Name() != name {
			newSections = append(newSections, section)
		}
	}
	p.sections = newSections
	return p
}

// Render 渲染完整提示词
func (p *Prompt) Render() string {
	var parts []string
	for _, section := range p.sections {
		content := section.Render()
		if content != "" {
			parts = append(parts, content)
		}
	}
	return joinSections(parts)
}

// Sections 获取所有分段
func (p *Prompt) Sections() []PromptSection {
	return p.sections
}

// joinSections 连接分段内容
func joinSections(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var result string
	for i, part := range parts {
		if i > 0 {
			result += "\n\n"
		}
		result += part
	}
	return result
}
