package bSdkLogic

import (
	"net/http"
	"net/http/httptest"
	"testing"

	xError "github.com/bamboo-services/bamboo-base-go/error"
	"github.com/gin-gonic/gin"
)

func TestBusinessLogicUserinfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		accessToken  string
		userinfoURI  string
		server       func() *httptest.Server
		expectErr    *xError.ErrorCode
		expectSub    string
		expectNick   string
		expectPUName string
		expectEmail  string
		expectPhone  string
		expectRawKey string
	}{
		{
			name:        "access token 为空",
			accessToken: "",
			expectErr:   xError.ParameterEmpty,
		},
		{
			name:        "userinfo uri 缺失",
			accessToken: "token",
			expectErr:   xError.OperationFailed,
		},
		{
			name:        "userinfo 返回 401",
			accessToken: "token",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") != "Bearer token" {
						t.Fatalf("Authorization 头不正确: %s", r.Header.Get("Authorization"))
					}
					w.WriteHeader(http.StatusUnauthorized)
				}))
			},
			expectErr: xError.Unauthorized,
		},
		{
			name:        "userinfo 返回非法 json",
			accessToken: "token",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte("{"))
				}))
			},
			expectErr: xError.OperationFailed,
		},
		{
			name:        "userinfo 成功",
			accessToken: "token",
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") != "Bearer token" {
						t.Fatalf("Authorization 头不正确: %s", r.Header.Get("Authorization"))
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"sub":"271545986450423808","nickname":"筱锋xiao_lfeng","preferred_username":"xiao_lfeng","email":"gm@x-lf.cn","phone":"13316569390","tenant":"dev"}`))
				}))
			},
			expectSub:    "271545986450423808",
			expectNick:   "筱锋xiao_lfeng",
			expectPUName: "xiao_lfeng",
			expectEmail:  "gm@x-lf.cn",
			expectPhone:  "13316569390",
			expectRawKey: "tenant",
		},
		{
			name:        "userinfo 网络错误",
			accessToken: "token",
			userinfoURI: "http://127.0.0.1:1/userinfo",
			expectErr:   xError.OperationFailed,
		},
	}

	logic := NewBusiness(nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var closeServer func()
			ctx := newTestGinContext()

			userinfoURI := tt.userinfoURI
			if tt.server != nil {
				srv := tt.server()
				closeServer = srv.Close
				userinfoURI = srv.URL
			}
			if closeServer != nil {
				defer closeServer()
			}

			t.Setenv("SSO_ENDPOINT_USERINFO_URI", userinfoURI)

			userinfo, xErr := logic.Userinfo(ctx, tt.accessToken)
			if tt.expectErr != nil {
				if xErr == nil {
					t.Fatalf("期望错误码 %d，实际无错误", tt.expectErr.Code)
				}
				if xErr.GetErrorCode().Code != tt.expectErr.Code {
					t.Fatalf("期望错误码 %d，实际 %d", tt.expectErr.Code, xErr.GetErrorCode().Code)
				}
				return
			}

			if xErr != nil {
				t.Fatalf("期望成功，实际错误: %v", xErr)
			}
			if userinfo == nil {
				t.Fatalf("期望返回用户信息，实际为 nil")
			}
			if userinfo.Sub != tt.expectSub {
				t.Fatalf("sub 不匹配，期望 %s，实际 %s", tt.expectSub, userinfo.Sub)
			}
			if userinfo.Nickname != tt.expectNick {
				t.Fatalf("nickname 不匹配，期望 %s，实际 %s", tt.expectNick, userinfo.Nickname)
			}
			if userinfo.PreferredUsername != tt.expectPUName {
				t.Fatalf("preferred_username 不匹配，期望 %s，实际 %s", tt.expectPUName, userinfo.PreferredUsername)
			}
			if userinfo.Email != tt.expectEmail {
				t.Fatalf("email 不匹配，期望 %s，实际 %s", tt.expectEmail, userinfo.Email)
			}
			if userinfo.Phone != tt.expectPhone {
				t.Fatalf("phone 不匹配，期望 %s，实际 %s", tt.expectPhone, userinfo.Phone)
			}
			if _, ok := userinfo.Raw[tt.expectRawKey]; !ok {
				t.Fatalf("Raw 未包含扩展字段: %s", tt.expectRawKey)
			}
		})
	}
}

func newTestGinContext() *gin.Context {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	return ctx
}
