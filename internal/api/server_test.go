package api

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewServerDisablesWriteTimeoutForSSE(t *testing.T) {
	t.Helper()

	server := NewServer(":7100", &Providers{}, zap.NewNop())

	if server.server.WriteTimeout != 0 {
		t.Fatalf("expected WriteTimeout to be 0 for SSE compatibility, got %s", server.server.WriteTimeout)
	}

	if server.server.ReadTimeout != 15*time.Second {
		t.Fatalf("expected ReadTimeout to remain 15s, got %s", server.server.ReadTimeout)
	}
}
