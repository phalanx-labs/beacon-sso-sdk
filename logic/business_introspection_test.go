package bSdkLogic

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	"github.com/gin-gonic/gin"
	bSdkConst "github.com/phalanx-labs/beacon-sso-sdk/constant"
)

func TestBusinessLogicIntrospection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := newIntrospectionGinContext()
	logic := NewBusiness(nil)

	t.Run("参数为空", func(t *testing.T) {
		_, xErr := logic.Introspection(ctx, "", "token")
		if xErr == nil || xErr.GetErrorCode().Code != xError.ParameterEmpty.Code {
			t.Fatalf("期望参数为空错误")
		}

		_, xErr = logic.Introspection(ctx, "access_token", "")
		if xErr == nil || xErr.GetErrorCode().Code != xError.ParameterEmpty.Code {
			t.Fatalf("期望参数为空错误")
		}
	})

	t.Run("端点缺失", func(t *testing.T) {
		t.Setenv(bSdkConst.EnvSsoEndpointIntrospectionURI.String(), "")
		t.Setenv(bSdkConst.EnvSsoClientID.String(), "cid")
		t.Setenv(bSdkConst.EnvSsoClientSecret.String(), "csecret")

		_, xErr := logic.Introspection(ctx, "access_token", "token")
		if xErr == nil || xErr.GetErrorCode().Code != xError.OperationFailed.Code {
			t.Fatalf("期望端点缺失错误")
		}
	})

	t.Run("调用成功", func(t *testing.T) {
		expUnix := time.Now().Add(5 * time.Minute).Unix()
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

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"active":true,"token_type":"Bearer","exp":` + strconv.FormatInt(expUnix, 10) + `}`))
		}))
		defer srv.Close()

		t.Setenv(bSdkConst.EnvSsoEndpointIntrospectionURI.String(), srv.URL)
		t.Setenv(bSdkConst.EnvSsoClientID.String(), "cid")
		t.Setenv(bSdkConst.EnvSsoClientSecret.String(), "csecret")

		result, xErr := logic.Introspection(ctx, "access_token", "token-value")
		if xErr != nil {
			t.Fatalf("期望成功，实际错误: %v", xErr)
		}
		if result == nil || !result.Active {
			t.Fatalf("期望 active=true")
		}
		if result.TokenType != "Bearer" {
			t.Fatalf("token_type 不匹配")
		}
		if result.Exp != expUnix {
			t.Fatalf("exp 不匹配")
		}
		if result.Expiry == "" {
			t.Fatalf("expiry 不应为空")
		}
		if result.ExpiresIn <= 0 {
			t.Fatalf("expires_in 应大于 0")
		}
	})
}

func newIntrospectionGinContext() *gin.Context {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return ctx
}
