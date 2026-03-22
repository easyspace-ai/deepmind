package middleware

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/config"
)

// UploadsMiddleware 上传文件追踪中间件
// 一比一复刻 DeerFlow 的 UploadsMiddleware
type UploadsMiddleware struct {
	*BaseMiddleware
	paths *config.Paths
}

// NewUploadsMiddleware 创建上传文件中间件
func NewUploadsMiddleware(baseDir string) *UploadsMiddleware {
	return &UploadsMiddleware{
		BaseMiddleware: NewBaseMiddleware("uploads"),
		paths:          config.NewPaths(baseDir),
	}
}

// NewDefaultUploadsMiddleware 使用默认配置创建上传文件中间件
func NewDefaultUploadsMiddleware() *UploadsMiddleware {
	return NewUploadsMiddleware("")
}

// FileInfo 上传文件信息
type FileInfo struct {
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	Path        string    `json:"path"`
	Extension   string    `json:"extension,omitempty"`
	Status      string    `json:"status,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at,omitempty"`
}

// createFilesMessage 创建文件列表消息
func (m *UploadsMiddleware) createFilesMessage(newFiles []FileInfo, historicalFiles []FileInfo) string {
	lines := []string{"<uploaded_files>"}
	lines = append(lines, "The following files were uploaded in this message:")
	lines = append(lines, "")

	if len(newFiles) > 0 {
		for _, file := range newFiles {
			sizeStr := formatSize(file.Size)
			lines = append(lines, fmt.Sprintf("- %s (%s)", file.Filename, sizeStr))
			lines = append(lines, fmt.Sprintf("  Path: %s", file.Path))
			lines = append(lines, "")
		}
	} else {
		lines = append(lines, "(empty)")
	}

	if len(historicalFiles) > 0 {
		lines = append(lines, "The following files were uploaded in previous messages and are still available:")
		lines = append(lines, "")
		for _, file := range historicalFiles {
			sizeStr := formatSize(file.Size)
			lines = append(lines, fmt.Sprintf("- %s (%s)", file.Filename, sizeStr))
			lines = append(lines, fmt.Sprintf("  Path: %s", file.Path))
			lines = append(lines, "")
		}
	}

	lines = append(lines, "You can read these files using the `read_file` tool with the paths shown above.")
	lines = append(lines, "</uploaded_files>")

	return joinLines(lines)
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	sizeKB := float64(size) / 1024
	if sizeKB < 1024 {
		return fmt.Sprintf("%.1f KB", sizeKB)
	}
	return fmt.Sprintf("%.1f MB", sizeKB/1024)
}

// joinLines 连接行
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

// BeforeAgent 在 Agent 执行前运行
func (m *UploadsMiddleware) BeforeAgent(ctx context.Context, ts *state.ThreadState, threadID string, additionalFiles []FileInfo) (map[string]interface{}, error) {
	if len(ts.Messages) == 0 {
		return nil, nil
	}

	lastMsg := ts.Messages[len(ts.Messages)-1]
	if lastMsg.Role != schema.User {
		return nil, nil
	}

	// 获取上传目录
	var uploadsDir string
	if threadID != "" {
		uploadsDir = m.paths.SandboxUploadsDir(threadID)
	}

	// 新上传的文件
	newFiles := additionalFiles
	if len(newFiles) == 0 {
		newFiles = []FileInfo{}
	}

	// 收集历史上传文件
	newFilenames := make(map[string]bool)
	for _, f := range newFiles {
		newFilenames[f.Filename] = true
	}

	historicalFiles := []FileInfo{}
	if uploadsDir != "" {
		if entries, err := os.ReadDir(uploadsDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && !newFilenames[entry.Name()] {
					if info, err := entry.Info(); err == nil {
						historicalFiles = append(historicalFiles, FileInfo{
							Filename:  entry.Name(),
							Size:      info.Size(),
							Path:      fmt.Sprintf("%s/%s", config.VirtualUploadsPath, entry.Name()),
							Extension: filepath.Ext(entry.Name()),
						})
					}
				}
			}
		}
	}

	if len(newFiles) == 0 && len(historicalFiles) == 0 {
		return nil, nil
	}

	// 创建文件消息并 prepend 到最后一条用户消息
	filesMessage := m.createFilesMessage(newFiles, historicalFiles)

	// 更新最后一条用户消息
	originalContent := getMessageContent(lastMsg)
	updatedContent := fmt.Sprintf("%s\n\n%s", filesMessage, originalContent)

	// 创建状态更新
	stateUpdate := make(map[string]interface{})

	// 更新 uploaded_files 状态
	if len(newFiles) > 0 {
		uploadedFiles := make([]state.UploadedFile, 0, len(newFiles))
		for _, f := range newFiles {
			uploadedFiles = append(uploadedFiles, state.UploadedFile{
				Filename:    f.Filename,
				Path:        f.Path,
				Size:        f.Size,
				ContentType: "", // TODO: 从文件扩展名推断
				UploadedAt:  time.Now(),
			})
		}
		ts.UploadedFiles = uploadedFiles
		stateUpdate["uploaded_files"] = uploadedFiles
	}

	// 更新消息
	// 注意：这里需要创建新的消息列表，因为 ts.Messages 是 slice
	updatedMessages := make([]*schema.Message, len(ts.Messages))
	copy(updatedMessages, ts.Messages)
	updatedMessages[len(updatedMessages)-1] = &schema.Message{
		Role:    schema.User,
		Content: updatedContent,
	}
	ts.Messages = updatedMessages
	stateUpdate["messages"] = updatedMessages

	return stateUpdate, nil
}

// getMessageContent 获取消息内容
func getMessageContent(msg *schema.Message) string {
	if msg == nil {
		return ""
	}
	return msg.Content
}
