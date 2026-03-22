package state

import (
	"github.com/cloudwego/eino/schema"
)

func mergeToolErrorFromMap(s *ThreadState, raw map[string]interface{}) {
	if s == nil || raw == nil {
		return
	}
	te := &ToolInvocationError{}
	if v, ok := raw["tool_name"].(string); ok {
		te.ToolName = v
	}
	if v, ok := raw["tool_call_id"].(string); ok {
		te.ToolCallID = v
	}
	if v, ok := raw["error"].(string); ok {
		te.Error = v
	}
	s.LastToolError = te
}

// ApplyMiddlewareUpdates 将 DeerFlow 风格中间件返回的 stateUpdate 合并进 ThreadState。
// 仅处理已约定的键；未知键忽略（可由调用方另行处理）。
func ApplyMiddlewareUpdates(s *ThreadState, u map[string]interface{}) {
	if s == nil || len(u) == 0 {
		return
	}
	if v, ok := u["messages"].([]*schema.Message); ok && v != nil {
		s.Messages = v
	}
	if v, ok := u["thread_data"].(*ThreadDataState); ok && v != nil {
		s.ThreadData = v
	}
	if v, ok := u["sandbox"].(*SandboxState); ok && v != nil {
		s.Sandbox = v
	}
	if v, ok := u["title"].(string); ok {
		s.Title = v
	}
	if v, ok := u["todos"].([]TodoItem); ok && v != nil {
		s.Todos = v
	}
	if v, ok := u["uploaded_files"].([]UploadedFile); ok && v != nil {
		s.UploadedFiles = v
	}
	if v, ok := u["viewed_images"].(map[string]ViewedImageData); ok && v != nil {
		s.ViewedImages = MergeViewedImages(s.ViewedImages, v)
	}
	if v, ok := u["artifacts"].([]string); ok && v != nil {
		s.Artifacts = MergeArtifacts(s.Artifacts, v)
	}
	if v, ok := u["interrupt"].(bool); ok && v {
		s.ClarificationPending = true
	}
	if v, ok := u["clarification_message"].(string); ok && v != "" {
		s.LastClarificationMessage = v
	}
	if raw, ok := u["tool_error"].(map[string]interface{}); ok && raw != nil {
		mergeToolErrorFromMap(s, raw)
	}
}
