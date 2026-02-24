package bSdkStartup

import (
	"context"
	"testing"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	bSdkConst "github.com/phalanx-labs/beacon-sso-sdk/constant"
)

func TestOAuthRedirectURI(t *testing.T) {
	const (
		userinfoURI = "https://sso.example.com/oauth2/userinfo"
		redirectURI = "https://app.example.com/callback"
	)
	_ = xEnv.SetEnv(bSdkConst.EnvSsoEndpointUserinfoURI, userinfoURI)
	_ = xEnv.SetEnv(bSdkConst.EnvSsoRedirectURI, redirectURI)

	node := oAuthRedirectURI()
	if node.Key != bSdkConst.CtxOAuthUserinfoURI {
		t.Fatalf("注入键不正确，期望 %s，实际 %s", bSdkConst.CtxOAuthUserinfoURI.String(), node.Key.String())
	}

	value, err := node.Node(context.Background())
	if err != nil {
		t.Fatalf("执行注入节点失败: %v", err)
	}

	uri, ok := value.(string)
	if !ok {
		t.Fatalf("注入值类型不正确，期望 string")
	}
	if uri != redirectURI {
		t.Fatalf("注入值不正确，期望 %s，实际 %s", redirectURI, uri)
	}
}
