package bSdkHandler

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
)

// AccountHandler 账户处理的请求器
//
// 该处理器负责处理用户账户相关的 HTTP 请求，包括邮箱注册、密码登录和修改密码。
type AccountHandler handler

// NewAccountHandler 创建并初始化一个 AccountHandler 实例
//
// 参数:
//   - ctx: 请求上下文，用于初始化业务逻辑层。
//
// 返回值:
//   - *AccountHandler: 配置完成的账户处理器实例指针。
func NewAccountHandler(ctx context.Context) *AccountHandler {
	newHandler := &AccountHandler{
		log: xLog.WithName(xLog.NamedCONT, "AccountHandler"),
	}
	(*handler)(newHandler).registerService(ctx)
	return newHandler
}

// RegisterByEmail 邮箱注册
//
// 该接口用于通过邮箱验证码完成用户注册，注册成功后自动生成登录 Token。
//
// 请求方法: POST
// 请求路径: /account/register/email
// 请求体 (JSON):
//   - email: 邮箱地址（必填）
//   - code: 邮箱验证码（必填）
//   - username: 用户名（必填）
//   - password: 密码（必填）
//   - nickname: 昵称（可选）
//
// 响应:
//   - 200: 注册成功，返回用户 ID 和登录 Token
//   - 400: 参数错误
//   - 500: 服务器内部错误
func (h *AccountHandler) RegisterByEmail(ctx *gin.Context) {
	h.log.Info(ctx, "RegisterByEmail - 处理邮箱注册请求")

	// 绑定请求体
	var req pb.RegisterByEmailRequest
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterError, "请求参数格式错误", false, bindErr))
		return
	}

	// 参数校验
	if req.Email == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "邮箱不能为空", false, nil))
		return
	}
	if req.Code == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "验证码不能为空", false, nil))
		return
	}
	if req.Username == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "用户名不能为空", false, nil))
		return
	}
	if req.Password == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "密码不能为空", false, nil))
		return
	}

	// 调用业务逻辑
	resp, err := h.service.authLogic.RegisterByEmail(ctx, &req)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.OperationFailed, "注册失败", false, err))
		return
	}

	xResult.SuccessHasData(ctx, "注册成功", resp)
}

// PasswordLogin 密码登录（OAuth 2.0 Password Grant）
//
// 该接口实现了 OAuth 2.0 Password Grant，允许受信任的第一方客户端
// 直接使用用户名和密码换取 Token。
//
// 请求方法: POST
// 请求路径: /account/login/password
// 请求体 (JSON):
//   - username: 用户名/邮箱/手机号（必填）
//   - password: 密码（必填）
//   - scope: 权限范围（必填）
//
// 响应:
//   - 200: 登录成功，返回访问令牌、刷新令牌等信息
//   - 400: 参数错误
//   - 401: 凭证无效
//   - 500: 服务器内部错误
func (h *AccountHandler) PasswordLogin(ctx *gin.Context) {
	h.log.Info(ctx, "PasswordLogin - 处理密码登录请求")

	// 绑定请求体
	var req pb.PasswordLoginRequest
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterError, "请求参数格式错误", false, bindErr))
		return
	}

	// 参数校验
	if req.Username == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "用户名不能为空", false, nil))
		return
	}
	if req.Password == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "密码不能为空", false, nil))
		return
	}
	if req.Scope == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "权限范围不能为空", false, nil))
		return
	}

	// 调用业务逻辑
	resp, err := h.service.authLogic.PasswordLogin(ctx, &req)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, "登录失败", false, err))
		return
	}

	xResult.SuccessHasData(ctx, "登录成功", resp)
}

// ChangePassword 修改密码
//
// 该接口允许已认证的应用为用户修改密码。
// 普通模式需要验证旧密码，强制重置模式可跳过旧密码验证。
//
// 请求方法: POST
// 请求路径: /account/password/change
// 请求体 (JSON):
//   - user_id: 用户 ID（必填）
//   - old_password: 旧密码（普通模式必填）
//   - new_password: 新密码（必填）
//
// 响应:
//   - 200: 修改成功
//   - 400: 参数错误
//   - 401: 旧密码错误
//   - 500: 服务器内部错误
func (h *AccountHandler) ChangePassword(ctx *gin.Context) {
	h.log.Info(ctx, "ChangePassword - 处理修改密码请求")

	// 绑定请求体
	var req pb.ChangePasswordRequest
	if bindErr := ctx.ShouldBindJSON(&req); bindErr != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterError, "请求参数格式错误", false, bindErr))
		return
	}

	// 参数校验
	if req.UserId == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "用户 ID 不能为空", false, nil))
		return
	}
	if req.NewPassword == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "新密码不能为空", false, nil))
		return
	}

	// 调用业务逻辑
	resp, err := h.service.authLogic.ChangePassword(ctx, &req)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.OperationFailed, "修改密码失败", false, err))
		return
	}

	xResult.SuccessHasData(ctx, "修改密码成功", resp)
}
