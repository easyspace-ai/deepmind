package state

import (
	"github.com/cloudwego/eino/schema"
	"testing"
	"time"
)

func TestNewThreadState(t *testing.T) {
	state := NewThreadState()

	if state.Messages == nil {
		t.Error("Messages should not be nil")
	}
	if len(state.Messages) != 0 {
		t.Error("Messages should be empty")
	}
	if state.Artifacts == nil {
		t.Error("Artifacts should not be nil")
	}
	if len(state.Artifacts) != 0 {
		t.Error("Artifacts should be empty")
	}
	if state.ViewedImages == nil {
		t.Error("ViewedImages should not be nil")
	}
	if len(state.ViewedImages) != 0 {
		t.Error("ViewedImages should be empty")
	}
}

func TestMergeArtifacts(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		new      []string
		expected []string
	}{
		{
			name:     "both nil",
			existing: nil,
			new:      nil,
			expected: []string{},
		},
		{
			name:     "existing nil",
			existing: nil,
			new:      []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "new nil",
			existing: []string{"a", "b"},
			new:      nil,
			expected: []string{"a", "b"},
		},
		{
			name:     "no duplicates",
			existing: []string{"a", "b"},
			new:      []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "with duplicates",
			existing: []string{"a", "b", "c"},
			new:      []string{"b", "c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "preserves order",
			existing: []string{"c", "a", "b"},
			new:      []string{"a", "d"},
			expected: []string{"c", "a", "b", "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeArtifacts(tt.existing, tt.new)
			if len(result) != len(tt.expected) {
				t.Errorf("MergeArtifacts() len = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("MergeArtifacts()[%v] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestMergeViewedImages(t *testing.T) {
	tests := []struct {
		name     string
		existing map[string]ViewedImageData
		new      map[string]ViewedImageData
		expected map[string]ViewedImageData
	}{
		{
			name:     "both nil",
			existing: nil,
			new:      nil,
			expected: map[string]ViewedImageData{},
		},
		{
			name:     "existing nil",
			existing: nil,
			new: map[string]ViewedImageData{
				"img1": {Base64: "base64_1", MimeType: "image/jpeg"},
			},
			expected: map[string]ViewedImageData{
				"img1": {Base64: "base64_1", MimeType: "image/jpeg"},
			},
		},
		{
			name: "new empty map clears",
			existing: map[string]ViewedImageData{
				"img1": {Base64: "base64_1", MimeType: "image/jpeg"},
			},
			new:      map[string]ViewedImageData{},
			expected: map[string]ViewedImageData{},
		},
		{
			name: "merge with override",
			existing: map[string]ViewedImageData{
				"img1": {Base64: "base64_1", MimeType: "image/jpeg"},
				"img2": {Base64: "base64_2", MimeType: "image/png"},
			},
			new: map[string]ViewedImageData{
				"img1": {Base64: "base64_1_updated", MimeType: "image/jpeg"},
				"img3": {Base64: "base64_3", MimeType: "image/gif"},
			},
			expected: map[string]ViewedImageData{
				"img1": {Base64: "base64_1_updated", MimeType: "image/jpeg"},
				"img2": {Base64: "base64_2", MimeType: "image/png"},
				"img3": {Base64: "base64_3", MimeType: "image/gif"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeViewedImages(tt.existing, tt.new)
			if len(result) != len(tt.expected) {
				t.Errorf("MergeViewedImages() len = %v, want %v", len(result), len(tt.expected))
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("MergeViewedImages()[%v] = %v, want %v", k, result[k], v)
				}
			}
		})
	}
}

func TestThreadState_AddArtifacts(t *testing.T) {
	state := NewThreadState()
	state.AddArtifacts("a", "b")
	state.AddArtifacts("b", "c")

	expected := []string{"a", "b", "c"}
	if len(state.Artifacts) != len(expected) {
		t.Errorf("AddArtifacts() len = %v, want %v", len(state.Artifacts), len(expected))
	}
	for i := range expected {
		if state.Artifacts[i] != expected[i] {
			t.Errorf("AddArtifacts()[%v] = %v, want %v", i, state.Artifacts[i], expected[i])
		}
	}
}

func TestThreadState_AddViewedImages(t *testing.T) {
	state := NewThreadState()
	state.AddViewedImages(map[string]ViewedImageData{
		"img1": {Base64: "base64_1", MimeType: "image/jpeg"},
	})
	state.AddViewedImages(map[string]ViewedImageData{
		"img2": {Base64: "base64_2", MimeType: "image/png"},
	})

	if len(state.ViewedImages) != 2 {
		t.Errorf("AddViewedImages() len = %v, want 2", len(state.ViewedImages))
	}

	state.ClearViewedImages()
	if len(state.ViewedImages) != 0 {
		t.Errorf("ClearViewedImages() len = %v, want 0", len(state.ViewedImages))
	}
}

func TestThreadState_TodoOperations(t *testing.T) {
	state := NewThreadState()

	todo := TodoItem{
		ID:          "1",
		Description: "Test todo",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	state.AddTodo(todo)
	if len(state.Todos) != 1 {
		t.Errorf("AddTodo() len = %v, want 1", len(state.Todos))
	}

	found := state.GetTodoByID("1")
	if found == nil {
		t.Error("GetTodoByID() should find todo")
	}
	if found.Description != "Test todo" {
		t.Errorf("GetTodoByID() description = %v, want 'Test todo'", found.Description)
	}

	updated := state.UpdateTodoStatus("1", "in_progress")
	if !updated {
		t.Error("UpdateTodoStatus() should return true")
	}
	if state.Todos[0].Status != "in_progress" {
		t.Errorf("UpdateTodoStatus() status = %v, want 'in_progress'", state.Todos[0].Status)
	}
}

func TestThreadState_UploadedFileOperations(t *testing.T) {
	state := NewThreadState()

	file := UploadedFile{
		Filename:    "test.txt",
		Path:        "/path/to/test.txt",
		Size:        100,
		ContentType: "text/plain",
		UploadedAt:  time.Now(),
	}

	state.AddUploadedFile(file)
	if len(state.UploadedFiles) != 1 {
		t.Errorf("AddUploadedFile() len = %v, want 1", len(state.UploadedFiles))
	}

	found := state.GetUploadedFile("test.txt")
	if found == nil {
		t.Error("GetUploadedFile() should find file")
	}

	removed := state.RemoveUploadedFile("test.txt")
	if !removed {
		t.Error("RemoveUploadedFile() should return true")
	}
	if len(state.UploadedFiles) != 0 {
		t.Errorf("RemoveUploadedFile() len = %v, want 0", len(state.UploadedFiles))
	}
}

func TestThreadState_MessageOperations(t *testing.T) {
	state := NewThreadState()

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: "Hello",
	}
	assistantMsg := &schema.Message{
		Role:    schema.Assistant,
		Content: "Hi there",
	}

	state.AddMessage(userMsg)
	state.AddMessage(assistantMsg)

	if len(state.Messages) != 2 {
		t.Errorf("AddMessage() len = %v, want 2", len(state.Messages))
	}

	last := state.GetLastMessage()
	if last == nil {
		t.Error("GetLastMessage() should not be nil")
	}
	if last.Content != "Hi there" {
		t.Errorf("GetLastMessage() content = %v, want 'Hi there'", last.Content)
	}

	userMsgs := state.GetUserMessages()
	if len(userMsgs) != 1 {
		t.Errorf("GetUserMessages() len = %v, want 1", len(userMsgs))
	}

	assistantMsgs := state.GetAssistantMessages()
	if len(assistantMsgs) != 1 {
		t.Errorf("GetAssistantMessages() len = %v, want 1", len(assistantMsgs))
	}
}

func TestThreadState_TitleOperations(t *testing.T) {
	state := NewThreadState()

	if state.HasTitle() {
		t.Error("HasTitle() should be false initially")
	}

	state.SetTitle("Test Title")
	if !state.HasTitle() {
		t.Error("HasTitle() should be true after SetTitle")
	}
	if state.Title != "Test Title" {
		t.Errorf("Title = %v, want 'Test Title'", state.Title)
	}
}

func TestThreadState_SandboxOperations(t *testing.T) {
	state := NewThreadState()

	if state.Sandbox != nil {
		t.Error("Sandbox should be nil initially")
	}

	state.SetSandbox("sandbox-123")
	if state.Sandbox == nil {
		t.Error("Sandbox should not be nil after SetSandbox")
	}
	if state.Sandbox.SandboxID != "sandbox-123" {
		t.Errorf("SandboxID = %v, want 'sandbox-123'", state.Sandbox.SandboxID)
	}
}

func TestThreadState_ThreadDataOperations(t *testing.T) {
	state := NewThreadState()

	if state.ThreadData != nil {
		t.Error("ThreadData should be nil initially")
	}

	state.SetThreadData("/workspace", "/uploads", "/outputs")
	if state.ThreadData == nil {
		t.Error("ThreadData should not be nil after SetThreadData")
	}
	if state.GetWorkspacePath() != "/workspace" {
		t.Errorf("GetWorkspacePath() = %v, want '/workspace'", state.GetWorkspacePath())
	}
	if state.GetUploadsPath() != "/uploads" {
		t.Errorf("GetUploadsPath() = %v, want '/uploads'", state.GetUploadsPath())
	}
	if state.GetOutputsPath() != "/outputs" {
		t.Errorf("GetOutputsPath() = %v, want '/outputs'", state.GetOutputsPath())
	}
}
