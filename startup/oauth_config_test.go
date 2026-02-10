package bSdkStartup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	xEnv "github.com/bamboo-services/bamboo-base-go/env"
	bSdkConst "github.com/phalanx/beacon-sso-sdk/constant"
	"golang.org/x/oauth2"
)

func TestOAuthConfigWellKnownIncludesIntrospectionAndRevocation(t *testing.T) {
	const (
		authURI          = "https://sso.example.com/oauth2/authorize"
		tokenURI         = "https://sso.example.com/oauth2/token"
		userinfoURI      = "https://sso.example.com/oauth2/userinfo"
		introspectionURI = "https://sso.example.com/oauth2/introspect"
		revocationURI    = "https://sso.example.com/oauth2/revoke"
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"authorization_endpoint":"` + authURI + `","token_endpoint":"` + tokenURI + `","userinfo_endpoint":"` + userinfoURI + `","introspection_endpoint":"` + introspectionURI + `","revocation_endpoint":"` + revocationURI + `"}`))
	}))
	defer srv.Close()

	t.Setenv(bSdkConst.EnvSsoWellKnownURI.String(), srv.URL)
	t.Setenv(bSdkConst.EnvSsoClientID.String(), "client-id")
	t.Setenv(bSdkConst.EnvSsoClientSecret.String(), "client-secret")
	t.Setenv(bSdkConst.EnvSsoRedirectURI.String(), "https://app.example.com/callback")
	unsetEnv(t, bSdkConst.EnvSsoEndpointAuthURI.String())
	unsetEnv(t, bSdkConst.EnvSsoEndpointTokenURI.String())
	unsetEnv(t, bSdkConst.EnvSsoEndpointUserinfoURI.String())
	unsetEnv(t, bSdkConst.EnvSsoEndpointIntrospectionURI.String())
	unsetEnv(t, bSdkConst.EnvSsoEndpointRevocationURI.String())

	node := oAuthConfig()
	value, err := node.Node(context.Background())
	if err != nil {
		t.Fatalf("初始化 OAuth 配置失败: %v", err)
	}

	cfg, ok := value.(*oauth2.Config)
	if !ok {
		t.Fatalf("返回值类型错误，期望 *oauth2.Config")
	}
	if cfg.Endpoint.AuthURL != authURI {
		t.Fatalf("auth url 不匹配，期望 %s，实际 %s", authURI, cfg.Endpoint.AuthURL)
	}
	if cfg.Endpoint.TokenURL != tokenURI {
		t.Fatalf("token url 不匹配，期望 %s，实际 %s", tokenURI, cfg.Endpoint.TokenURL)
	}
	if xEnv.GetEnvString(bSdkConst.EnvSsoEndpointUserinfoURI, "") != userinfoURI {
		t.Fatalf("userinfo endpoint 未正确写入环境变量")
	}
	if xEnv.GetEnvString(bSdkConst.EnvSsoEndpointIntrospectionURI, "") != introspectionURI {
		t.Fatalf("introspection endpoint 未正确写入环境变量")
	}
	if xEnv.GetEnvString(bSdkConst.EnvSsoEndpointRevocationURI, "") != revocationURI {
		t.Fatalf("revocation endpoint 未正确写入环境变量")
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	oldValue, exist := os.LookupEnv(key)
	_ = os.Unsetenv(key)
	t.Cleanup(func() {
		if exist {
			_ = os.Setenv(key, oldValue)
			return
		}
		_ = os.Unsetenv(key)
	})
}
