package bSdkHandler

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
	bSdkUtil "github.com/phalanx-labs/beacon-sso-sdk/utility"
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

// GetCurrentUser 获取当前用户信
//
// @Summary     [玩家] 用户信息
// @Description 通过访问令牌获取当前登录用户的详细信息，包括基础信息、联系方式、验证状态和角色列表
// @Tags        用户接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Success     200  {object}  xBase.BaseResponse{data=pb.User}  "获取成功"
// @Failure     400  {object}  xBase.BaseResponse               "请求参数错误"
// @Failure     401  {object}  xBase.BaseResponse               "未授权或令牌失效"
// @Router      /sso/user/userinfo [GET]
func (h *UserHandler) GetCurrentUser(ctx *gin.Context) {
	h.log.Info(ctx, "GetCurrentUser - 获取当前用户信息")

	accessToken, xErr := bSdkUtil.GetAccessToken(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	userinfo, err := h.service.userLogic.GetCurrentUser(ctx, accessToken)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, xError.ErrMessage(err.Error()), false, err))
		return
	}

	xResult.SuccessHasData(ctx, "获取用户信息成功", userinfo)
}

// GetUserByID 根据 ID 获取用户信息
//
// 该方法允许已认证的应用查询指定用户的完整信息。
// 主要用于接入 App 需要获取其他用户信息的场景。
//
// @Summary     [用户] 根据ID获取用户
// @Description 根据用户ID获取指定用户的详细信息，包括基础信息、联系方式、验证状态和角色列表
// @Tags        用户接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       user_id        query   string  true  "用户ID（雪花ID格式）"
// @Success     200  {object}  xBase.BaseResponse{data=pb.User}  "获取成功"
// @Failure     400  {object}  xBase.BaseResponse               "请求参数错误"
// @Failure     401  {object}  xBase.BaseResponse               "未授权或令牌失效"
// @Router      /sso/user/by-id [GET]
func (h *UserHandler) GetUserByID(ctx *gin.Context) {
	h.log.Info(ctx, "GetUserByID - 根据ID获取用户信息")

	// 从 Context 获取已验证的 Token
	accessToken, xErr := bSdkUtil.GetAccessToken(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	// 获取 user_id 参数
	userID := ctx.Query("user_id")
	if userID == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ParameterEmpty, "user_id 不能为空", false, nil))
		return
	}

	// 构建请求
	req := &pb.GetUserByIDRequest{
		UserId: userID,
	}

	// 调用业务逻辑
	userinfo, err := h.service.userLogic.GetUserByID(ctx, accessToken, req)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.Unauthorized, xError.ErrMessage(err.Error()), false, err))
		return
	}

	xResult.SuccessHasData(ctx, "获取用户信息成功", userinfo)
}
