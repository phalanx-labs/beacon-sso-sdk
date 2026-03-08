package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	pb "github.com/phalanx-labs/beacon-sso-sdk/client/api/beacon/sso/v1"
	"github.com/phalanx-labs/beacon-sso-sdk/client/connect/beacon/sso/v1/pbconnect"
)

// MerchantService 封装了 MerchantService 的 proto 调用逻辑
type MerchantService struct {
	headers map[string]string
	client  pbconnect.MerchantServiceClient
}

// NewMerchantService 创建 MerchantService 实例
func NewMerchantService(client pbconnect.MerchantServiceClient, headers map[string]string) *MerchantService {
	return &MerchantService{client: client, headers: headers}
}

// MerchantTag 商户标签信息
type MerchantTag struct {
	ID          string
	Code        string
	Name        string
	Description *string
	Color       *string
	Icon        *string
	SortOrder   int32
	Status      int32
}

// ConvertFromPB 将 proto MerchantTag 转换为本地结构
func (t *MerchantTag) ConvertFromPB(pb *pb.MerchantTag) {
	if pb == nil {
		return
	}
	t.ID = pb.Id
	t.Code = pb.Code
	t.Name = pb.Name
	t.Description = pb.Description
	t.Color = pb.Color
	t.Icon = pb.Icon
	t.SortOrder = pb.SortOrder
	t.Status = pb.Status
}

// MerchantAnnouncement 商户公告信息
type MerchantAnnouncement struct {
	ID           string
	Title        string
	Content      string
	Scope        int32
	DisplayUntil *string
	CreatedAt    string
}

// ConvertFromPB 将 proto MerchantAnnouncement 转换为本地结构
func (a *MerchantAnnouncement) ConvertFromPB(pb *pb.MerchantAnnouncement) {
	if pb == nil {
		return
	}
	a.ID = pb.Id
	a.Title = pb.Title
	a.Content = pb.Content
	a.Scope = pb.Scope
	a.DisplayUntil = pb.DisplayUntil
	a.CreatedAt = pb.CreatedAt
}

// AnnouncementListMeta 公告列表元信息
type AnnouncementListMeta struct {
	MD5Hash     string
	SHA256Hash  string
	Count       int32
	GeneratedAt string
}

// ConvertFromPB 将 proto AnnouncementListMeta 转换为本地结构
func (m *AnnouncementListMeta) ConvertFromPB(pb *pb.AnnouncementListMeta) {
	if pb == nil {
		return
	}
	m.MD5Hash = pb.Md5Hash
	m.SHA256Hash = pb.Sha256Hash
	m.Count = pb.Count
	m.GeneratedAt = pb.GeneratedAt
}

// GetMerchantTagsRequest 获取商户标签列表请求
type GetMerchantTagsRequest struct {
	EnabledOnly bool
}

// GetMerchantTagsResponse 获取商户标签列表响应
type GetMerchantTagsResponse struct {
	Tags []MerchantTag
}

// GetMerchantTags 获取当前应用所属商户的所有标签
//
// 返回应用所属商户下的所有标签列表，可用于客户端展示或标签匹配。
func (s *MerchantService) GetMerchantTags(ctx context.Context, req *GetMerchantTagsRequest) (*GetMerchantTagsResponse, error) {
	// 构建 proto 请求
	protoReq := connect.NewRequest(&pb.GetMerchantTagsRequest{
		EnabledOnly: &req.EnabledOnly,
	})

	// 添加 headers (App 认证凭证)
	for k, v := range s.headers {
		protoReq.Header().Set(k, v)
	}

	// 调用 proto client
	resp, err := s.client.GetMerchantTags(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	tags := make([]MerchantTag, 0, len(resp.Msg.Tags))
	for _, t := range resp.Msg.Tags {
		var tag MerchantTag
		tag.ConvertFromPB(t)
		tags = append(tags, tag)
	}

	return &GetMerchantTagsResponse{Tags: tags}, nil
}

// GetUserTagsRequest 获取用户标签列表请求
type GetUserTagsRequest struct {
	UserID      string
	EnabledOnly bool
}

// GetUserTagsResponse 获取用户标签列表响应
type GetUserTagsResponse struct {
	Tags []MerchantTag
}

// GetUserTags 获取指定用户在当前商户的所有标签
//
// 根据用户 ID 获取该用户在当前应用所属商户下的所有标签。
func (s *MerchantService) GetUserTags(ctx context.Context, req *GetUserTagsRequest) (*GetUserTagsResponse, error) {
	// 验证数据
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id 不能为空")
	}

	// 构建 proto 请求
	protoReq := connect.NewRequest(&pb.GetUserTagsRequest{
		UserId:      req.UserID,
		EnabledOnly: &req.EnabledOnly,
	})

	// 添加 headers (App 认证凭证)
	for k, v := range s.headers {
		protoReq.Header().Set(k, v)
	}

	// 调用 proto client
	resp, err := s.client.GetUserTags(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	tags := make([]MerchantTag, 0, len(resp.Msg.Tags))
	for _, t := range resp.Msg.Tags {
		var tag MerchantTag
		tag.ConvertFromPB(t)
		tags = append(tags, tag)
	}

	return &GetUserTagsResponse{Tags: tags}, nil
}

// CheckUserHasTagRequest 检查用户是否有指定标签请求
type CheckUserHasTagRequest struct {
	UserID  string
	TagCode string
}

// CheckUserHasTagResponse 检查用户是否有指定标签响应
type CheckUserHasTagResponse struct {
	HasTag bool
}

// CheckUserHasTag 检查用户是否有指定标签
//
// 通过标签代码（code）快速检查用户是否拥有该标签。
func (s *MerchantService) CheckUserHasTag(ctx context.Context, req *CheckUserHasTagRequest) (*CheckUserHasTagResponse, error) {
	// 验证数据
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id 不能为空")
	}
	if req.TagCode == "" {
		return nil, fmt.Errorf("tag_code 不能为空")
	}

	// 构建 proto 请求
	protoReq := connect.NewRequest(&pb.CheckUserHasTagRequest{
		UserId:  req.UserID,
		TagCode: req.TagCode,
	})

	// 添加 headers (App 认证凭证)
	for k, v := range s.headers {
		protoReq.Header().Set(k, v)
	}

	// 调用 proto client
	resp, err := s.client.CheckUserHasTag(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	return &CheckUserHasTagResponse{HasTag: resp.Msg.HasTag}, nil
}

// GetRecentAnnouncementsRequest 获取最近公告请求
type GetRecentAnnouncementsRequest struct {
	Limit      int32
	ActiveOnly bool
}

// GetRecentAnnouncementsResponse 获取最近公告响应
type GetRecentAnnouncementsResponse struct {
	Announcements []MerchantAnnouncement
	Meta          AnnouncementListMeta
}

// GetRecentAnnouncements 获取最近公告列表
//
// 获取当前应用所属商户的最近公告（最多 10 条）。
// 返回结果包含 MD5 和 SHA256 哈希值，客户端可用于判断是否需要重新展示。
//
// 使用场景：
//   - 客户端首次获取公告后，存储哈希值
//   - 后续请求时，对比哈希值是否变化
//   - 如果哈希值相同，说明公告无变化，可不展示
//   - 如果哈希值不同，说明公告有更新，需要展示
func (s *MerchantService) GetRecentAnnouncements(ctx context.Context, req *GetRecentAnnouncementsRequest) (*GetRecentAnnouncementsResponse, error) {
	// 构建 proto 请求
	protoReq := connect.NewRequest(&pb.GetRecentAnnouncementsRequest{
		Limit:      &req.Limit,
		ActiveOnly: &req.ActiveOnly,
	})

	// 添加 headers (App 认证凭证)
	for k, v := range s.headers {
		protoReq.Header().Set(k, v)
	}

	// 调用 proto client
	resp, err := s.client.GetRecentAnnouncements(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	announcements := make([]MerchantAnnouncement, 0, len(resp.Msg.Announcements))
	for _, a := range resp.Msg.Announcements {
		var ann MerchantAnnouncement
		ann.ConvertFromPB(a)
		announcements = append(announcements, ann)
	}

	var meta AnnouncementListMeta
	meta.ConvertFromPB(resp.Msg.Meta)

	return &GetRecentAnnouncementsResponse{
		Announcements: announcements,
		Meta:          meta,
	}, nil
}

// GetAnnouncementRequest 获取单个公告请求
type GetAnnouncementRequest struct {
	AnnouncementID string
}

// GetAnnouncementResponse 获取单个公告响应
type GetAnnouncementResponse struct {
	Announcement MerchantAnnouncement
}

// GetAnnouncement 获取单个公告详情
//
// 根据公告 ID 获取详细信息。
func (s *MerchantService) GetAnnouncement(ctx context.Context, req *GetAnnouncementRequest) (*GetAnnouncementResponse, error) {
	// 验证数据
	if req.AnnouncementID == "" {
		return nil, fmt.Errorf("announcement_id 不能为空")
	}

	// 构建 proto 请求
	protoReq := connect.NewRequest(&pb.GetAnnouncementRequest{
		AnnouncementId: req.AnnouncementID,
	})

	// 添加 headers (App 认证凭证)
	for k, v := range s.headers {
		protoReq.Header().Set(k, v)
	}

	// 调用 proto client
	resp, err := s.client.GetAnnouncement(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	var ann MerchantAnnouncement
	ann.ConvertFromPB(resp.Msg.Announcement)

	return &GetAnnouncementResponse{Announcement: ann}, nil
}
