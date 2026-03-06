package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
	"github.com/phalanx-labs/beacon-sso-sdk/client/connect/beacon/sso/v1/pbconnect"
)

// AuthService 封装了 AuthService 的 proto 调用逻辑
type AuthService struct {
	headers map[string]string
	client  pbconnect.AuthServiceClient
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(client pbconnect.AuthServiceClient, headers map[string]string) *AuthService {
	return &AuthService{client: client, headers: headers}
}

// RegisterByEmail 通过邮箱注册
//
// 该接口用于通过邮箱验证码完成用户注册，注册成功后自动生成登录 Token。
// 注册流程：
// 1. 验证邮箱验证码
// 2. 检查邮箱是否已注册
// 3. 验证密码强度
// 4. 创建用户账号
// 5. 绑定邮箱并标记为已验证
// 6. 生成登录 Token
func (s *AuthService) RegisterByEmail(ctx context.Context, req *pb.RegisterByEmailRequest) (*pb.RegisterByEmailResponse, error) {
	// 验证数据
	if req.Email == "" {
		return nil, fmt.Errorf("email 不能为空")
	}
	if req.Code == "" {
		return nil, fmt.Errorf("code 不能为空")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password 不能为空")
	}

	// 构建 proto 请求
	protoReq := connect.NewRequest(req)

	// 添加 headers (App 认证凭证)
	for k, v := range s.headers {
		protoReq.Header().Set(k, v)
	}

	// 调用 proto client
	resp, err := s.client.RegisterByEmail(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	return resp.Msg, nil
}
