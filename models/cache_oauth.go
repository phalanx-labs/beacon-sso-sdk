package bSdkModels

// CacheOAuth 用于在 OAuth 2.0 认证流程的 State 与 Callback 阶段之间临时存储和校验安全参数
//
// 该结构体通常缓存于 Redis 中，以确保 state 参数的防篡改校验（防止 CSRF 攻击），
// 并支持 PKCE (Proof Key for Code Exchange) 扩展模式的验证器存储。
//
// 字段说明:
//   - State: 生成的随机状态码，用于验证请求的完整性和一致性。
//   - Verifier: PKCE 流程生成的 code_verifier，用于换取令牌时的安全校验。
type CacheOAuth struct {
	State    string `redis:"state" json:"state"`       // State 码
	Verifier string `redis:"verifier" json:"verifier"` // PCKE 挑战验证码
}
