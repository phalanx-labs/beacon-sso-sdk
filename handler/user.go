package bSdkHandler

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xHttp "github.com/bamboo-services/bamboo-base-go/defined/http"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户相关请求处理器
type UserHandler handler

// NewUserHandler 创建并初始化一个 UserHandler 实例
func NewUserHandler(ctx context.Context) *UserHandler {
	newHandler := &UserHandler{
		log: xLog.WithName(xLog.NamedCONT, "UserHandler"),
	}
	(*handler)(newHandler).registerService(ctx)
	return newHandler
}

// GetCurrentUser 获取当前用户信息
//
// 请求方法: GET
// 请求路径: /user/userinfo
// 请求头:
//   - Authorization: Bearer {access_token}
//
// 响应:
//   - 200: 返回用户信息
//   - 400: 参数错误
//   - 401: 未授权或令牌失效
func (h *UserHandler) GetCurrentUser(ctx *gin.Context) {
	h.log.Info(ctx, "GetCurrentUser - 获取当前用户信息")

	accessToken := xHttp.GetToken(ctx, xHttp.HeaderAuthorization)
	if accessToken == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "需要访问令牌参数", false, nil))
		return
	}

	userinfo, err := h.service.userLogic.GetCurrentUser(ctx, accessToken)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, xError.ErrMessage(err.Error()), false, err))
		return
	}

	xResult.SuccessHasData(ctx, "获取用户信息成功", userinfo)
}
