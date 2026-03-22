package deerflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/config"
)

func TestViewImageTool_Invoke_ReadsFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.png"), []byte{0x89, 0x50, 0x4e, 0x47}, 0o644))

	ts := state.NewThreadState()
	ts.ThreadData = &state.ThreadDataState{WorkspacePath: dir}
	cfg := &ToolConfig{ThreadState: ts}
	tool := NewViewImageTool(cfg).(*ViewImageTool)

	out, err := tool.Invoke(context.Background(), map[string]interface{}{
		"path": config.VirtualWorkspacePath + "/a.png",
	})
	require.NoError(t, err)
	m, ok := out.(map[string]interface{})
	require.True(t, ok)
	require.NotEmpty(t, m["content"])
	vi, ok := m["viewed_images"].(map[string]state.ViewedImageData)
	require.True(t, ok)
	entry := vi[config.VirtualWorkspacePath+"/a.png"]
	require.NotEmpty(t, entry.Base64)
	require.Contains(t, entry.MimeType, "image/")
}
