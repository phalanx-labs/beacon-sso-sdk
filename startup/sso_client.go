package bSdkStartup

import (
	"context"
	"log/slog"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xRegNode "github.com/bamboo-services/bamboo-base-go/major/register/node"
	bSdkClient "github.com/phalanx-labs/beacon-sso-sdk/client"
	bSdkConst "github.com/phalanx-labs/beacon-sso-sdk/constant"
)

// ssoClient 初始化 SsoClient 并注册依赖项。
//
// 该函数从环境变量读取 gRPC 连接配置，创建 SsoClient 实例。
// 如果必要的配置缺失，会触发 Panic 终止程序。
//
// 环境变量：
//   - SSO_GRPC_HOST: gRPC 主机地址
//   - SSO_GRPC_PORT: gRPC 端口
//   - SSO_APP_ACCESS_ID: App Access ID
//   - SSO_APP_SECRET_KEY: App Secret Key
//
// 注册的上下文键为 `CtxSsoClient`。
func ssoClient() xRegNode.RegNodeList {
	return xRegNode.RegNodeList{
		Key: bSdkConst.CtxSsoClient,
		Node: func(ctx context.Context) (any, error) {
			log := xLog.WithName(xLog.NamedINIT)
			log.Info(ctx, "初始化 SsoClient")

			// 获取环境变量
			host := xEnv.GetEnvString(bSdkConst.EnvSsoGrpcHost, "")
			port := xEnv.GetEnvString(bSdkConst.EnvSsoGrpcPort, "")
			appClientID := xEnv.GetEnvString(bSdkConst.EnvSsoClientID, "")
			appClientSecret := xEnv.GetEnvString(bSdkConst.EnvSsoClientSecret, "")

			// 校验配置
			if host == "" || port == "" || appClientID == "" || appClientSecret == "" {
				xLog.Panic(ctx, "SSO gRPC 客户端配置缺失",
					slog.String("host", host),
					slog.String("port", port),
					slog.String("app_client_id", appClientID),
					slog.String("app_client_secret", appClientSecret),
				)
			}

			// 创建 SsoClient
			client := bSdkClient.NewClient(
				bSdkClient.WithConnect(host, port),
				bSdkClient.WithAppAccess(appClientID, appClientSecret),
			)

			log.Info(ctx, "SsoClient 初始化成功",
				slog.String("host", host),
				slog.String("port", port),
			)

			return client, nil
		},
	}
}
