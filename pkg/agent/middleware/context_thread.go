package middleware

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
)

type deerflowThreadIDKey struct{}

// deerflowToolInputKey 在单次 Tool 回调周期内传递 *tool.CallbackInput（供 OnEnd 解析参数）。
type deerflowToolInputKey struct{}

// WithThreadID 将 DeerFlow 线程 ID 写入 context，供中间件链 Eino 回调使用。
func WithThreadID(ctx context.Context, threadID string) context.Context {
	if threadID == "" {
		return ctx
	}
	return context.WithValue(ctx, deerflowThreadIDKey{}, threadID)
}

// ThreadIDFromContext 读取 WithThreadID 注入的线程 ID；未设置时返回空串。
func ThreadIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(deerflowThreadIDKey{}).(string)
	return v
}

// WithToolCallbackInput 供 Eino Tool OnStart 注入，便于 OnEnd 中读取 JSON 参数。
func WithToolCallbackInput(ctx context.Context, in *tool.CallbackInput) context.Context {
	return context.WithValue(ctx, deerflowToolInputKey{}, in)
}

func toolCallbackInputFromContext(ctx context.Context) *tool.CallbackInput {
	v, _ := ctx.Value(deerflowToolInputKey{}).(*tool.CallbackInput)
	return v
}
