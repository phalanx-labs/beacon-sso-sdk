package bSdkModels

// OAuthIntrospection 表示 RFC 7662 自省接口的关键信息。
//
// 该结构体用于返回令牌当前状态以及有效期信息，
// 同时保留原始响应以兼容供应商扩展字段。
type OAuthIntrospection struct {
	Active    bool           `json:"active"`
	TokenType string         `json:"token_type,omitempty"`
	Exp       int64          `json:"exp,omitempty"`
	Expiry    string         `json:"expiry,omitempty"`
	ExpiresIn int64          `json:"expires_in,omitempty"`
	IsExpired bool           `json:"is_expired"`
	Raw       map[string]any `json:"raw,omitempty"`
}
