package bSdkModels

// CacheOAuthToken 用于在 OAuth 2.0 认证流程中缓存已换取的访问令牌信息
//
// 该结构体通常缓存于 Redis 中，以避免每次请求都需要重新访问 OAuth2 平台验证身份，
// 从而减少不必要的请求开销并提升响应速度。
//
// 字段说明:
//   - AccessToken: 访问令牌，用于后续 API 请求的身份认证。
//   - TokenType: 令牌类型，通常为 "Bearer"。
//   - RefreshToken: 刷新令牌，用于在访问令牌过期后获取新的令牌。
//   - Expiry: 令牌过期时间，以 RFC3339 格式存储。
type CacheOAuthToken struct {
	AccessToken  string `redis:"access_token" json:"access_token"`
	TokenType    string `redis:"token_type" json:"token_type"`
	RefreshToken string `redis:"refresh_token" json:"refresh_token"`
	Expiry       string `redis:"expiry" json:"expiry"` // RFC3339 格式
}
