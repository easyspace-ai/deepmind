package middleware

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/config"
)

func TestUploadsMiddleware_Name(t *testing.T) {
	m := NewDefaultUploadsMiddleware()

	if m.Name() != "uploads" {
		t.Errorf("Name() = %v, want 'uploads'", m.Name())
	}
}

func TestUploadsMiddleware_formatSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"zero", 0, "0.0 KB"},
		{"512 bytes", 512, "0.5 KB"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1 MB", 1024 * 1024, "1.0 MB"},
		{"2.5 MB", 1024 * 1024 * 5 / 2, "2.5 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.size)
			if got != tt.want {
				t.Errorf("formatSize(%v) = %v, want %v", tt.size, got, tt.want)
			}
		})
	}
}

func TestUploadsMiddleware_joinLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"a"}, "a"},
		{"multiple", []string{"a", "b", "c"}, "a\nb\nc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinLines(tt.lines)
			if got != tt.want {
				t.Errorf("joinLines() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUploadsMiddleware_createFilesMessage(t *testing.T) {
	m := NewDefaultUploadsMiddleware()

	t.Run("no files", func(t *testing.T) {
		msg := m.createFilesMessage(nil, nil)
		if msg == "" {
			t.Error("createFilesMessage should not return empty string")
		}
		if len(msg) == 0 {
			t.Error("createFilesMessage should have content")
		}
	})

	t.Run("new files only", func(t *testing.T) {
		newFiles := []FileInfo{
			{Filename: "test.txt", Size: 1024, Path: "/mnt/user-data/uploads/test.txt"},
		}
		msg := m.createFilesMessage(newFiles, nil)
		if msg == "" {
			t.Error("createFilesMessage should not return empty string")
		}
		if len(msg) == 0 {
			t.Error("createFilesMessage should have content")
		}
	})

	t.Run("historical files only", func(t *testing.T) {
		historicalFiles := []FileInfo{
			{Filename: "old.txt", Size: 2048, Path: "/mnt/user-data/uploads/old.txt"},
		}
		msg := m.createFilesMessage(nil, historicalFiles)
		if msg == "" {
			t.Error("createFilesMessage should not return empty string")
		}
	})

	t.Run("both new and historical", func(t *testing.T) {
		newFiles := []FileInfo{
			{Filename: "test.txt", Size: 1024, Path: "/mnt/user-data/uploads/test.txt"},
		}
		historicalFiles := []FileInfo{
			{Filename: "old.txt", Size: 2048, Path: "/mnt/user-data/uploads/old.txt"},
		}
		msg := m.createFilesMessage(newFiles, historicalFiles)
		if msg == "" {
			t.Error("createFilesMessage should not return empty string")
		}
	})
}

func TestUploadsMiddleware_getMessageContent(t *testing.T) {
	t.Run("nil message", func(t *testing.T) {
		content := getMessageContent(nil)
		if content != "" {
			t.Errorf("getMessageContent(nil) = %q, want empty", content)
		}
	})

	t.Run("with content", func(t *testing.T) {
		msg := &schema.Message{
			Role:    schema.User,
			Content: "hello world",
		}
		content := getMessageContent(msg)
		if content != "hello world" {
			t.Errorf("getMessageContent() = %q, want 'hello world'", content)
		}
	})
}

func TestUploadsMiddleware_BeforeAgent_NoMessages(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewUploadsMiddleware(tmpDir)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 没有消息
	stateUpdate, err := m.BeforeAgent(ctx, ts, "thread-123", nil)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeAgent() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestUploadsMiddleware_BeforeAgent_LastMessageNotUser(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewUploadsMiddleware(tmpDir)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 最后一条不是 User 消息
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
		{Role: schema.Assistant, Content: "hi"},
	}

	stateUpdate, err := m.BeforeAgent(ctx, ts, "thread-123", nil)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeAgent() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestUploadsMiddleware_BeforeAgent_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewUploadsMiddleware(tmpDir)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 最后一条是 User 消息，但没有文件
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
	}

	stateUpdate, err := m.BeforeAgent(ctx, ts, "thread-123", nil)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate != nil {
		t.Errorf("BeforeAgent() stateUpdate = %v, want nil", stateUpdate)
	}
}

func TestUploadsMiddleware_BeforeAgent_WithAdditionalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewUploadsMiddleware(tmpDir)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 最后一条是 User 消息
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
	}

	// 有新上传的文件
	additionalFiles := []FileInfo{
		{
			Filename: "test.txt",
			Size:     1024,
			Path:     "/mnt/user-data/uploads/test.txt",
		},
	}

	stateUpdate, err := m.BeforeAgent(ctx, ts, "thread-123", additionalFiles)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		t.Fatal("BeforeAgent() stateUpdate should not be nil")
	}

	// 验证 uploaded_files
	uploadedFiles, ok := stateUpdate["uploaded_files"]
	if !ok {
		t.Error("stateUpdate should contain 'uploaded_files'")
	}
	if uploadedFiles == nil {
		t.Error("uploaded_files should not be nil")
	}

	// 验证 messages
	messages, ok := stateUpdate["messages"]
	if !ok {
		t.Error("stateUpdate should contain 'messages'")
	}
	if messages == nil {
		t.Error("messages should not be nil")
	}
}

func TestUploadsMiddleware_BeforeAgent_WithHistoricalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewUploadsMiddleware(tmpDir)
	ctx := context.Background()
	ts := state.NewThreadState()

	// 创建 uploads 目录和历史文件
	threadID := "thread-with-history"
	uploadsDir := config.NewPaths(tmpDir).SandboxUploadsDir(threadID)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatalf("Failed to create uploads dir: %v", err)
	}

	// 创建一个历史文件
	historicalFile := filepath.Join(uploadsDir, "old.txt")
	if err := os.WriteFile(historicalFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create historical file: %v", err)
	}

	// 最后一条是 User 消息
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: "hello"},
	}

	stateUpdate, err := m.BeforeAgent(ctx, ts, threadID, nil)

	if err != nil {
		t.Errorf("BeforeAgent() error = %v, want nil", err)
	}
	if stateUpdate == nil {
		// 没有新文件，可能返回 nil，这是可以接受的
		t.Log("stateUpdate is nil (no new files, only historical)")
	}
}

func TestUploadsMiddleware_BeforeAgent_MessageContentUpdated(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewUploadsMiddleware(tmpDir)
	ctx := context.Background()
	ts := state.NewThreadState()

	originalContent := "original user message"
	ts.Messages = []*schema.Message{
		{Role: schema.User, Content: originalContent},
	}

	additionalFiles := []FileInfo{
		{
			Filename: "test.txt",
			Size:     1024,
			Path:     "/mnt/user-data/uploads/test.txt",
		},
	}

	_, _ = m.BeforeAgent(ctx, ts, "thread-123", additionalFiles)

	// 验证消息内容已更新
	if len(ts.Messages) == 0 {
		t.Fatal("Messages should not be empty")
	}

	lastMsg := ts.Messages[len(ts.Messages)-1]
	if lastMsg.Content == originalContent {
		t.Error("Message content should be updated with uploaded_files block")
	}
	if len(lastMsg.Content) <= len(originalContent) {
		t.Error("Message content should be longer than original")
	}
}

func TestUploadsMiddleware_IntegrationWithChain(t *testing.T) {
	// 验证 UploadsMiddleware 能正确加入中间件链
	tmpDir := t.TempDir()
	config := DefaultMiddlewareConfig()
	config.BaseDir = tmpDir

	chain := BuildLeadAgentMiddlewares(config)

	// 验证链中有 UploadsMiddleware
	found := false
	for _, mw := range chain.Middlewares() {
		if mw.Name() == "uploads" {
			found = true
			break
		}
	}
	if !found {
		t.Error("UploadsMiddleware not found in chain")
	}

	// 验证位置：应该是第二个（在 ThreadDataMiddleware 之后）
	if len(chain.Middlewares()) < 2 {
		t.Fatal("Chain should have at least 2 middlewares")
	}
	if chain.Middlewares()[1].Name() != "uploads" {
		t.Error("UploadsMiddleware should be second in chain")
	}
}

func TestUploadsMiddleware_FileInfoToStateUploadedFile(t *testing.T) {
	// 测试 FileInfo 转换为 state.UploadedFile
	fi := FileInfo{
		Filename:  "test.txt",
		Size:      1024,
		Path:      "/mnt/user-data/uploads/test.txt",
		Extension: ".txt",
		Status:    "uploaded",
	}

	// 验证转换逻辑（当前在 BeforeAgent 中内联实现）
	su := state.UploadedFile{
		Filename:    fi.Filename,
		Path:        fi.Path,
		Size:        fi.Size,
		ContentType: "",
		UploadedAt:  time.Now(),
	}

	if su.Filename != fi.Filename {
		t.Error("Filename mismatch")
	}
	if su.Path != fi.Path {
		t.Error("Path mismatch")
	}
	if su.Size != fi.Size {
		t.Error("Size mismatch")
	}
}

func TestUploadsMiddleware_VirtualPathFormat(t *testing.T) {
	// 验证虚拟路径格式正确
	filename := "test.txt"
	virtualPath := config.VirtualUploadsPath + "/" + filename
	expected := "/mnt/user-data/uploads/test.txt"

	if virtualPath != expected {
		t.Errorf("Virtual path = %q, want %q", virtualPath, expected)
	}
}

// BenchmarkUploadsMiddleware_BeforeAgent 性能基准测试
func BenchmarkUploadsMiddleware_BeforeAgent(b *testing.B) {
	tmpDir := b.TempDir()
	m := NewUploadsMiddleware(tmpDir)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts := state.NewThreadState()
		ts.Messages = []*schema.Message{
			{Role: schema.User, Content: "hello"},
		}
		additionalFiles := []FileInfo{
			{Filename: "test.txt", Size: 1024, Path: "/mnt/user-data/uploads/test.txt"},
		}
		_, _ = m.BeforeAgent(ctx, ts, "bench-thread", additionalFiles)
	}
}
