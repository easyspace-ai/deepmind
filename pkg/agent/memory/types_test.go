package memory

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================
// 测试：Memory 数据结构
// ============================================

func TestCreateEmptyMemory(t *testing.T) {
	m := CreateEmptyMemory()
	require.NotNil(t, m)
	require.Equal(t, "1.0", m.Version)
	require.NotEmpty(t, m.LastUpdated)
	require.Empty(t, m.User.WorkContext.Summary)
	require.Empty(t, m.Facts)
}

func TestFactCategories(t *testing.T) {
	require.Equal(t, "preference", FactCategoryPreference)
	require.Equal(t, "knowledge", FactCategoryKnowledge)
	require.Equal(t, "context", FactCategoryContext)
	require.Equal(t, "behavior", FactCategoryBehavior)
	require.Equal(t, "goal", FactCategoryGoal)
}

func TestMemoryData_Serialization(t *testing.T) {
	m := CreateEmptyMemory()
	m.User.WorkContext.Summary = "Test summary"
	m.Facts = []Fact{
		{ID: "1", Content: "Test fact", Category: "context", Confidence: 0.9},
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)
	require.Contains(t, string(data), "Test summary")
	require.Contains(t, string(data), "Test fact")

	var m2 MemoryData
	err = json.Unmarshal(data, &m2)
	require.NoError(t, err)
	require.Equal(t, m.User.WorkContext.Summary, m2.User.WorkContext.Summary)
	require.Len(t, m2.Facts, 1)
}
