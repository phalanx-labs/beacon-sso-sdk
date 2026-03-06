package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
	"github.com/phalanx-labs/beacon-sso-sdk/client/connect/beacon/sso/v1/pbconnect"
)

// PublicService 封装了 PublicService 的 proto 调用逻辑
type PublicService struct {
	headers map[string]string
	client  pbconnect.PublicServiceClient
}

// NewPublicService 创建 PublicService 实例
func NewPublicService(client pbconnect.PublicServiceClient, headers map[string]string) *PublicService {
	return &PublicService{client: client, headers: headers}
}

// SendRegisterEmailCode 发送注册邮箱验证码
func (s *PublicService) SendRegisterEmailCode(ctx context.Context, req *pb.SendRegisterEmailCodeRequest) (*pb.SendRegisterEmailCodeResponse, error) {
	// 验证数据
	if req.Email == "" {
		return nil, fmt.Errorf("email 不能为空")
	}

	// 构建 proto 请求
	protoReq := connect.NewRequest(req)

	// 添加 headers
	for k, v := range s.headers {
		protoReq.Header().Set(k, v)
	}

	// 调用 proto client
	resp, err := s.client.SendRegisterEmailCode(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	return resp.Msg, nil
}
