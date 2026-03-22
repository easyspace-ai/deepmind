package websocket

// Config WebSocket 渠道配置
type Config struct {
	Addr        string `json:"addr"`         // 监听地址，如 ":8080"
	Path        string `json:"path"`         // WebSocket 路径，如 "/ws/chat"
	ChannelCode string `json:"channel_code"` // 渠道编码
	ChannelID   uint   `json:"channel_id"`   // 渠道 ID（数据库中的 ID）
	AgentCode   string `json:"agent_code"`   // 绑定的 Agent 编码

	// SandboxRelease 连接关闭且连接上已绑定 SessionID 时调用，释放与 session_key 对应的本地沙箱（DeerFlow 线程目录）。
	SandboxRelease func(sessionKey string)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addr: ":8080",
		Path: "/ws/chat",
	}
}
