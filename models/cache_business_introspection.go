package bSdkModels

// CacheBusinessIntrospection 业务层 Introspection 缓存模型。
//
// 该结构体用于 Redis Hash 存储业务层令牌状态，
// Raw 字段以 JSON 字符串形式保存扩展字段。
type CacheBusinessIntrospection struct {
	Active    bool   `redis:"active" json:"active"`
	TokenType string `redis:"token_type" json:"token_type"`
	Exp       int64  `redis:"exp" json:"exp"`
	Expiry    string `redis:"expiry" json:"expiry"`
	ExpiresIn int64  `redis:"expires_in" json:"expires_in"`
	IsExpired bool   `redis:"is_expired" json:"is_expired"`
	Raw       string `redis:"raw" json:"raw"`
}
