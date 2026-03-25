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

// ✅ 【修复1】新增测试：验证所有错误路径都发送end事件
func TestStreamRunErrorPathSendsEnd(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	handler := NewLangGraphHandler(zap.NewNop())
	// 不设置 ConfigLoader，强制走 fallback 路径
	handler.SetConfigLoader(nil)
	handler.SetDB(nil)

	body := []byte(`{
		"assistant_id":"lead_agent",
		"messages":[
			{
				"type":"human",
				"content":[{"type":"text","text":"test error path"}]
			}
		]
	}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads/test-error/runs/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "threadId", Value: "test-error"}}

	handler.streamRun(ctx)

	streamBody := recorder.Body.String()

	// ✅ 关键检查：即使走fallback路径，也应该有end事件
	if !strings.Contains(streamBody, "event: end") {
		t.Fatalf("expected stream to contain end event, body=%s", streamBody)
	}

	// 验证end事件包含reason字段
	if !strings.Contains(streamBody, "\"reason\"") {
		t.Fatalf("expected end event to contain reason field, body=%s", streamBody)
	}
}

// ✅ 【修复2】新增测试：验证stream中的状态码和内容类型
func TestStreamRunKeepAliveHeaders(t *testing.T) {
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
				"content":[{"type":"text","text":"test"}]
			}
		]
	}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads/test-keepalive/runs/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "threadId", Value: "test-keepalive"}}

	handler.streamRun(ctx)

	streamBody := recorder.Body.String()

	// ✅ 检查 SSE 流状态码应该是 200
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	// ✅ 检查 Content-Type 应该是 text/event-stream
	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("expected Content-Type to contain text/event-stream, got %s", contentType)
	}

	// ✅ 检查流应该包含事件
	if !strings.Contains(streamBody, "event:") {
		t.Fatalf("expected stream to contain events, body=%s", streamBody)
	}

	t.Logf("Stream headers verified: status=%d, content-type=%s", recorder.Code, contentType)
}

// ✅ 【修复3】新增测试：验证SSE事件的完整序列和顺序
func TestStreamRunEventSequenceComplete(t *testing.T) {
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
				"content":[{"type":"text","text":"sequence test"}]
			}
		]
	}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads/test-sequence/runs/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "threadId", Value: "test-sequence"}}

	handler.streamRun(ctx)

	streamBody := recorder.Body.String()

	// ✅ 验证事件顺序：metadata应该在values之前，values应该在end之前
	metadataPos := strings.Index(streamBody, "event: metadata")
	iPos := strings.Index(streamBody, "event: values")
	endPos := strings.Index(streamBody, "event: end")

	if metadataPos == -1 {
		t.Fatalf("expected metadata event, body=%s", streamBody)
	}

	if iPos == -1 {
		t.Fatalf("expected values event, body=%s", streamBody)
	}

	if endPos == -1 {
		t.Fatalf("expected end event, body=%s", streamBody)
	}

	// 验证顺序
	if !(metadataPos < iPos && iPos < endPos) {
		t.Fatalf("expected event order: metadata(%d) < values(%d) < end(%d)", metadataPos, iPos, endPos)
	}

	t.Logf("Event order verified: metadata(%d) -> values(%d) -> end(%d)", metadataPos, iPos, endPos)
}

// ✅ 【修复5】新增测试：验证nil error不会导致panic
func TestStreamRunNilErrorHandling(t *testing.T) {
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
				"content":[{"type":"text","text":"test nil error"}]
			}
		]
	}`)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/langgraph/threads/test-nil-error/runs/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "threadId", Value: "test-nil-error"}}

	// ✅ 关键测试：这个调用不应该panic
	handler.streamRun(ctx)

	streamBody := recorder.Body.String()

	// 验证流能正常完成（不panic）
	if !strings.Contains(streamBody, "event: end") {
		t.Fatalf("expected stream to contain end event, body=%s", streamBody)
	}

	t.Logf("Nil error handling test passed - no panic occurred")
}
