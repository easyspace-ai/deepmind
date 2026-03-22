package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDir_SKILLmd(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "demo")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	content := "---\nname: Demo\ndescription: Test skill\n---\n\n# Body\n"
	require.NoError(t, os.WriteFile(filepath.Join(sub, "SKILL.md"), []byte(content), 0o644))

	list, err := LoadDir(dir)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Demo", list[0].Name)
	require.Contains(t, list[0].Body, "# Body")
}
