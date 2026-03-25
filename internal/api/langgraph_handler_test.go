package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestStreamRunFallbackCompletesSSESequence(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	handler := NewLangGraphHandler(zap.NewNop())
	handler.SetConfigLoader(nil)
	handler.SetDB(nil)

	body := []byte(`{
		"assistant_id":"lead_agent",
		"messages":[
			{
				"type":"human",
				"content":[{"type":"text","text":"调研一下deer-flow"}]
			}
		]
	}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads/test-thread/runs/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "threadId", Value: "test-thread"}}

	handler.streamRun(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	streamBody := recorder.Body.String()
	expectedEvents := []string{
		"event: metadata",
		"event: messages-tuple",
		"event: values",
		"event: end",
	}

	for _, expected := range expectedEvents {
		if !strings.Contains(streamBody, expected) {
			t.Fatalf("expected stream to contain %q, body=%s", expected, streamBody)
		}
	}
}
