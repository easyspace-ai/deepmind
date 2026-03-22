package state

import (
	"github.com/cloudwego/eino/schema"
	"time"
)

// SandboxState 沙箱状态
type SandboxState struct {
	SandboxID string `json:"sandbox_id,omitempty"`
}

// ThreadDataState 线程数据状态
type ThreadDataState struct {
	WorkspacePath string `json:"workspace_path,omitempty"`
	UploadsPath   string `json:"uploads_path,omitempty"`
	OutputsPath   string `json:"outputs_path,omitempty"`
}

// ViewedImageData 已查看图像数据
type ViewedImageData struct {
	Base64   string `json:"base64"`
	MimeType string `json:"mime_type"`
}

// TodoItem 待办事项
type TodoItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // pending, in_progress, completed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UploadedFile 已上传文件
type UploadedFile struct {
	Filename    string    `json:"filename"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// ToolInvocationError 最近一次工具调用失败快照（供 UI / 编排读取；不替代对话内 tool 消息）。
type ToolInvocationError struct {
	ToolName   string `json:"tool_name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ThreadState DeerFlow 完整线程状态
// 一比一复刻 DeerFlow 的 ThreadState 结构
type ThreadState struct {
	// Messages 消息列表（继承自 Eino）
	Messages []*schema.Message `json:"messages"`

	// Sandbox 沙箱状态
	Sandbox *SandboxState `json:"sandbox,omitempty"`

	// ThreadData 线程数据（路径映射）
	ThreadData *ThreadDataState `json:"thread_data,omitempty"`

	// Title 线程标题（自动生成）
	Title string `json:"title,omitempty"`

	// Artifacts 产物列表（去重合并）
	Artifacts []string `json:"artifacts"`

	// Todos 待办事项列表
	Todos []TodoItem `json:"todos,omitempty"`

	// UploadedFiles 已上传文件列表
	UploadedFiles []UploadedFile `json:"uploaded_files,omitempty"`

	// ViewedImages 已查看图像（路径 -> 图像数据）
	ViewedImages map[string]ViewedImageData `json:"viewed_images"`

	// ClarificationPending 为 true 表示 ask_clarification 已触发，上层应暂停自动工具链或等待用户补充。
	ClarificationPending bool `json:"clarification_pending,omitempty"`
	// LastClarificationMessage 最近一次格式化的澄清文案（面向用户/模型）。
	LastClarificationMessage string `json:"last_clarification_message,omitempty"`

	// LastToolError 最近一次工具失败（OnEnd 正常返回但 next 报错、或 OnError 路径写入）。
	LastToolError *ToolInvocationError `json:"last_tool_error,omitempty"`

	// PendingToolErrorForModel 工具 OnError 时由 DeerFlow Eino 回调写入；下一轮 ChatModel OnStart 会注入为一条合成 Tool 消息后再清空（不序列化到持久化载荷）。
	PendingToolErrorForModel string `json:"-"`
}

// NewThreadState 创建新的 ThreadState
func NewThreadState() *ThreadState {
	return &ThreadState{
		Messages:     make([]*schema.Message, 0),
		Artifacts:    make([]string, 0),
		Todos:        make([]TodoItem, 0),
		UploadedFiles: make([]UploadedFile, 0),
		ViewedImages: make(map[string]ViewedImageData),
	}
}
