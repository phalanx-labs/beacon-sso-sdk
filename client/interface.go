package client

import (
	"context"

	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
	"github.com/phalanx-labs/beacon-sso-sdk/client/service"
)

// IPublic 定义了公共服务操作的标准接口
type IPublic interface {
	// SendRegisterEmailCode 发送注册邮箱验证码
	//
	// 该接口用于向指定邮箱发送注册验证码，验证码有效期为环境变量 EMAIL_VERIFY_CODE_EXPIRE 配置的时间（默认 15 分钟）。
	// 同一邮箱在 1 分钟内只能发送一次验证码。
	//
	// 参数说明:
	//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
	//   - req: 发送验证码请求，包含目标邮箱地址。
	//
	// 返回值:
	//   - *pb.SendRegisterEmailCodeResponse: 包含基础响应信息。
	//   - error: 如果发送失败（如邮箱格式无效、发送服务不可用），则返回非 nil 的错误。
	SendRegisterEmailCode(ctx context.Context, req *pb.SendRegisterEmailCodeRequest) (*pb.SendRegisterEmailCodeResponse, error)
}

// IAuth 定义了认证服务操作的标准接口
//
// 该服务的所有方法都需要在 metadata 中提供有效的 App 凭证：
//   - app-access-id: App 的 Access ID
//   - app-secret-key: App 的 Secret Key
type IAuth interface {
	// RegisterByEmail 通过邮箱注册
	//
	// 该接口用于通过邮箱验证码完成用户注册，注册成功后自动生成登录 Token。
	// 注册流程：
	// 1. 验证邮箱验证码
	// 2. 检查邮箱是否已注册
	// 3. 验证密码强度
	// 4. 创建用户账号
	// 5. 绑定邮箱并标记为已验证
	// 6. 生成登录 Token
	//
	// 参数说明:
	//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
	//   - req: 注册请求，包含邮箱、验证码、用户名、密码和昵称。
	//
	// 返回值:
	//   - *pb.RegisterByEmailResponse: 包含用户 ID 和登录 Token。
	//   - error: 如果注册失败（如验证码错误、邮箱已注册），则返回非 nil 的错误。
	RegisterByEmail(ctx context.Context, req *pb.RegisterByEmailRequest) (*pb.RegisterByEmailResponse, error)

	// PasswordLogin 密码登录（Resource Owner Password Credentials Grant）
	//
	// 该方法实现了 OAuth 2.0 Password Grant，允许受信任的第一方客户端
	// 直接使用用户名和密码换取 Token。
	//
	// 安全特性：
	// - 仅限第一方应用使用（App.FirstParty = enabled）
	// - 支持用户名/邮箱/手机号三种登录方式（自动识别）
	//
	// 参数说明:
	//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
	//   - req: 密码登录请求，包含用户名、密码和权限范围。
	//
	// 返回值:
	//   - *pb.PasswordLoginResponse: 包含访问令牌、刷新令牌等信息。
	//   - error: 如果登录失败（如凭证无效），则返回非 nil 的错误。
	PasswordLogin(ctx context.Context, req *pb.PasswordLoginRequest) (*pb.PasswordLoginResponse, error)

	// ChangePassword 修改用户密码
	//
	// 该方法允许已认证的应用为用户修改密码。
	// 普通模式需要验证旧密码，强制重置模式可跳过旧密码验证。
	//
	// 模式说明：
	//   - 普通模式（NeedResetPassword=false）：必须提供 old_password 进行验证
	//   - 强制重置模式（NeedResetPassword=true）：可省略 old_password，直接设置新密码
	//
	// 参数说明:
	//   - ctx: 上下文，用于控制请求的生命周期和超时控制。
	//   - req: 修改密码请求，包含用户 ID、旧密码和新密码。
	//
	// 返回值:
	//   - *pb.ChangePasswordResponse: 包含基础响应信息。
	//   - error: 如果修改失败（如旧密码错误），则返回非 nil 的错误。
	ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error)
}

// IMerchant 定义了商户服务操作的标准接口
//
// 该服务的所有方法都需要在 metadata 中提供有效的 App 凭证：
//   - app-access-id: App 的 Access ID
//   - app-secret-key: App 的 Secret Key
type IMerchant interface {
	// GetMerchantTags 获取当前应用所属商户的所有标签
	//
	// 返回应用所属商户下的所有标签列表，可用于客户端展示或标签匹配。
	GetMerchantTags(ctx context.Context, req *service.GetMerchantTagsRequest) (*service.GetMerchantTagsResponse, error)

	// GetUserTags 获取指定用户在当前商户的所有标签
	//
	// 根据用户 ID 获取该用户在当前应用所属商户下的所有标签。
	GetUserTags(ctx context.Context, req *service.GetUserTagsRequest) (*service.GetUserTagsResponse, error)

	// CheckUserHasTag 检查用户是否有指定标签
	//
	// 通过标签代码（code）快速检查用户是否拥有该标签。
	CheckUserHasTag(ctx context.Context, req *service.CheckUserHasTagRequest) (*service.CheckUserHasTagResponse, error)

	// GetRecentAnnouncements 获取最近公告列表
	//
	// 获取当前应用所属商户的最近公告（最多 10 条）。
	// 返回结果包含 MD5 和 SHA256 哈希值，客户端可用于判断是否需要重新展示。
	GetRecentAnnouncements(ctx context.Context, req *service.GetRecentAnnouncementsRequest) (*service.GetRecentAnnouncementsResponse, error)

	// GetAnnouncement 获取单个公告详情
	//
	// 根据公告 ID 获取详细信息。
	GetAnnouncement(ctx context.Context, req *service.GetAnnouncementRequest) (*service.GetAnnouncementResponse, error)
}
