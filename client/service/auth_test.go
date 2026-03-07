package service

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
)

// mockAuthServiceClient 是 pbconnect.AuthServiceClient 的 mock 实现
type mockAuthServiceClient struct {
	changePasswordFunc func(ctx context.Context, req *connect.Request[pb.ChangePasswordRequest]) (*connect.Response[pb.ChangePasswordResponse], error)
}

func (m *mockAuthServiceClient) RegisterByEmail(ctx context.Context, req *connect.Request[pb.RegisterByEmailRequest]) (*connect.Response[pb.RegisterByEmailResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (m *mockAuthServiceClient) PasswordLogin(ctx context.Context, req *connect.Request[pb.PasswordLoginRequest]) (*connect.Response[pb.PasswordLoginResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (m *mockAuthServiceClient) ChangePassword(ctx context.Context, req *connect.Request[pb.ChangePasswordRequest]) (*connect.Response[pb.ChangePasswordResponse], error) {
	if m.changePasswordFunc != nil {
		return m.changePasswordFunc(ctx, req)
	}
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func TestAuthService_ChangePassword(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		req            *pb.ChangePasswordRequest
		mockResponse   *pb.ChangePasswordResponse
		mockError      error
		expectErr      bool
		expectErrMatch string
	}{
		{
			name:           "user_id 为空",
			req:            &pb.ChangePasswordRequest{UserId: "", NewPassword: "NewPass123"},
			expectErr:      true,
			expectErrMatch: "user_id 不能为空",
		},
		{
			name:           "new_password 为空",
			req:            &pb.ChangePasswordRequest{UserId: "271545986450423808", NewPassword: ""},
			expectErr:      true,
			expectErrMatch: "new_password 不能为空",
		},
		{
			name: "成功修改密码（普通模式）",
			req: &pb.ChangePasswordRequest{
				UserId:      "271545986450423808",
				OldPassword: ptrString("OldPass123"),
				NewPassword: "NewPass456",
			},
			mockResponse: &pb.ChangePasswordResponse{},
			expectErr:    false,
		},
		{
			name: "成功修改密码（强制重置模式，无旧密码）",
			req: &pb.ChangePasswordRequest{
				UserId:      "271545986450423808",
				NewPassword: "NewPass456",
			},
			mockResponse: &pb.ChangePasswordResponse{},
			expectErr:    false,
		},
		{
			name: "服务端返回错误",
			req: &pb.ChangePasswordRequest{
				UserId:      "271545986450423808",
				NewPassword: "NewPass456",
			},
			mockError: connect.NewError(connect.CodeInternal, errors.New("服务器内部错误")),
			expectErr: true,
		},
		{
			name: "服务端返回 BadRequest",
			req: &pb.ChangePasswordRequest{
				UserId:      "271545986450423808",
				NewPassword: "weak",
			},
			mockError: connect.NewError(connect.CodeInvalidArgument, errors.New("密码强度不足")),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock client
			mockClient := &mockAuthServiceClient{
				changePasswordFunc: func(ctx context.Context, req *connect.Request[pb.ChangePasswordRequest]) (*connect.Response[pb.ChangePasswordResponse], error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return connect.NewResponse(tt.mockResponse), nil
				},
			}

			// 创建 AuthService
			authService := NewAuthService(mockClient, map[string]string{
				"app-access-id":  "test-access-id",
				"app-secret-key": "test-secret-key",
			})

			// 执行测试
			resp, err := authService.ChangePassword(ctx, tt.req)

			// 验证结果
			if tt.expectErr {
				if err == nil {
					t.Fatalf("期望返回错误，实际返回 nil")
				}
				if tt.expectErrMatch != "" && err.Error() != tt.expectErrMatch {
					t.Fatalf("错误消息不匹配，期望 %q，实际 %q", tt.expectErrMatch, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("期望成功，实际返回错误: %v", err)
			}
			if resp == nil {
				t.Fatalf("期望返回响应，实际为 nil")
			}
		})
	}
}

// ptrString 返回字符串的指针
func ptrString(s string) *string {
	return &s
}
