package bSdkLogic

import (
	"net/http"
	"net/http/httptest"
	"testing"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	"github.com/gin-gonic/gin"
	bSdkConst "github.com/phalanx-labs/beacon-sso-sdk/constant"
)

func TestOAuthLogicLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := newOAuthTestGinContext()
	logic := NewOAuth(nil)

	t.Run("参数为空", func(t *testing.T) {
		xErr := logic.Logout(ctx, "", "token")
		if xErr == nil || xErr.GetErrorCode().Code != xError.ParameterEmpty.Code {
			t.Fatalf("期望参数为空错误")
		}

		xErr = logic.Logout(ctx, "access_token", "")
		if xErr == nil || xErr.GetErrorCode().Code != xError.ParameterEmpty.Code {
			t.Fatalf("期望参数为空错误")
		}
	})

	t.Run("注销成功", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("请求方法错误: %s", r.Method)
			}
			if user, pass, ok := r.BasicAuth(); !ok || user != "cid" || pass != "csecret" {
				t.Fatalf("Basic Auth 不正确")
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("解析表单失败: %v", err)
			}
			if r.Form.Get("token") != "token-value" {
				t.Fatalf("token 参数错误")
			}
			if r.Form.Get("token_type_hint") != "access_token" {
				t.Fatalf("token_type_hint 参数错误")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		t.Setenv(bSdkConst.EnvSsoEndpointRevocationURI.String(), srv.URL)
		t.Setenv(bSdkConst.EnvSsoClientID.String(), "cid")
		t.Setenv(bSdkConst.EnvSsoClientSecret.String(), "csecret")

		xErr := logic.Logout(ctx, "access_token", "token-value")
		if xErr != nil {
			t.Fatalf("期望注销成功，实际错误: %v", xErr)
		}
	})
}

func newOAuthTestGinContext() *gin.Context {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return ctx
}
