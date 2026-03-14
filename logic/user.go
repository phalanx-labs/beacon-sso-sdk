package bSdkLogic

import (
	"context"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	bSdkClient "github.com/phalanx-labs/beacon-sso-sdk/client"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
	bSdkUtil "github.com/phalanx-labs/beacon-sso-sdk/utility"
)

// UserLogic 用户业务逻辑组件，封装当前用户信息获取流程。
type UserLogic struct {
	log       *xLog.LogNamedLogger // 日志实例
	ssoClient bSdkClient.IUser     // SsoClient User 服务接口
}

// NewUser 创建并初始化一个新的 UserLogic 业务逻辑实例。
//
// 参数:
//   - ctx: 请求上下文，用于获取 SsoClient 实例。
//
// 返回值:
//   - *UserLogic: 配置完成的用户逻辑层实例指针。
func NewUser(ctx context.Context) *UserLogic {
	client := bSdkUtil.GetSsoClient(ctx)
	return &UserLogic{
		log:       xLog.WithName(xLog.NamedLOGC, "UserLogic"),
		ssoClient: client.User,
	}
}

// GetCurrentUser 获取当前登录用户信息
//
// 参数说明:
//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
//   - accessToken: 用户访问令牌（Bearer 格式或裸 Token）。
//
// 返回值:
//   - *pb.GetCurrentUserResponse: 用户信息响应。
//   - error: 获取失败时返回错误。
func (l *UserLogic) GetCurrentUser(ctx context.Context, accessToken string) (*pb.GetCurrentUserResponse, error) {
	l.log.Info(ctx, "GetCurrentUser - 获取当前用户信息")
	return l.ssoClient.GetCurrentUser(ctx, accessToken)
}
