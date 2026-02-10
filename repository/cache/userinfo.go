package bSdkCache

import (
	"context"
	"errors"
	"fmt"
	"time"

	xCache "github.com/bamboo-services/bamboo-base-go/cache"
	xEnv "github.com/bamboo-services/bamboo-base-go/env"
	bSdkConst "github.com/phalanx/beacon-sso-sdk/constant"
	bSdkModels "github.com/phalanx/beacon-sso-sdk/models"
	"github.com/redis/go-redis/v9"
)

// UserinfoCache 业务层 Userinfo 缓存管理器
//
// 该类型封装了与 Redis 的交互，用于缓存 SSO 用户信息。
// 通过控制键值对的生命周期（TTL）来减少对 SSO 服务的重复请求。
type UserinfoCache xCache.Cache

// NewUserinfoCache 创建并初始化一个 Userinfo 缓存管理器实例
//
// 参数:
//   - rdb: 已初始化的 Redis 客户端连接，用于底层数据交互。
//
// 返回值:
//   - *UserinfoCache: 配置完成的缓存管理器指针，默认 TTL 为 10 分钟。
func NewUserinfoCache(rdb *redis.Client) *UserinfoCache {
	return &UserinfoCache{
		RDB: rdb,
		TTL: time.Minute * 10,
	}
}

// Get 从缓存中获取指定字段的值
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回未命中。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//   - field: 要获取的字段名。
//
// 返回值:
//   - *string: 字段值的指针。
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）。
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) Get(ctx context.Context, accessToken string, field string) (*string, bool, error) {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil, false, nil
	}

	if accessToken == "" {
		return nil, false, fmt.Errorf("令牌为空")
	}
	if field == "" {
		return nil, false, fmt.Errorf("字段为空")
	}

	value, err := c.RDB.HGet(ctx, c.buildKey(accessToken), field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &value, true, nil
}

// GetAllStruct 从缓存中获取完整的 Userinfo 数据结构
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回未命中。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//
// 返回值:
//   - *bSdkModels.CacheBusinessUserinfo: 缓存的用户信息对象。
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）。
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) GetAllStruct(ctx context.Context, accessToken string) (*bSdkModels.CacheBusinessUserinfo, bool, error) {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil, false, nil
	}

	if accessToken == "" {
		return nil, false, fmt.Errorf("令牌为空")
	}

	result, err := c.RDB.HGetAll(ctx, c.buildKey(accessToken)).Result()
	if err != nil {
		return nil, false, err
	}

	if len(result) == 0 {
		return nil, false, nil
	}

	return &bSdkModels.CacheBusinessUserinfo{
		Sub:               result["sub"],
		Nickname:          result["nickname"],
		PreferredUsername: result["preferred_username"],
		Email:             result["email"],
		Phone:             result["phone"],
		Raw:               result["raw"],
	}, true, nil
}

// Set 设置指定字段的值
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回 nil。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//   - field: 要设置的字段名。
//   - value: 字段值的指针。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) Set(ctx context.Context, accessToken string, field string, value *string) error {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil
	}

	if accessToken == "" {
		return fmt.Errorf("令牌为空")
	}
	if field == "" {
		return fmt.Errorf("字段为空")
	}
	if value == nil {
		return fmt.Errorf("缓存值为空")
	}

	if err := c.RDB.HSet(ctx, c.buildKey(accessToken), field, *value).Err(); err != nil {
		return err
	}

	return c.RDB.Expire(ctx, c.buildKey(accessToken), c.TTL).Err()
}

// SetAllStruct 将完整的 Userinfo 数据结构存储到缓存
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回 nil。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//   - userinfo: 要缓存的用户信息对象。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) SetAllStruct(ctx context.Context, accessToken string, userinfo *bSdkModels.CacheBusinessUserinfo) error {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil
	}

	if accessToken == "" {
		return fmt.Errorf("令牌为空")
	}
	if userinfo == nil {
		return fmt.Errorf("缓存值为空")
	}

	if err := c.RDB.HSet(ctx, c.buildKey(accessToken), userinfo).Err(); err != nil {
		return err
	}

	return c.RDB.Expire(ctx, c.buildKey(accessToken), c.TTL).Err()
}

// GetAll 从缓存中获取所有字段和值
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//
// 返回值:
//   - map[string]string: 所有字段和值的映射。
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) GetAll(ctx context.Context, accessToken string) (map[string]string, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("令牌为空")
	}

	return c.RDB.HGetAll(ctx, c.buildKey(accessToken)).Result()
}

// SetAll 批量设置多个字段的值
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回 nil。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//   - fields: 字段名到值指针的映射。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) SetAll(ctx context.Context, accessToken string, fields map[string]*string) error {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil
	}

	if accessToken == "" {
		return fmt.Errorf("令牌为空")
	}
	if len(fields) == 0 {
		return nil
	}

	values := make(map[string]interface{}, len(fields))
	for field, value := range fields {
		if field == "" {
			return fmt.Errorf("字段为空")
		}
		if value == nil {
			return fmt.Errorf("缓存值为空")
		}
		values[field] = *value
	}

	if err := c.RDB.HSet(ctx, c.buildKey(accessToken), values).Err(); err != nil {
		return err
	}
	return c.RDB.Expire(ctx, c.buildKey(accessToken), c.TTL).Err()
}

// Exists 检查指定字段是否存在
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//   - field: 要检查的字段名。
//
// 返回值:
//   - bool: 字段是否存在。
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) Exists(ctx context.Context, accessToken string, field string) (bool, error) {
	if accessToken == "" {
		return false, fmt.Errorf("令牌为空")
	}
	if field == "" {
		return false, fmt.Errorf("字段为空")
	}

	return c.RDB.HExists(ctx, c.buildKey(accessToken), field).Result()
}

// Remove 从缓存中移除指定的字段
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//   - fields: 要移除的字段名列表。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) Remove(ctx context.Context, accessToken string, fields ...string) error {
	if accessToken == "" {
		return fmt.Errorf("令牌为空")
	}
	if len(fields) == 0 {
		return nil
	}

	return c.RDB.HDel(ctx, c.buildKey(accessToken), fields...).Err()
}

// Delete 删除指定令牌的缓存数据
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - accessToken: 访问令牌，用作缓存键。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *UserinfoCache) Delete(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return fmt.Errorf("令牌为空")
	}

	return c.RDB.Del(ctx, c.buildKey(accessToken)).Err()
}

// buildKey 构建 Redis 缓存键
//
// 参数:
//   - accessToken: 访问令牌。
//
// 返回值:
//   - string: 格式化后的缓存键。
func (c *UserinfoCache) buildKey(accessToken string) string {
	return bSdkConst.RedisBusinessUserinfo.Get(accessToken).String()
}
