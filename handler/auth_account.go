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
// @Summary     [公开] 邮箱注册
// @Description 通过邮箱验证码完成用户注册，注册成功后自动生成登录 Token
// @Tags        账户接口
// @Accept      json
// @Produce     json
// @Param       request  body  pb.RegisterByEmailRequest  true  "邮箱注册请求"
// @Success     200  {object}  xBase.BaseResponse{data=pb.RegisterByEmailResponse}  "注册成功"
// @Failure     400  {object}  xBase.BaseResponse  "请求参数错误"
// @Failure     500  {object}  xBase.BaseResponse  "服务器内部错误"
// @Router      /account/register/email [POST]
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
		_ = ctx.Error(xError.NewError(ctx, xError.OperationFailed, xError.ErrMessage(err.Error()), false, err))
		return
	}

	xResult.SuccessHasData(ctx, "注册成功", resp)
}

// PasswordLogin 密码登录（OAuth 2.0 Password Grant）
//
// 该接口实现了 OAuth 2.0 Password Grant，允许受信任的第一方客户端
// 直接使用用户名和密码换取 Token。
//
// @Summary     [公开] 密码登录
// @Description 使用用户名和密码进行登录（OAuth 2.0 Password Grant），返回访问令牌和刷新令牌
// @Tags        账户接口
// @Accept      json
// @Produce     json
// @Param       request  body  pb.PasswordLoginRequest  true  "密码登录请求"
// @Success     200  {object}  xBase.BaseResponse{data=pb.PasswordLoginResponse}  "登录成功"
// @Failure     400  {object}  xBase.BaseResponse  "请求参数错误"
// @Failure     401  {object}  xBase.BaseResponse  "凭证无效"
// @Failure     500  {object}  xBase.BaseResponse  "服务器内部错误"
// @Router      /account/login/password [POST]
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
		_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, xError.ErrMessage(err.Error()), false, err))
		return
	}

	xResult.SuccessHasData(ctx, "登录成功", resp)
}

// ChangePassword 修改密码
//
// @Summary     [用户] 修改密码
//
// 该接口允许已认证的应用为用户修改密码。
// 普通模式需要验证旧密码，强制重置模式可跳过旧密码验证。
//
// @Description 修改用户密码，普通模式需要验证旧密码，强制重置模式可跳过旧密码验证
// @Tags        账户接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header  string                    true  "Bearer Access Token"
// @Param       request        body    pb.ChangePasswordRequest  true  "修改密码请求"
// @Success     200  {object}  xBase.BaseResponse{data=pb.ChangePasswordResponse}  "修改成功"
// @Failure     400  {object}  xBase.BaseResponse  "请求参数错误"
// @Failure     401  {object}  xBase.BaseResponse  "未授权或旧密码错误"
// @Failure     500  {object}  xBase.BaseResponse  "服务器内部错误"
// @Router      /account/password/change [POST]
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
		_ = ctx.Error(xError.NewError(ctx, xError.OperationFailed, xError.ErrMessage(err.Error()), false, err))
		return
	}

	xResult.SuccessHasData(ctx, "修改密码成功", resp)
}

// RevokeToken 注销令牌（登出）
//
// 该方法用于注销当前用户的 Access Token，实现用户登出功能。
// 符合 RFC 7009 OAuth 2.0 Token Revocation 规范。
//
// @Summary     [用户] 注销令牌
// @Description 注销当前用户的访问令牌或刷新令牌，实现用户登出功能（符合 RFC 7009 OAuth 2.0 Token Revocation）
// @Tags        账户接口
// @Accept      json
// @Produce     json
// @Param       Authorization     header  string  true  "Bearer Access Token"
// @Param       token_type_hint   query   string  false  "令牌类型提示：access_token 或 refresh_token"  Enums(access_token, refresh_token)
// @Success     200  {object}  xBase.BaseResponse{data=pb.RevokeTokenResponse}  "注销成功"
// @Failure     400  {object}  xBase.BaseResponse  "请求参数错误"
// @Failure     401  {object}  xBase.BaseResponse  "未授权或令牌失效"
// @Failure     500  {object}  xBase.BaseResponse  "服务器内部错误"
// @Router      /account/token/revoke [POST]
func (h *AccountHandler) RevokeToken(ctx *gin.Context) {
	h.log.Info(ctx, "RevokeToken - 处理注销令牌请求")

	// 获取 Authorization Token
	accessToken := ctx.GetHeader("Authorization")
	if accessToken == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "需要访问令牌参数", false, nil))
		return
	}

	// 获取可选的 token_type_hint 参数
	tokenTypeHint := ctx.Query("token_type_hint")

	// 构建请求
	req := &pb.RevokeTokenRequest{}
	if tokenTypeHint != "" {
		req.TokenTypeHint = &tokenTypeHint
	}

	// 调用业务逻辑
	resp, err := h.service.authLogic.RevokeToken(ctx, accessToken, req)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, xError.ErrMessage(err.Error()), false, err))
		return
	}

	xResult.SuccessHasData(ctx, "注销令牌成功", resp)
}
