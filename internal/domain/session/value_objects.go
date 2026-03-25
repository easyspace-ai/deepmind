package session

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ThreadData 值对象：会话的虚拟文件系统状态
type ThreadData struct {
	userCode   string
	sessionID  string
	workspace  string
	uploads    string
	outputs    string
	metadata   map[string]interface{}
}

// NewThreadData 创建新的 ThreadData
func NewThreadData(userCode string, sessionID string) *ThreadData {
	return &ThreadData{
		userCode:  userCode,
		sessionID: sessionID,
		workspace: fmt.Sprintf("/mnt/user-data/workspace/%s", sessionID),
		uploads:   fmt.Sprintf("/mnt/user-data/uploads/%s", sessionID),
		outputs:   fmt.Sprintf("/mnt/user-data/outputs/%s", sessionID),
		metadata:  make(map[string]interface{}),
	}
}

// GetWorkspacePath 返回工作区路径
func (td *ThreadData) GetWorkspacePath() string {
	return td.workspace
}

// GetUploadsPath 返回上传文件路径
func (td *ThreadData) GetUploadsPath() string {
	return td.uploads
}

// GetOutputsPath 返回输出文件路径
func (td *ThreadData) GetOutputsPath() string {
	return td.outputs
}

// GetMetadata 返回元数据
func (td *ThreadData) GetMetadata() map[string]interface{} {
	return td.metadata
}

// SetMetadata 设置元数据
func (td *ThreadData) SetMetadata(key string, value interface{}) {
	td.metadata[key] = value
}

// PendingToolCall 值对象：工具调用状态
type PendingToolCall struct {
	ID           string
	Name         string
	Arguments    map[string]interface{}
	Result       interface{}
	Error        bool
	RegisteredAt time.Time
	ResolvedAt   time.Time
}

// NewPendingToolCall 创建新的 PendingToolCall
func NewPendingToolCall(name string, arguments map[string]interface{}) *PendingToolCall {
	return &PendingToolCall{
		ID:           uuid.New().String(),
		Name:         name,
		Arguments:    arguments,
		RegisteredAt: time.Now(),
	}
}

// IsResolved 检查工具调用是否已解决
func (ptc *PendingToolCall) IsResolved() bool {
	return !ptc.ResolvedAt.IsZero()
}

// Resolve 解决工具调用
func (ptc *PendingToolCall) Resolve(result interface{}, isError bool) {
	ptc.Result = result
	ptc.Error = isError
	ptc.ResolvedAt = time.Now()
}
