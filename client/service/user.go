package service

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
	"github.com/phalanx-labs/beacon-sso-sdk/client/connect/beacon/sso/v1/pbconnect"
)

// UserService 封装了 UserService 的 proto 调用逻辑
type UserService struct {
	headers map[string]string
	client  pbconnect.UserServiceClient
}

// NewUserService 创建 UserService 实例
func NewUserService(client pbconnect.UserServiceClient, headers map[string]string) *UserService {
	return &UserService{client: client, headers: headers}
}

// GetCurrentUser 获取当前登录用户的详细信息
func (s *UserService) GetCurrentUser(ctx context.Context, accessToken string) (*pb.GetCurrentUserResponse, error) {
	authorization := strings.TrimSpace(accessToken)
	if authorization == "" {
		return nil, fmt.Errorf("access_token 不能为空")
	}

	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		authorization = "Bearer " + authorization[7:]
	} else if !strings.HasPrefix(authorization, "Bearer ") {
		authorization = "Bearer " + authorization
	}

	// 构建 proto 请求
	protoReq := connect.NewRequest(&pb.GetCurrentUserRequest{})

	// 添加 headers (App 认证凭证)
	for k, v := range s.headers {
		protoReq.Header().Set(k, v)
	}
	protoReq.Header().Set("authorization", authorization)

	// 调用 proto client
	resp, err := s.client.GetCurrentUser(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}
