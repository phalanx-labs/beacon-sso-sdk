package bSdkRoute

import (
	"context"
)

// Route 路由注册器
//
// 该结构体封装了上下文，用于在路由注册时传递给 Handler 初始化。
type Route struct {
	ctx context.Context // 上下文，用于控制取消和超时
}

// NewRoute 创建并返回一个新的 Route 实例
//
// 参数:
//   - ctx: 请求上下文，用于初始化 Handler。
//
// 返回值:
//   - *Route: 配置完成的路由注册器实例指针。
func NewRoute(ctx context.Context) *Route {
	return &Route{
		ctx: ctx,
	}
}
