package memory

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// ============================================
// Prompt Templates（一比一复刻 DeerFlow）
// ============================================

// MemoryUpdatePrompt 记忆更新提示词模板
const MemoryUpdatePrompt = `You are a memory management system. Your task is to analyze a conversation and update the user's memory profile.

Current Memory State:
<current_memory>
{current_memory}
</current_memory>

New Conversation to Process:
<conversation>
{conversation}
</conversation>

Instructions:
1. Analyze the conversation for important information about the user
2. Extract relevant facts, preferences, and context with specific details (numbers, names, technologies)
3. Update the memory sections as needed following the detailed length guidelines below

Memory Section Guidelines:

**User Context** (Current state - concise summaries):
- workContext: Professional role, company, key projects, main technologies (2-3 sentences)
  Example: Core contributor, project names with metrics (16k+ stars), technical stack
- personalContext: Languages, communication preferences, key interests (1-2 sentences)
  Example: Bilingual capabilities, specific interest areas, expertise domains
- topOfMind: Multiple ongoing focus areas and priorities (3-5 sentences, detailed paragraph)
  Example: Primary project work, parallel technical investigations, ongoing learning/tracking
  Include: Active implementation work, troubleshooting issues, market/research interests
  Note: This captures SEVERAL concurrent focus areas, not just one task

**History** (Temporal context - rich paragraphs):
- recentMonths: Detailed summary of recent activities (4-6 sentences or 1-2 paragraphs)
  Timeline: Last 1-3 months of interactions
  Include: Technologies explored, projects worked on, problems solved, interests demonstrated
- earlierContext: Important historical patterns (3-5 sentences or 1 paragraph)
  Timeline: 3-12 months ago
  Include: Past projects, learning journeys, established patterns
- longTermBackground: Persistent background and foundational context (2-4 sentences)
  Timeline: Overall/foundational information
  Include: Core expertise, longstanding interests, fundamental working style

**Facts Extraction**:
- Extract specific, quantifiable details (e.g., "16k+ GitHub stars", "200+ datasets")
- Include proper nouns (company names, project names, technology names)
- Preserve technical terminology and version numbers
- Categories:
  * preference: Tools, styles, approaches user prefers/dislikes
  * knowledge: Specific expertise, technologies mastered, domain knowledge
  * context: Background facts (job title, projects, locations, languages)
  * behavior: Working patterns, communication habits, problem-solving approaches
  * goal: Stated objectives, learning targets, project ambitions
- Confidence levels:
  * 0.9-1.0: Explicitly stated facts ("I work on X", "My role is Y")
  * 0.7-0.8: Strongly implied from actions/discussions
  * 0.5-0.6: Inferred patterns (use sparingly, only for clear patterns)

**What Goes Where**:
- workContext: Current job, active projects, primary tech stack
- personalContext: Languages, personality, interests outside direct work tasks
- topOfMind: Multiple ongoing priorities and focus areas user cares about recently (gets updated most frequently)
  Should capture 3-5 concurrent themes: main work, side explorations, learning/tracking interests
- recentMonths: Detailed account of recent technical explorations and work
- earlierContext: Patterns from slightly older interactions still relevant
- longTermBackground: Unchanging foundational facts about the user

**Multilingual Content**:
- Preserve original language for proper nouns and company names
- Keep technical terms in their original form (DeepSeek, LangGraph, etc.)
- Note language capabilities in personalContext

Output Format (JSON):
{
  "user": {
    "workContext": { "summary": "...", "shouldUpdate": true/false },
    "personalContext": { "summary": "...", "shouldUpdate": true/false },
    "topOfMind": { "summary": "...", "shouldUpdate": true/false }
  },
  "history": {
    "recentMonths": { "summary": "...", "shouldUpdate": true/false },
    "earlierContext": { "summary": "...", "shouldUpdate": true/false },
    "longTermBackground": { "summary": "...", "shouldUpdate": true/false }
  },
  "newFacts": [
    { "content": "...", "category": "preference|knowledge|context|behavior|goal", "confidence": 0.0-1.0 }
  ],
  "factsToRemove": ["fact_id_1", "fact_id_2"]
}

Important Rules:
- Only set shouldUpdate=true if there's meaningful new information
- Follow length guidelines: workContext/personalContext are concise (1-3 sentences), topOfMind and history sections are detailed (paragraphs)
- Include specific metrics, version numbers, and proper nouns in facts
- Only add facts that are clearly stated (0.9+) or strongly implied (0.7+)
- Remove facts that are contradicted by new information
- When updating topOfMind, integrate new focus areas while removing completed/abandoned ones
  Keep 3-5 concurrent focus themes that are still active and relevant
- For history sections, integrate new information chronologically into appropriate time period
- Preserve technical accuracy - keep exact names of technologies, companies, projects
- Focus on information useful for future interactions and personalization
- IMPORTANT: Do NOT record file upload events in memory. Uploaded files are
  session-specific and ephemeral — they will not be accessible in future sessions.
  Recording upload events causes confusion in subsequent conversations.

Return ONLY valid JSON, no explanation or markdown.`

// FactExtractionPrompt 单条消息事实提取提示词
const FactExtractionPrompt = `Extract factual information about the user from this message.

Message:
{message}

Extract facts in this JSON format:
{
  "facts": [
    { "content": "...", "category": "preference|knowledge|context|behavior|goal", "confidence": 0.0-1.0 }
  ]
}

Categories:
- preference: User preferences (likes/dislikes, styles, tools)
- knowledge: User's expertise or knowledge areas
- context: Background context (location, job, projects)
- behavior: Behavioral patterns
- goal: User's goals or objectives

Rules:
- Only extract clear, specific facts
- Confidence should reflect certainty (explicit statement = 0.9+, implied = 0.6-0.8)
- Skip vague or temporary information

Return ONLY valid JSON.`

// ============================================
// 上传提及过滤（一比一复刻 DeerFlow）
// ============================================

var (
	// uploadSentenceRegex 匹配描述文件上传事件的句子
	uploadSentenceRegex = regexp.MustCompile(`(?i)[^.!?]*\b(?:` +
		`upload(?:ed|ing)?(?:\s+\w+){0,3}\s+(?:file|files?|document|documents?|attachment|attachments?)` +
		`|file\s+upload` +
		`|/mnt/user-data/uploads/` +
		`|<uploaded_files>` +
		`)[^.!?]*[.!?]?\s*`)
)

// StripUploadMentions 从记忆中移除上传提及
func StripUploadMentions(memoryData *MemoryData) *MemoryData {
	if memoryData == nil {
		return nil
	}

	// 清理 user/history 节中的摘要
	userSections := map[string]*MemorySection{
		"workContext":     &memoryData.User.WorkContext,
		"personalContext": &memoryData.User.PersonalContext,
		"topOfMind":       &memoryData.User.TopOfMind,
	}
	for _, section := range userSections {
		section.Summary = stripUploadSentences(section.Summary)
	}

	historySections := map[string]*MemorySection{
		"recentMonths":       &memoryData.History.RecentMonths,
		"earlierContext":     &memoryData.History.EarlierContext,
		"longTermBackground": &memoryData.History.LongTermBackground,
	}
	for _, section := range historySections {
		section.Summary = stripUploadSentences(section.Summary)
	}

	// 清理 facts
	filteredFacts := make([]Fact, 0, len(memoryData.Facts))
	for _, fact := range memoryData.Facts {
		if !uploadSentenceRegex.MatchString(fact.Content) {
			filteredFacts = append(filteredFacts, fact)
		}
	}
	memoryData.Facts = filteredFacts

	return memoryData
}

func stripUploadSentences(text string) string {
	if text == "" {
		return ""
	}
	cleaned := uploadSentenceRegex.ReplaceAllString(text, "")
	// 合并多个空格
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	return strings.TrimSpace(cleaned)
}

// ============================================
// 对话格式化（一比一复刻 DeerFlow）
// ============================================

var uploadedFilesTagRegex = regexp.MustCompile(`(?s)<uploaded_files>[\s\S]*?</uploaded_files>\n*`)

// FormatConversationForUpdate 格式化对话用于记忆更新
func FormatConversationForUpdate(messages []*schema.Message) string {
	lines := make([]string, 0, len(messages))

	for _, msg := range messages {
		if msg == nil {
			continue
		}

		role := string(msg.Role)
		content := msg.Content

		// 处理多模态内容列表
		if len(msg.MultiContent) > 0 {
			textParts := make([]string, 0)
			for _, part := range msg.MultiContent {
				if part.Type == "text" && part.Text != "" {
					textParts = append(textParts, part.Text)
				}
			}
			if len(textParts) > 0 {
				content = strings.Join(textParts, " ")
			}
		}

		// 从用户消息中移除 uploaded_files 标签
		if role == string(schema.User) {
			content = uploadedFilesTagRegex.ReplaceAllString(content, "")
			content = strings.TrimSpace(content)
			if content == "" {
				continue // 跳过仅包含上传的消息
			}
		}

		// 截断过长消息
		if len(content) > 1000 {
			content = content[:1000] + "..."
		}

		// 标准化角色名称
		displayRole := "User"
		if role == string(schema.Assistant) {
			displayRole = "Assistant"
		}

		lines = append(lines, fmt.Sprintf("%s: %s", displayRole, content))
	}

	return strings.Join(lines, "\n\n")
}

// ============================================
// 记忆格式化与注入（一比一复刻 DeerFlow）
// ============================================

// FormatMemoryForInjection 格式化记忆用于提示词注入
func FormatMemoryForInjection(memoryData *MemoryData, maxTokens int) string {
	if memoryData == nil {
		return ""
	}

	sections := make([]string, 0, 3)

	// 格式化 user context
	userSections := make([]string, 0, 3)
	if memoryData.User.WorkContext.Summary != "" {
		userSections = append(userSections, fmt.Sprintf("Work: %s", memoryData.User.WorkContext.Summary))
	}
	if memoryData.User.PersonalContext.Summary != "" {
		userSections = append(userSections, fmt.Sprintf("Personal: %s", memoryData.User.PersonalContext.Summary))
	}
	if memoryData.User.TopOfMind.Summary != "" {
		userSections = append(userSections, fmt.Sprintf("Current Focus: %s", memoryData.User.TopOfMind.Summary))
	}
	if len(userSections) > 0 {
		var sb strings.Builder
		sb.WriteString("User Context:\n")
		for _, s := range userSections {
			sb.WriteString(fmt.Sprintf("- %s\n", s))
		}
		sections = append(sections, sb.String())
	}

	// 格式化 history
	historySections := make([]string, 0, 2)
	if memoryData.History.RecentMonths.Summary != "" {
		historySections = append(historySections, fmt.Sprintf("Recent: %s", memoryData.History.RecentMonths.Summary))
	}
	if memoryData.History.EarlierContext.Summary != "" {
		historySections = append(historySections, fmt.Sprintf("Earlier: %s", memoryData.History.EarlierContext.Summary))
	}
	if len(historySections) > 0 {
		var sb strings.Builder
		sb.WriteString("History:\n")
		for _, s := range historySections {
			sb.WriteString(fmt.Sprintf("- %s\n", s))
		}
		sections = append(sections, sb.String())
	}

	// 格式化 facts（按置信度排序，token 预算控制）
	if len(memoryData.Facts) > 0 {
		// 按置信度降序排序
		rankedFacts := make([]Fact, 0, len(memoryData.Facts))
		for _, f := range memoryData.Facts {
			if f.Content != "" {
				rankedFacts = append(rankedFacts, f)
			}
		}
		sort.Slice(rankedFacts, func(i, j int) bool {
			return coerceConfidence(rankedFacts[i].Confidence, 0.0) > coerceConfidence(rankedFacts[j].Confidence, 0.0)
		})

		// 计算基础 token 数
		baseText := strings.Join(sections, "\n\n")
		baseTokens := countTokens(baseText)
		factsHeader := "Facts:\n"
		separatorTokens := 0
		if baseText != "" {
			separatorTokens = countTokens("\n\n" + factsHeader)
		} else {
			separatorTokens = countTokens(factsHeader)
		}
		runningTokens := baseTokens + separatorTokens

		factLines := make([]string, 0, len(rankedFacts))
		for _, fact := range rankedFacts {
			content := strings.TrimSpace(fact.Content)
			if content == "" {
				continue
			}
			category := strings.TrimSpace(fact.Category)
			if category == "" {
				category = "context"
			}
			confidence := coerceConfidence(fact.Confidence, 0.0)
			line := fmt.Sprintf("- [%s | %.2f] %s", category, confidence, content)

			// 计算增量 token
			lineText := line
			if len(factLines) > 0 {
				lineText = "\n" + line
			}
			lineTokens := countTokens(lineText)

			if runningTokens+lineTokens <= maxTokens {
				factLines = append(factLines, line)
				runningTokens += lineTokens
			} else {
				break
			}
		}

		if len(factLines) > 0 {
			var sb strings.Builder
			sb.WriteString("Facts:\n")
			sb.WriteString(strings.Join(factLines, "\n"))
			sections = append(sections, sb.String())
		}
	}

	if len(sections) == 0 {
		return ""
	}

	result := strings.Join(sections, "\n\n")

	// 最终 token 检查
	tokenCount := countTokens(result)
	if tokenCount > maxTokens {
		// 简单截断
		charPerToken := float64(len(result)) / float64(tokenCount)
		targetChars := int(float64(maxTokens) * charPerToken * 0.95)
		if targetChars > 0 && targetChars < len(result) {
			result = result[:targetChars] + "\n..."
		}
	}

	return result
}

// ============================================
// Token 计数（简易实现）
// ============================================

// countTokens 简易 token 计数（基于字符数估算）
func countTokens(text string) int {
	// 简易估算：4 字符 ≈ 1 token
	return len(text) / 4
}

// coerceConfidence 强制转换置信度到 [0, 1] 范围
func coerceConfidence(value float64, defaultValue float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		value = defaultValue
	}
	if value < 0.0 {
		return 0.0
	}
	if value > 1.0 {
		return 1.0
	}
	return value
}

// ============================================
// 提示词构建辅助函数
// ============================================

// BuildMemoryUpdatePrompt 构建记忆更新提示词
func BuildMemoryUpdatePrompt(currentMemory *MemoryData, conversation string) (string, error) {
	memoryJSON, err := json.MarshalIndent(currentMemory, "", "  ")
	if err != nil {
		return "", err
	}

	prompt := strings.ReplaceAll(MemoryUpdatePrompt, "{current_memory}", string(memoryJSON))
	prompt = strings.ReplaceAll(prompt, "{conversation}", conversation)
	return prompt, nil
}

// BuildFactExtractionPrompt 构建事实提取提示词
func BuildFactExtractionPrompt(message string) string {
	return strings.ReplaceAll(FactExtractionPrompt, "{message}", message)
}
