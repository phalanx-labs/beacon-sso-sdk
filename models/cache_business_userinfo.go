package bSdkModels

// CacheBusinessUserinfo 业务层 Userinfo 缓存模型。
//
// 该结构体用于 Redis Hash 存储业务层用户信息，
// Raw 字段以 JSON 字符串形式保存扩展字段。
type CacheBusinessUserinfo struct {
	Sub               string `redis:"sub" json:"sub"`
	Nickname          string `redis:"nickname" json:"nickname"`
	PreferredUsername string `redis:"preferred_username" json:"preferred_username"`
	Email             string `redis:"email" json:"email"`
	Phone             string `redis:"phone" json:"phone"`
	Raw               string `redis:"raw" json:"raw"`
}
