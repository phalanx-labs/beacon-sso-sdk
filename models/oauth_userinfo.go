package bSdkModels

// OAuthUserinfo 表示 OAuth2 Userinfo Endpoint 的标准用户信息响应。
//
// 该结构体聚合了常见的 OIDC 用户字段，同时保留原始响应内容，
// 以便兼容不同 SSO 提供商的扩展字段。
type OAuthUserinfo struct {
	Sub               string         `json:"sub"`
	Nickname          string         `json:"nickname,omitempty"`
	PreferredUsername string         `json:"preferred_username,omitempty"`
	Email             string         `json:"email,omitempty"`
	Phone             string         `json:"phone,omitempty"`
	Raw               map[string]any `json:"raw,omitempty"`
}
