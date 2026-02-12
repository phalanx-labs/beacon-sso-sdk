package bSdkLogic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	xEnv "github.com/bamboo-services/bamboo-base-go/env"
	xError "github.com/bamboo-services/bamboo-base-go/error"
	xLog "github.com/bamboo-services/bamboo-base-go/log"
	xUtil "github.com/bamboo-services/bamboo-base-go/utility"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/utility/ctxutil"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	bSdkConst "github.com/phalanx/beacon-sso-sdk/constant"
	bSdkModels "github.com/phalanx/beacon-sso-sdk/models"
	bSdkRepo "github.com/phalanx/beacon-sso-sdk/repository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// BusinessLogic 提供与业务相关的独立 OAuth 能力。
type BusinessLogic struct {
	db                *gorm.DB
	rdb               *redis.Client
	log               *xLog.LogNamedLogger
	userinfoData      *bSdkRepo.UserinfoRepo
	introspectionData *bSdkRepo.IntrospectionRepo
}

// NewBusiness 创建并初始化 BusinessLogic。
func NewBusiness(ctx context.Context) *BusinessLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &BusinessLogic{
		db:                db,
		rdb:               rdb,
		log:               xLog.WithName(xLog.NamedLOGC, "BusinessLogic"),
		userinfoData:      bSdkRepo.NewUserinfoRepo(db, rdb),
		introspectionData: bSdkRepo.NewIntrospectionRepo(db, rdb),
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
	l.log.Info(ctx, "Userinfo - 获取用户信息")

	if accessToken == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	cacheValue, cacheExists, cacheErr := l.userinfoData.GetCache(ctx, accessToken)
	if cacheErr != nil {
		l.log.Warn(ctx, "BusinessLogic|Userinfo - 读取缓存失败",
			slog.String("error", cacheErr.Error()),
		)
	} else if cacheExists {
		return cacheValue, nil
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

	userinfo := &bSdkModels.OAuthUserinfo{Raw: raw}
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

	if cacheErr = l.userinfoData.StoreCache(ctx, accessToken, userinfo); cacheErr != nil {
		l.log.Warn(ctx, "BusinessLogic|Userinfo - 写入缓存失败",
			slog.String("error", cacheErr.Error()),
		)
	}
	return userinfo, nil
}

// Introspection 调用 OAuth2 Introspection Endpoint 查询令牌状态与有效期。
func (l *BusinessLogic) Introspection(ctx *gin.Context, tokenType string, token string) (*bSdkModels.OAuthIntrospection, *xError.Error) {
	l.log.Info(ctx, "Introspection - 查询令牌有效期")

	if tokenType == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "令牌类型为空", false, nil)
	}
	if token == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	cacheValue, cacheExists, cacheErr := l.introspectionData.GetCache(ctx, tokenType, token)
	if cacheErr != nil {
		l.log.Warn(ctx, "BusinessLogic|Introspection - 读取缓存失败",
			slog.String("error", cacheErr.Error()),
		)
	} else if cacheExists {
		return cacheValue, nil
	}

	introspectionURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointIntrospectionURI, "")
	if introspectionURI == "" {
		return nil, xError.NewError(ctx, xError.OperationFailed, "自省端点为空", false, nil)
	}

	clientID := xEnv.GetEnvString(bSdkConst.EnvSsoClientID, "")
	clientSecret := xEnv.GetEnvString(bSdkConst.EnvSsoClientSecret, "")
	if clientID == "" || clientSecret == "" {
		return nil, xError.NewError(ctx, xError.OperationFailed, "客户端配置缺失", false, nil)
	}

	client := resty.New()
	resp, reqErr := client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"token":           token,
			"token_type_hint": tokenType,
		}).
		Post(introspectionURI)
	if reqErr != nil {
		return nil, xError.NewError(ctx, xError.OperationFailed, "查询令牌状态失败", false, reqErr)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, xError.NewError(
			ctx,
			xError.OperationFailed,
			xError.ErrMessage(fmt.Sprintf("查询令牌状态失败，状态码: %d", resp.StatusCode())),
			false,
			nil,
		)
	}

	raw := make(map[string]any)
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return nil, xError.NewError(ctx, xError.OperationFailed, "解析令牌状态失败", false, err)
	}

	result := &bSdkModels.OAuthIntrospection{Raw: raw}
	if active, ok := raw["active"].(bool); ok {
		result.Active = active
	}
	if value, ok := raw["token_type"].(string); ok {
		result.TokenType = value
	}

	expUnix, ok := xUtil.Parse().Int64(raw["exp"])
	if ok {
		result.Exp = expUnix
		expiry := time.Unix(expUnix, 0)
		result.Expiry = expiry.Format(time.RFC3339)

		expiresIn := int64(time.Until(expiry).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
			result.IsExpired = true
		} else {
			result.IsExpired = false
		}
		result.ExpiresIn = expiresIn
	}

	if cacheErr = l.introspectionData.StoreCache(ctx, tokenType, token, result); cacheErr != nil {
		l.log.Warn(ctx, "BusinessLogic|Introspection - 写入缓存失败",
			slog.String("error", cacheErr.Error()),
		)
	}
	return result, nil
}
