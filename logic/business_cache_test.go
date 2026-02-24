package bSdkLogic

import (
	"testing"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	bSdkConst "github.com/phalanx-labs/beacon-sso-sdk/constant"
)

func TestBusinessCacheFlag(t *testing.T) {
	_ = xEnv.SetEnv(bSdkConst.EnvSsoBusinessCache, "true")
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		t.Fatalf("SSO_BUSINESS_CACHE=true 时应解析为 true")
	}

	_ = xEnv.SetEnv(bSdkConst.EnvSsoBusinessCache, "false")
	if xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, true) {
		t.Fatalf("SSO_BUSINESS_CACHE=false 时应解析为 false")
	}
}
