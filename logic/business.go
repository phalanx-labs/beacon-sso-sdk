package bSdkLogic

import (
	"encoding/json"
	"fmt"
	"net/http"

	xEnv "github.com/bamboo-services/bamboo-base-go/env"
	xError "github.com/bamboo-services/bamboo-base-go/error"
	xLog "github.com/bamboo-services/bamboo-base-go/log"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	bSdkConst "github.com/phalanx/beacon-sso-sdk/constant"
	bSdkModels "github.com/phalanx/beacon-sso-sdk/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// BusinessLogic 提供与业务相关的独立 OAuth 能力。
type BusinessLogic struct {
	db  *gorm.DB
	rdb *redis.Client
	log *xLog.LogNamedLogger
}

// NewBusiness 创建并初始化 BusinessLogic。
func NewBusiness(db *gorm.DB, rdb *redis.Client) *BusinessLogic {
	return &BusinessLogic{
		db:  db,
		rdb: rdb,
		log: xLog.WithName(xLog.NamedLOGC),
	}
}

// Userinfo 通过 Access Token 获取并解析 SSO 用户信息
//
// 该方法接收一个有效的 Access Token，向配置的 SSO Userinfo 端点发起请求，
// 以获取当前授权用户的详细信息（如 Sub、昵称、邮箱等）。同时，它会将
// 响应的原始 JSON 数据映射到结构化对象及 Raw 字段中，以兼容标准字段及
// 扩展字段。
//
// 参数说明:
//   - ctx: Gin 上下文对象，用于传递请求上下文及日志追踪。
//   - accessToken: 访问令牌，用于 Bearer 认证。
//
// 返回值:
//   - *bSdkModels.OAuthUserinfo: 解析后的用户信息对象，包含标准字段和原始数据。
//   - *xError.Error: 操作过程中发生的错误，如令牌为空、网络请求失败或解析错误。
func (l *BusinessLogic) Userinfo(ctx *gin.Context, accessToken string) (*bSdkModels.OAuthUserinfo, *xError.Error) {
	l.log.Info(ctx, "BusinessLogic|Userinfo - 获取用户信息")

	if accessToken == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	userinfoURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointUserinfoURI, "")
	if userinfoURI == "" {
		return nil, xError.NewError(ctx, xError.OperationFailed, "用户信息端点为空", false, nil)
	}

	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetAuthToken(accessToken).
		Get(userinfoURI)
	if err != nil {
		return nil, xError.NewError(ctx, xError.OperationFailed, "请求用户信息失败", false, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, xError.NewError(
			ctx,
			xError.Unauthorized,
			xError.ErrMessage(fmt.Sprintf("获取用户信息失败，状态码: %d", resp.StatusCode())),
			false,
			nil,
		)
	}

	raw := make(map[string]any)
	if err = json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, xError.NewError(ctx, xError.OperationFailed, "解析用户信息失败", false, err)
	}

	userinfo := &bSdkModels.OAuthUserinfo{
		Raw: raw,
	}
	if value, ok := raw["sub"].(string); ok {
		userinfo.Sub = value
	}
	if value, ok := raw["nickname"].(string); ok {
		userinfo.Nickname = value
	}
	if value, ok := raw["preferred_username"].(string); ok {
		userinfo.PreferredUsername = value
	}
	if value, ok := raw["email"].(string); ok {
		userinfo.Email = value
	}
	if value, ok := raw["phone"].(string); ok {
		userinfo.Phone = value
	}

	return userinfo, nil
}
