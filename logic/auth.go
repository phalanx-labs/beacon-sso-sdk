package bSdkLogic

import (
	"context"
	"log/slog"
	"time"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	bSdkClient "github.com/phalanx-labs/beacon-sso-sdk/client"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
	bSdkModels "github.com/phalanx-labs/beacon-sso-sdk/models"
	bSdkRepo "github.com/phalanx-labs/beacon-sso-sdk/repository"
	bSdkUtil "github.com/phalanx-labs/beacon-sso-sdk/utility"
)

// AuthLogic 认证业务逻辑组件，封装了用户认证流程的核心处理能力。
//
// 该结构体作为业务层的聚合器，整合了 SsoClient 的认证服务接口，
// 用于处理用户注册、登录和密码修改等复杂逻辑。
type AuthLogic struct {
	log       *xLog.LogNamedLogger     // 日志实例
	ssoClient bSdkClient.IAuth         // SsoClient Auth 服务接口
	tokenData *bSdkRepo.OAuthTokenRepo // OAuth Token 数据仓储实例
}

// NewAuth 创建并初始化一个新的 AuthLogic 业务逻辑实例。
//
// 该函数通过从上下文获取 SsoClient 来构建认证业务层。
// 在初始化过程中，它会注入带有命名上下文的日志记录器，
// 从而为认证流程提供完整的日志追踪能力。
//
// 参数:
//   - ctx: 请求上下文，用于获取 SsoClient 实例、数据库和 Redis 实例。
//
// 返回值:
//   - *AuthLogic: 配置完成的认证逻辑层实例指针。
func NewAuth(ctx context.Context) *AuthLogic {
	client := bSdkUtil.GetSsoClient(ctx)
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &AuthLogic{
		log:       xLog.WithName(xLog.NamedLOGC, "AuthLogic"),
		ssoClient: client.Auth,
		tokenData: bSdkRepo.NewOAuthTokenRepo(db, rdb),
	}
}

// RegisterByEmail 通过邮箱注册
//
// 该方法封装了 gRPC 调用，用于通过邮箱验证码完成用户注册。
// 注册成功后自动生成登录 Token。
//
// 参数说明:
//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
//   - req: 注册请求，包含邮箱、验证码、用户名、密码和昵称。
//
// 返回值:
//   - *pb.RegisterByEmailResponse: 包含用户 ID 和登录 Token。
//   - error: 如果注册失败（如验证码错误、邮箱已注册），则返回非 nil 的错误。
func (l *AuthLogic) RegisterByEmail(ctx context.Context, req *pb.RegisterByEmailRequest) (*pb.RegisterByEmailResponse, error) {
	l.log.Info(ctx, "RegisterByEmail - 处理邮箱注册请求")
	return l.ssoClient.RegisterByEmail(ctx, req)
}

// PasswordLogin 密码登录（Resource Owner Password Credentials Grant）
//
// 该方法封装了 gRPC 调用，实现了 OAuth 2.0 Password Grant，
// 允许受信任的第一方客户端直接使用用户名和密码换取 Token。
// 登录成功后会将 Token 缓存到 Redis，以支持后续的 Token 验证和刷新功能。
//
// 参数说明:
//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
//   - req: 密码登录请求，包含用户名、密码和权限范围。
//
// 返回值:
//   - *pb.PasswordLoginResponse: 包含访问令牌、刷新令牌等信息。
//   - error: 如果登录失败（如凭证无效），则返回非 nil 的错误。
func (l *AuthLogic) PasswordLogin(ctx context.Context, req *pb.PasswordLoginRequest) (*pb.PasswordLoginResponse, error) {
	l.log.Info(ctx, "PasswordLogin - 处理密码登录请求")

	// 调用 gRPC 服务
	resp, err := l.ssoClient.PasswordLogin(ctx, req)
	if err != nil {
		return nil, err
	}

	// 缓存令牌到 Redis，失败仅记录警告日志不阻断流程
	if resp.AccessToken != "" {
		expiry := time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
		cacheToken := &bSdkModels.CacheOAuthToken{
			AccessToken:  resp.AccessToken,
			TokenType:    resp.TokenType,
			RefreshToken: resp.GetRefreshToken(),
			Expiry:       expiry.Format(time.RFC3339),
		}
		if storeErr := l.tokenData.Store(ctx, cacheToken); storeErr != nil {
			l.log.Warn(ctx, "PasswordLogin - 缓存令牌失败",
				slog.String("error", storeErr.Error()),
			)
		}
	}

	return resp, nil
}

// ChangePassword 修改用户密码
//
// 该方法封装了 gRPC 调用，允许已认证的应用为用户修改密码。
// 普通模式需要验证旧密码，强制重置模式可跳过旧密码验证。
//
// 参数说明:
//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
//   - req: 修改密码请求，包含用户 ID、旧密码和新密码。
//
// 返回值:
//   - *pb.ChangePasswordResponse: 包含基础响应信息。
//   - error: 如果修改失败（如旧密码错误），则返回非 nil 的错误。
func (l *AuthLogic) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	l.log.Info(ctx, "ChangePassword - 处理修改密码请求")
	return l.ssoClient.ChangePassword(ctx, req)
}

// RevokeToken 注销用户 Token（登出）
//
// 该方法封装了 gRPC 调用，用于注销当前用户的 Access Token，实现用户登出功能。
// 符合 RFC 7009 OAuth 2.0 Token Revocation 规范。
// 注销成功后会尝试清理本地缓存的 Token，失败仅记录警告日志不阻断流程。
//
// 参数说明:
//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
//   - accessToken: 用户访问令牌（Bearer 格式或裸 Token）。
//   - req: 注销请求，可选 token_type_hint 指定注销类型。
//
// 返回值:
//   - *pb.RevokeTokenResponse: 注销结果。
//   - error: 注销失败时返回错误。
func (l *AuthLogic) RevokeToken(ctx context.Context, accessToken string, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	l.log.Info(ctx, "RevokeToken - 处理注销令牌请求")

	// 调用 gRPC 服务
	resp, err := l.ssoClient.RevokeToken(ctx, accessToken, req)
	if err != nil {
		return nil, err
	}

	// 尝试清理本地缓存的 Token，失败仅记录警告日志不阻断流程
	if delErr := l.tokenData.Delete(ctx, accessToken); delErr != nil {
		l.log.Warn(ctx, "RevokeToken - 清理令牌缓存失败",
			slog.String("error", delErr.Error()),
		)
	}

	return resp, nil
}
