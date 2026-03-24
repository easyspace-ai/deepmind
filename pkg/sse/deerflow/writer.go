package deerflow

import (
	"net/http"

	"github.com/weibaohui/nanobot-go/pkg/sse"
)

// Writer DeerFlow 专用 SSE Writer
// 封装 DeerFlow 特定的三种事件：
//   - messages-tuple: 消息增量更新
//   - values: 完整状态快照
//   - end: 流结束
type Writer struct {
	base *sse.Writer
}

// NewWriter 创建 DeerFlow SSE Writer
func NewWriter(id string, w http.ResponseWriter) *Writer {
	base := sse.NewWriter(id, w)
	if base == nil {
		return nil
	}
	return &Writer{base: base}
}

// WriteMessagesTuple 发送 messages-tuple 事件
// 用于发送 AI 消息的增量更新
func (w *Writer) WriteMessagesTuple(msg interface{}) error {
	return w.base.WriteEventJSON("", "messages-tuple", []interface{}{msg})
}

// WriteValues 发送 values 事件
// 用于发送完整的 thread 状态快照
func (w *Writer) WriteValues(values map[string]interface{}) error {
	return w.base.WriteEventJSON("", "values", values)
}

// WriteEnd 发送 end 事件
// 用于标记流结束，包含 usage 信息
func (w *Writer) WriteEnd(usage map[string]interface{}) error {
	data := map[string]interface{}{
		"usage": usage,
	}
	return w.base.WriteEventJSON("", "end", data)
}

// Close 关闭 Writer
func (w *Writer) Close() {
	w.base.Close()
}

// IsClosed 检查是否已关闭
func (w *Writer) IsClosed() bool {
	return w.base.IsClosed()
}

// SetDone 设置完成回调
func (w *Writer) SetDone(f func(http.ResponseWriter) error) {
	w.base.SetDone(f)
}
