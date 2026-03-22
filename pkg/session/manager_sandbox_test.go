package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weibaohui/nanobot-go/config"
	"go.uber.org/zap"
)

func TestManager_SetSandboxReleaser_OnCancelSession(t *testing.T) {
	m := NewManager(&config.Config{}, zap.NewNop(), nil)
	var released string
	m.SetSandboxReleaser(func(key string) { released = key })

	s := m.GetOrCreate("sk-1")
	ctx, cancel := context.WithCancel(context.Background())
	s.SetContext(ctx, cancel)

	ok := m.CancelSession("sk-1")
	require.True(t, ok)
	require.Equal(t, "sk-1", released)
}
