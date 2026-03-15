package main

import (
	_ "github.com/bamboo-services/bamboo-base-go/common"
	_ "github.com/phalanx-labs/beacon-sso-sdk/client/connect/beacon/sso/v1/pbconnect"
	"github.com/phalanx-labs/beacon-sso-sdk/docs"
)

// @title Beacon SSO SDK 聚合文档
// @version v1.0.0
// @description Beacon SSO SDK 聚合文档，用于 SDK 内，内置 handler 方法展示。
// @info_instance_name beacon-sso-sdk
func init() {
	docs.SwaggerInfo.InfoInstanceName = "beacon-sso-sdk"
}
