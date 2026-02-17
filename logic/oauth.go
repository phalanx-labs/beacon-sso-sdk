package bSdkLogic

import (
	"context"
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
	bSdkUtil "github.com/phalanx/beacon-sso-sdk/utility"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// OAuthLogic OAuth 业务逻辑组件，封装了身份认证流程的核心处理能力。
//
// 该结构体作为业务层的聚合器，整合了底层数据资源（GORM、Redis）和特定的数据仓储，
// 用于处理诸如令牌颁发、用户信息检索及权限校验等复杂逻辑。
type OAuthLogic struct {
	db        *gorm.DB                 // GORM 数据库实例
	rdb       *redis.Client            // Redis 客户端实例
	log       *xLog.LogNamedLogger     // 日志实例
	data      *bSdkRepo.OAuthRepo      // OAuth 数据仓储实例
	tokenData *bSdkRepo.OAuthTokenRepo // OAuth Token 数据仓储实例
}

// NewOAuth 创建并初始化一个新的 OAuthLogic 业务逻辑实例。
//
// 该函数通过组合 GORM 数据库实例和 Redis 客户端来构建 OAuth 业务层。
// 在初始化过程中，它会注入底层的 OAuth 数据仓储（包含默认 30 分钟的
// 缓存策略）以及带有命名上下文的日志记录器，从而为 OAuth 认证流程
// 提供完整的数据持久化、缓存加速和日志追踪能力。
//
// 参数:
//   - ctx: 请求上下文，用于获取数据库和 Redis 实例。
//
// 返回值:
//   - *OAuthLogic: 配置完成的 OAuth 逻辑层实例指针。
func NewOAuth(ctx context.Context) *OAuthLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &OAuthLogic{
		db:        db,
		rdb:       rdb,
		log:       xLog.WithName(xLog.NamedLOGC, "OAuthLogic"),
		data:      bSdkRepo.NewOAuthRepo(db, rdb),
		tokenData: bSdkRepo.NewOAuthTokenRepo(db, rdb),
	}
}

// Create 初始化并存储 OAuth 2.0 认证流程所需的 State 和 PKCE Verifier
//
// 该方法生成一个随机的 State 字符串和一个符合 OAuth 2.0 PKCE 规范的 Code Verifier，
// 并将这对参数存储在缓存（通常是 Redis）中。State 参数用于在回调请求中验证请求的
// 一致性以防止 CSRF 攻击，而 Verifier 则用于后续换取 Token 时的安全校验。
//
// 参数说明:
//   - ctx: Gin 请求上下文，用于传递请求范围的数据、控制超时及日志记录。
//
// 返回值:
//   - *bSdkModels.CacheOAuth: 包含生成的 State 和 Verifier 的缓存对象。
//   - *xError.Error: 存储操作失败时返回错误信息（例如 Redis 连接问题）。
//
// 注意: 此方法仅负责数据的创建与存储，不直接处理 HTTP 请求或响应。
func (l *OAuthLogic) Create(ctx *gin.Context) (*bSdkModels.CacheOAuth, *xError.Error) {
	l.log.Info(ctx, "Create - 创建 STATE 和 PCKE 码")

	generateVerifier := oauth2.GenerateVerifier()
	state := xUtil.GenerateRandomUpperString(32)

	if err := l.data.Store(ctx, state, generateVerifier); err != nil {
		return nil, err
	}

	return &bSdkModels.CacheOAuth{
		State:    state,
		Verifier: generateVerifier,
	}, nil
}

// BuildURL 构建 OAuth 2.0 授权跳转 URL
//
// 该方法根据传入的 OAuth 缓存对象（包含 State 和 PKCE Verifier），
// 结合系统配置生成完整的授权码请求 URL。它利用 S256 (SHA-256) 方法
// 生成 Code Challenge，以满足 PKCE (Proof Key for Code Exchange) 安全规范。
//
// 参数说明:
//   - ctx: Gin 请求上下文，用于日志记录和获取配置。
//   - oAuth: 包含 State 和 Verifier 信息的缓存对象。
//
// 返回值:
//   - string: 生成的授权跳转 URL。
//   - *xError.Error: 构建失败时（例如获取配置失败）返回错误，否则为 nil。
func (l *OAuthLogic) BuildURL(ctx *gin.Context, oAuth *bSdkModels.CacheOAuth) (string, *xError.Error) {
	l.log.Info(ctx, "BuildURL - 构建跳转地址")

	var authCodeConfig = []oauth2.AuthCodeOption{
		oauth2.S256ChallengeOption(oAuth.Verifier),
	}
	authURL := bSdkUtil.GetOAuthConfig(ctx).AuthCodeURL(oAuth.State, authCodeConfig...)
	return authURL, nil
}

// Verify 验证 OAuth State 并获取 PKCE Verifier
//
// 该方法负责 OAuth 2.0 流程中的安全校验环节。它根据传入的 state 参数从缓存中检索对应的
// PKCE Verifier。为了防止重放攻击，验证成功后会尝试清理缓存中的 state 数据。
//
// 参数说明:
//   - ctx: Gin 请求上下文，用于传递请求范围的数据、控制超时及日志记录。
//   - state: 客户端在请求授权时生成的随机状态码，用于验证请求的完整性。
//
// 返回值:
//   - *bSdkModels.CacheOAuth: 包含验证通过的 State 和 Verifier 信息的缓存对象。
//   - *xError.Error: 当 state 为空、缓存中不存在对应记录或数据读取失败时返回错误。
//
// 注意: 验证通过后，该方法会尝试从缓存中删除该 state 记录。如果删除操作失败，
// 仅记录警告日志而不阻断验证流程。
func (l *OAuthLogic) Verify(ctx *gin.Context, state string) (*bSdkModels.CacheOAuth, *xError.Error) {
	l.log.Info(ctx, "Verify - 校验 STATE")

	if state == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "状态为空", false, nil)
	}

	cacheValue, err := l.data.Get(ctx, state)
	if err != nil {
		return nil, err
	}
	if cacheValue.Verifier == "" {
		return nil, xError.NewError(ctx, xError.NotExist, "验证器不存在", false, nil)
	}

	if delErr := l.data.Delete(ctx, state); delErr != nil {
		l.log.Warn(ctx, "OAuthLogic|Verify - 删除缓存失败",
			slog.String("state", state),
			slog.String("error", delErr.Error()),
		)
	}

	return cacheValue, nil
}

// Exchange 使用授权码和 PKCE 验证器换取访问令牌
//
// 该方法是 OAuth 2.0 授权码流程的最后一步，负责使用从回调地址中获取的授权码（code）
// 和在 Create 阶段生成的 PKCE 验证器（verifier）向认证服务器请求访问令牌。
//
// 参数说明:
//   - ctx: Gin 请求上下文，用于传递请求范围的数据、控制超时及日志记录。
//   - code: OAuth 回调返回的授权码。
//   - verifier: PKCE 代码验证器，必须与 Create 阶段生成的值一致。
//   - codeType: 授权码类型，默认为 "authorization_code"，如果是刷新令牌则为 "refresh_token"。
//
// 返回值:
//   - *oauth2.Token: 包含访问令牌、刷新令牌及过期时间信息的对象。
//   - *xError.Error: 如果授权码无效、验证器不匹配或网络请求失败，则返回具体的错误信息。
func (l *OAuthLogic) Exchange(ctx *gin.Context, code string, verifier string) (*oauth2.Token, *xError.Error) {
	l.log.Info(ctx, "Exchange - 换取令牌")

	var authCodeConfig = []oauth2.AuthCodeOption{
		oauth2.VerifierOption(verifier),
	}
	getToken, oAuthErr := bSdkUtil.GetOAuthConfig(ctx).Exchange(ctx, code, authCodeConfig...)
	if oAuthErr != nil {
		return nil, xError.NewError(ctx, xError.Unauthorized, "未登录", false, oAuthErr)
	}

	// 缓存令牌到 Redis，失败仅记录警告日志不阻断流程
	cacheToken := &bSdkModels.CacheOAuthToken{
		AccessToken:  getToken.AccessToken,
		TokenType:    getToken.TokenType,
		RefreshToken: getToken.RefreshToken,
		Expiry:       getToken.Expiry.Format(time.RFC3339),
	}
	if storeErr := l.tokenData.Store(ctx, cacheToken); storeErr != nil {
		l.log.Warn(ctx, "Exchange - 缓存令牌失败",
			slog.String("error", storeErr.Error()),
		)
	}

	return getToken, nil
}

func (l *OAuthLogic) TokenSource(ctx *gin.Context, cacheToken *bSdkModels.CacheOAuthToken, rt string) (*oauth2.Token, *xError.Error) {
	l.log.Info(ctx, "TokenSource - 刷新令牌")

	// 校验 RT 是否一致
	if cacheToken.RefreshToken != rt {
		return nil, l.tokenData.Delete(ctx, cacheToken.AccessToken)
	}

	// 构造 oauth2.Token
	parseTime, timeErr := time.Parse(time.RFC3339, cacheToken.Expiry)
	if timeErr != nil {
		return nil, xError.NewError(ctx, xError.OperationFailed, "解析令牌过期时间失败", false, timeErr)
	}
	oldToke := &oauth2.Token{
		AccessToken:  cacheToken.AccessToken,
		TokenType:    cacheToken.TokenType,
		RefreshToken: cacheToken.RefreshToken,
		Expiry:       parseTime,
	}

	// 尝试刷新
	tokenSource, err := bSdkUtil.GetOAuthConfig(ctx).TokenSource(ctx, oldToke).Token()
	if err != nil {
		return nil, xError.NewError(ctx, xError.Unauthorized, "未登录", false, err)
	}
	newToken := &bSdkModels.CacheOAuthToken{
		AccessToken:  tokenSource.AccessToken,
		TokenType:    tokenSource.TokenType,
		RefreshToken: tokenSource.RefreshToken,
		Expiry:       tokenSource.Expiry.Format(time.RFC3339),
	}
	if storeErr := l.tokenData.Store(ctx, newToken); storeErr != nil {
		l.log.Warn(ctx, "Exchange - 缓存令牌失败",
			slog.String("error", storeErr.Error()),
		)
	}
	return tokenSource, nil
}

// GetToken 根据 AccessToken 从缓存中获取令牌信息
//
// 该方法供中间件调用，用于通过 Bearer Token 直接从 Redis 缓存中验证和获取令牌信息，
// 避免每次请求都需要重新访问 OAuth2 平台。
//
// 参数说明:
//   - ctx: Gin 请求上下文。
//   - accessToken: 客户端传入的访问令牌。
//
// 返回值:
//   - *bSdkModels.CacheOAuthToken: 缓存中的令牌信息。
//   - *xError.Error: 查询失败时返回错误信息。
func (l *OAuthLogic) GetToken(ctx *gin.Context, accessToken string) (*bSdkModels.CacheOAuthToken, *xError.Error) {
	l.log.Info(ctx, "GetToken - 获取缓存令牌")

	if accessToken == "" {
		return nil, xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	return l.tokenData.Get(ctx, accessToken)
}

// VerifyExpiry 校验令牌是否已过期
//
// 该方法根据提供的访问令牌从缓存中查询令牌详情，并将其过期时间与当前时间进行比对。
//
// 参数说明:
//   - ctx: Gin 请求上下文。
//   - accessToken: 待校验的访问令牌字符串。
//
// 返回值:
//   - bool: 如果令牌已过期返回 true，否则返回 false。
//   - *xError.Error: 查询令牌信息失败时返回错误，成功时为 nil。
func (l *OAuthLogic) VerifyExpiry(ctx *gin.Context, accessToken string) (bool, *xError.Error) {
	l.log.Info(ctx, "VerifyExpiry - 验证令牌过期")

	getToken, xErr := l.tokenData.Get(ctx, accessToken)
	if xErr != nil {
		return false, xErr
	}
	parseTime, timeErr := time.Parse(time.RFC3339, getToken.Expiry)
	if timeErr != nil {
		return false, xError.NewError(ctx, xError.OperationFailed, "解析令牌过期时间失败", false, timeErr)
	}
	return parseTime.Before(time.Now()), nil
}

// Logout 调用 OAuth2 Revocation Endpoint 注销指定令牌。
//
// 该方法会把指定 token 发送到 revocation endpoint 完成远端注销，
// 并尝试清理本地缓存中的 access token。缓存清理失败仅记录告警，不阻断主流程。
func (l *OAuthLogic) Logout(ctx *gin.Context, tokenType string, token string) *xError.Error {
	l.log.Info(ctx, "Logout - 注销令牌")

	if tokenType == "" {
		return xError.NewError(ctx, xError.ParameterEmpty, "令牌类型为空", false, nil)
	}
	if token == "" {
		return xError.NewError(ctx, xError.ParameterEmpty, "令牌为空", false, nil)
	}

	revocationURI := xEnv.GetEnvString(bSdkConst.EnvSsoEndpointRevocationURI, "")
	if revocationURI == "" {
		return xError.NewError(ctx, xError.OperationFailed, "注销端点为空", false, nil)
	}

	clientID := xEnv.GetEnvString(bSdkConst.EnvSsoClientID, "")
	clientSecret := xEnv.GetEnvString(bSdkConst.EnvSsoClientSecret, "")
	if clientID == "" || clientSecret == "" {
		return xError.NewError(ctx, xError.OperationFailed, "客户端配置缺失", false, nil)
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
		Post(revocationURI)
	if reqErr != nil {
		return xError.NewError(ctx, xError.OperationFailed, "注销令牌失败", false, reqErr)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return xError.NewError(
			ctx,
			xError.OperationFailed,
			xError.ErrMessage(fmt.Sprintf("注销令牌失败，状态码: %d", resp.StatusCode())),
			false,
			nil,
		)
	}

	if l.rdb != nil {
		if delErr := l.tokenData.Delete(ctx, token); delErr != nil {
			l.log.Warn(ctx, "OAuthLogic|Logout - 清理令牌缓存失败",
				slog.String("token", token),
				slog.String("error", delErr.Error()),
			)
		}
	}

	return nil
}
