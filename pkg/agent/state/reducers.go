package state

import (
	"github.com/cloudwego/eino/schema"
	"time"
)

// MergeArtifacts 合并产物列表，去重并保持顺序
// 一比一复刻 DeerFlow 的 merge_artifacts reducer
func MergeArtifacts(existing, new []string) []string {
	if existing == nil {
		if new == nil {
			return []string{}
		}
		return new
	}
	if new == nil {
		return existing
	}

	// 使用 map 去重，同时保持顺序
	seen := make(map[string]bool)
	result := make([]string, 0, len(existing)+len(new))

	// 先添加已存在的
	for _, item := range existing {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	// 再添加新的
	for _, item := range new {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// MergeViewedImages 合并已查看图像
// 一比一复刻 DeerFlow 的 merge_viewed_images reducer
// 特殊情况：如果 new 是空 map {}，则清空所有图像
func MergeViewedImages(existing, new map[string]ViewedImageData) map[string]ViewedImageData {
	if existing == nil {
		if new == nil {
			return make(map[string]ViewedImageData)
		}
		return new
	}
	if new == nil {
		return existing
	}

	// 特殊情况：空 map 表示清空所有
	if len(new) == 0 {
		return make(map[string]ViewedImageData)
	}

	// 合并 map，新值覆盖旧值
	result := make(map[string]ViewedImageData)
	for k, v := range existing {
		result[k] = v
	}
	for k, v := range new {
		result[k] = v
	}

	return result
}

// AddArtifacts 添加产物到状态
func (s *ThreadState) AddArtifacts(artifacts ...string) {
	s.Artifacts = MergeArtifacts(s.Artifacts, artifacts)
}

// AddViewedImages 添加已查看图像到状态
func (s *ThreadState) AddViewedImages(images map[string]ViewedImageData) {
	s.ViewedImages = MergeViewedImages(s.ViewedImages, images)
}

// ClearViewedImages 清空已查看图像
func (s *ThreadState) ClearViewedImages() {
	s.ViewedImages = make(map[string]ViewedImageData)
}

// AddTodo 添加待办事项
func (s *ThreadState) AddTodo(todo TodoItem) {
	s.Todos = append(s.Todos, todo)
}

// UpdateTodoStatus 更新待办事项状态
func (s *ThreadState) UpdateTodoStatus(id string, status string) bool {
	for i, todo := range s.Todos {
		if todo.ID == id {
			s.Todos[i].Status = status
			s.Todos[i].UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetTodoByID 根据 ID 获取待办事项
func (s *ThreadState) GetTodoByID(id string) *TodoItem {
	for _, todo := range s.Todos {
		if todo.ID == id {
			return &todo
		}
	}
	return nil
}

// AddUploadedFile 添加已上传文件
func (s *ThreadState) AddUploadedFile(file UploadedFile) {
	s.UploadedFiles = append(s.UploadedFiles, file)
}

// RemoveUploadedFile 移除已上传文件
func (s *ThreadState) RemoveUploadedFile(filename string) bool {
	for i, file := range s.UploadedFiles {
		if file.Filename == filename {
			s.UploadedFiles = append(s.UploadedFiles[:i], s.UploadedFiles[i+1:]...)
			return true
		}
	}
	return false
}

// GetUploadedFile 获取已上传文件
func (s *ThreadState) GetUploadedFile(filename string) *UploadedFile {
	for _, file := range s.UploadedFiles {
		if file.Filename == filename {
			return &file
		}
	}
	return nil
}

// AddMessage 添加消息
func (s *ThreadState) AddMessage(msg *schema.Message) {
	s.Messages = append(s.Messages, msg)
}

// GetLastMessage 获取最后一条消息
func (s *ThreadState) GetLastMessage() *schema.Message {
	if len(s.Messages) == 0 {
		return nil
	}
	return s.Messages[len(s.Messages)-1]
}

// GetUserMessages 获取所有用户消息
func (s *ThreadState) GetUserMessages() []*schema.Message {
	var result []*schema.Message
	for _, msg := range s.Messages {
		if msg.Role == schema.User {
			result = append(result, msg)
		}
	}
	return result
}

// GetAssistantMessages 获取所有助手消息
func (s *ThreadState) GetAssistantMessages() []*schema.Message {
	var result []*schema.Message
	for _, msg := range s.Messages {
		if msg.Role == schema.Assistant {
			result = append(result, msg)
		}
	}
	return result
}

// HasTitle 检查是否已有标题
func (s *ThreadState) HasTitle() bool {
	return s.Title != ""
}

// SetTitle 设置标题
func (s *ThreadState) SetTitle(title string) {
	s.Title = title
}

// SetSandbox 设置沙箱状态
func (s *ThreadState) SetSandbox(sandboxID string) {
	s.Sandbox = &SandboxState{
		SandboxID: sandboxID,
	}
}

// SetThreadData 设置线程数据
func (s *ThreadState) SetThreadData(workspacePath, uploadsPath, outputsPath string) {
	s.ThreadData = &ThreadDataState{
		WorkspacePath: workspacePath,
		UploadsPath:   uploadsPath,
		OutputsPath:   outputsPath,
	}
}

// GetWorkspacePath 获取工作区路径
func (s *ThreadState) GetWorkspacePath() string {
	if s.ThreadData == nil {
		return ""
	}
	return s.ThreadData.WorkspacePath
}

// GetUploadsPath 获取上传路径
func (s *ThreadState) GetUploadsPath() string {
	if s.ThreadData == nil {
		return ""
	}
	return s.ThreadData.UploadsPath
}

// GetOutputsPath 获取输出路径
func (s *ThreadState) GetOutputsPath() string {
	if s.ThreadData == nil {
		return ""
	}
	return s.ThreadData.OutputsPath
}
