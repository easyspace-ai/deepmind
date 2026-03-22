package skills

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// metaOnly 用于解析 YAML frontmatter（--- ... ---）。
type metaOnly struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// LoadDir 递归扫描目录，加载所有 SKILL.md（解析简单 frontmatter + 正文）。
func LoadDir(root string) ([]Skill, error) {
	if root == "" {
		return nil, nil
	}
	var out []Skill
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.EqualFold(d.Name(), "SKILL.md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		s := parseSkillFile(path, string(data))
		out = append(out, s)
		return nil
	})
	return out, nil
}

func parseSkillFile(path, content string) Skill {
	s := Skill{Path: path, Body: content}
	trim := strings.TrimSpace(content)
	if !strings.HasPrefix(trim, "---") {
		return s
	}
	end := strings.Index(trim[3:], "\n---")
	if end < 0 {
		return s
	}
	yamlBlock := strings.TrimSpace(trim[3 : 3+end])
	rest := strings.TrimSpace(trim[3+end+4:])
	var meta metaOnly
	_ = yaml.Unmarshal([]byte(yamlBlock), &meta)
	s.Name = meta.Name
	s.Description = meta.Description
	s.Body = rest
	return s
}

// PromptSection 生成注入提示词的一段文本。
func PromptSection(skills []Skill) string {
	if len(skills) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("<skill_system>\n")
	for _, sk := range skills {
		b.WriteString("- ")
		if sk.Name != "" {
			b.WriteString(sk.Name)
		} else {
			b.WriteString(filepath.Base(filepath.Dir(sk.Path)))
		}
		if sk.Description != "" {
			b.WriteString(": ")
			b.WriteString(sk.Description)
		}
		b.WriteByte('\n')
	}
	b.WriteString("</skill_system>\n")
	return b.String()
}
