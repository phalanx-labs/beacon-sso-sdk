package client

// SSOClient SSO SDK 客户端
type SSOClient struct {
	host    string
	port    string
	headers map[string]string

	// 服务实例
	Public IPublic
	Auth   IAuth
}
