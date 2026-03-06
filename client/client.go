package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"connectrpc.com/connect"
	"github.com/phalanx-labs/beacon-sso-sdk/client/connect/beacon/sso/v1/pbconnect"
	service2 "github.com/phalanx-labs/beacon-sso-sdk/client/service"
	"golang.org/x/net/http2"
)

// Option 定义客户端选项函数
type Option func(*SsoClient)

// WithConnect 设置主机地址和端口
func WithConnect(host, port string) Option {
	return func(c *SsoClient) {
		c.host = host
		c.port = port
	}
}

// WithAppAccess 设置 App 认证信息（用于 AuthService）
func WithAppAccess(appAccessID, appSecretKey string) Option {
	return func(c *SsoClient) {
		c.headers["app-access-id"] = appAccessID
		c.headers["app-secret-key"] = appSecretKey
	}
}

// WithProtoClient 直接传入 proto client（用于测试）
func WithProtoPublicClient(protoClient pbconnect.PublicServiceClient) Option {
	return func(c *SsoClient) {
		c.protoPublicClient = protoClient
	}
}

func WithProtoAuthClient(protoClient pbconnect.AuthServiceClient) Option {
	return func(c *SsoClient) {
		c.protoAuthClient = protoClient
	}
}

// SsoClient SSO SDK 客户端
type SsoClient struct {
	host              string
	port              string
	headers           map[string]string
	protoPublicClient pbconnect.PublicServiceClient
	protoAuthClient   pbconnect.AuthServiceClient
	Public            IPublic
	Auth              IAuth
}

// NewClient 创建并返回一个新的 SsoClient 实例
//
// 支持多种初始化方式:
//   - 方式 1: 通过 host/port 创建
//     client := bSdk.NewClient(WithConnect("localhost", "5566"))
//   - 方式 2: 直接传入 proto client（用于测试)
//     client := bSdk.NewClient(WithProtoPublicClient(publicClient), WithProtoAuthClient(authClient))
//   - 方式 3: 通过 Option 设置认证信息
//     client := bSdk.NewClient(
//     WithConnect("localhost", "5566"),
//     WithAppAccess("your-app-access-id", "your-app-secret-key"),
//     )
func NewClient(opts ...Option) *SsoClient {
	c := &SsoClient{
		headers: make(map[string]string),
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(c)
	}

	// 如果没有传入 proto client，则创建
	if c.protoPublicClient == nil && c.protoAuthClient == nil {
		c.protoPublicClient = c.createProtoPublicClient()
		c.protoAuthClient = c.createProtoAuthClient()
	}

	// 创建服务封装
	c.Public = service2.NewPublicService(c.protoPublicClient, c.headers)
	c.Auth = service2.NewAuthService(c.protoAuthClient, c.headers)

	return c
}

// createProtoPublicClient 创建 h2c proto client for PublicService
func (c *SsoClient) createProtoPublicClient() pbconnect.PublicServiceClient {
	h2cClient := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.Dial(network, addr)
			},
		},
	}
	return pbconnect.NewPublicServiceClient(
		h2cClient,
		fmt.Sprintf("http://%s:%s", c.host, c.port),
		connect.WithGRPC(),
	)
}

// createProtoAuthClient 创建 h2c proto client for AuthService
func (c *SsoClient) createProtoAuthClient() pbconnect.AuthServiceClient {
	h2cClient := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.Dial(network, addr)
			},
		},
	}
	return pbconnect.NewAuthServiceClient(
		h2cClient,
		fmt.Sprintf("http://%s:%s", c.host, c.port),
		connect.WithGRPC(),
	)
}
