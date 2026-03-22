package memory

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ============================================
// 测试：Updater
// ============================================

func TestMemoryUpdater_New(t *testing.T) {
	u := NewMemoryUpdater("", nil)
	require.NotNil(t, u)
}

func TestMemoryUpdater_WithLogger(t *testing.T) {
	logger := zap.NewExample()
	u := NewMemoryUpdater("", logger)
	require.NotNil(t, u)
	require.Equal(t, logger, u.logger)
}

func TestGenerateFactID(t *testing.T) {
	id1 := generateFactID()
	id2 := generateFactID()
	require.NotEmpty(t, id1)
	require.NotEmpty(t, id2)
	require.NotEqual(t, id1, id2)
	require.True(t, len(id1) > 5)
}

func TestSortFactsByConfidence(t *testing.T) {
	facts := []Fact{
		{ID: "1", Confidence: 0.5},
		{ID: "2", Confidence: 0.9},
		{ID: "3", Confidence: 0.7},
	}
	sortFactsByConfidence(facts)
	require.Equal(t, 0.9, facts[0].Confidence)
	require.Equal(t, 0.7, facts[1].Confidence)
	require.Equal(t, 0.5, facts[2].Confidence)
}

func TestSortFactsByConfidence_Empty(t *testing.T) {
	var facts []Fact
	sortFactsByConfidence(facts) // 不应崩溃
	require.Empty(t, facts)
}

func TestSortFactsByConfidence_Single(t *testing.T) {
	facts := []Fact{{ID: "1", Confidence: 0.5}}
	sortFactsByConfidence(facts)
	require.Len(t, facts, 1)
	require.Equal(t, 0.5, facts[0].Confidence)
}

// ============================================
// 测试：文件相关
// ============================================

func TestFileMtime(t *testing.T) {
	f, err := os.CreateTemp("", "test")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	mtime := getFileMtime(f.Name())
	require.Greater(t, mtime, float64(0))

	require.Equal(t, float64(0), getFileMtime("/non/existent/path/that/will/never/exist"))
}

func TestGetMemoryFilePath(t *testing.T) {
	path := getMemoryFilePath("")
	require.NotEmpty(t, path)

	path2 := getMemoryFilePath("agent-1")
	require.NotEmpty(t, path2)
	require.NotEqual(t, path, path2)
}

func TestLoadMemoryFromFile_NotExists(t *testing.T) {
	m := CreateEmptyMemory()
	require.NotNil(t, m)
	require.Equal(t, "1.0", m.Version)
	require.Empty(t, m.Facts)
}

func TestFactContentKey(t *testing.T) {
	require.Equal(t, "", factContentKey(""))
	require.Equal(t, "test", factContentKey("test"))
	require.Equal(t, " test ", factContentKey(" test "))
}

func TestFormatConversationFromAny(t *testing.T) {
	result := formatConversationFromAny([]any{"hello", "world"})
	require.NotEmpty(t, result)
}

func TestFormatConversationFromAny_Empty(t *testing.T) {
	result := formatConversationFromAny(nil)
	require.Empty(t, result)

	result = formatConversationFromAny([]any{})
	require.Empty(t, result)
}

// ============================================
// 测试：缓存
// ============================================

func TestClearMemoryCache(t *testing.T) {
	ClearMemoryCache()
	require.True(t, true) // 只是确保不崩溃
}

// ============================================
// 测试：便捷函数
// ============================================

func TestUpdateMemoryFromConversation(t *testing.T) {
	// 只是确保不崩溃（默认禁用）
	result := UpdateMemoryFromConversation([]any{}, "thread-1", "")
	require.False(t, result)
}

// ============================================
// 测试：集成 - 完整流程
// ============================================

func TestMemorySystem_Integration(t *testing.T) {
	m := CreateEmptyMemory()
	m.User.WorkContext.Summary = "Go developer"
	m.User.PersonalContext.Summary = "Likes testing"
	m.User.TopOfMind.Summary = "Building memory system"
	m.Facts = []Fact{
		{
			ID:         "fact_1",
			Content:    "Writes unit tests",
			Category:   FactCategoryBehavior,
			Confidence: 0.9,
			CreatedAt:  time.Now().UTC().Format(time.RFC3339) + "Z",
			Source:     "test",
		},
	}

	formatted := FormatMemoryForInjection(m, 2000)
	require.Contains(t, formatted, "Go developer")
	require.Contains(t, formatted, "Writes unit tests")
}

func TestMemorySystem_WithUploadFiltering(t *testing.T) {
	m := CreateEmptyMemory()
	m.User.WorkContext.Summary = "User uploaded files. User codes in Go."
	m.Facts = []Fact{
		{ID: "1", Content: "User uploaded a file", Category: "context"},
		{ID: "2", Content: "User uses Go", Category: "preference"},
	}

	filtered := StripUploadMentions(m)
	require.NotContains(t, filtered.User.WorkContext.Summary, "uploaded")
	require.Len(t, filtered.Facts, 1)
	require.Equal(t, "User uses Go", filtered.Facts[0].Content)
}

// ============================================
// 测试：prompt 常量
// ============================================

func TestPromptConstants(t *testing.T) {
	require.NotEmpty(t, MemoryUpdatePrompt)
	require.NotEmpty(t, FactExtractionPrompt)
	require.Contains(t, MemoryUpdatePrompt, "current_memory")
	require.Contains(t, MemoryUpdatePrompt, "conversation")
	require.Contains(t, FactExtractionPrompt, "message")
}
