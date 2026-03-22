package models

// DeerFlowModelEntry 对齐 DeerFlow 配置里单条模型描述（use / thinking / vision 等开关的语义锚点）。
// 实际运行时创建 ToolCallingChatModel 请使用 pkg/agent/provider 中的 NewChatModelAdapter / NewChatModelAdapterV2
//（数据库与会话绑定）；本包不重复实现反射式工厂，避免与 nanobot 统一模型入口分叉。
type DeerFlowModelEntry struct {
	Use      string `json:"use"`
	Thinking bool   `json:"thinking,omitempty"`
	Vision   bool   `json:"vision,omitempty"`
}
