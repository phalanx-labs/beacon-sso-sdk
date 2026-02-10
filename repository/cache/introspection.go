package bSdkCache

import (
	"context"
	"errors"
	"fmt"
	"time"

	xCache "github.com/bamboo-services/bamboo-base-go/cache"
	xEnv "github.com/bamboo-services/bamboo-base-go/env"
	xUtil "github.com/bamboo-services/bamboo-base-go/utility"
	bSdkConst "github.com/phalanx/beacon-sso-sdk/constant"
	bSdkModels "github.com/phalanx/beacon-sso-sdk/models"
	"github.com/redis/go-redis/v9"
)

// IntrospectionCache 业务层 Introspection 缓存管理器
//
// 该类型封装了与 Redis 的交互，用于缓存 SSO 令牌自省结果。
// 通过控制键值对的生命周期（TTL）来减少对 SSO 服务的重复请求。
type IntrospectionCache xCache.Cache

// NewIntrospectionCache 创建并初始化一个 Introspection 缓存管理器实例
//
// 参数:
//   - rdb: 已初始化的 Redis 客户端连接，用于底层数据交互。
//
// 返回值:
//   - *IntrospectionCache: 配置完成的缓存管理器指针，默认 TTL 为 15 分钟。
func NewIntrospectionCache(rdb *redis.Client) *IntrospectionCache {
	return &IntrospectionCache{
		RDB: rdb,
		TTL: time.Minute * 15,
	}
}

// Get 从缓存中获取指定字段的值
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回未命中。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//   - field: 要获取的字段名。
//
// 返回值:
//   - *string: 字段值的指针。
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）。
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) Get(ctx context.Context, key string, field string) (*string, bool, error) {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil, false, nil
	}

	if key == "" {
		return nil, false, fmt.Errorf("缓存键为空")
	}
	if field == "" {
		return nil, false, fmt.Errorf("字段为空")
	}

	value, err := c.RDB.HGet(ctx, c.buildKey(key), field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &value, true, nil
}

// GetAllStruct 从缓存中获取完整的 Introspection 数据结构
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回未命中。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//
// 返回值:
//   - *bSdkModels.CacheBusinessIntrospection: 缓存的令牌自省结果对象。
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）。
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) GetAllStruct(ctx context.Context, key string) (*bSdkModels.CacheBusinessIntrospection, bool, error) {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil, false, nil
	}

	if key == "" {
		return nil, false, fmt.Errorf("缓存键为空")
	}

	result, err := c.RDB.HGetAll(ctx, c.buildKey(key)).Result()
	if err != nil {
		return nil, false, err
	}

	if len(result) == 0 {
		return nil, false, nil
	}

	active, _ := xUtil.Parse().Bool(result["active"])
	exp, _ := xUtil.Parse().Int64(result["exp"])
	expiresIn, _ := xUtil.Parse().Int64(result["expires_in"])
	isExpired, _ := xUtil.Parse().Bool(result["is_expired"])

	return &bSdkModels.CacheBusinessIntrospection{
		Active:    active,
		TokenType: result["token_type"],
		Exp:       exp,
		Expiry:    result["expiry"],
		ExpiresIn: expiresIn,
		IsExpired: isExpired,
		Raw:       result["raw"],
	}, true, nil
}

// Set 设置指定字段的值
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回 nil。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//   - field: 要设置的字段名。
//   - value: 字段值的指针。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) Set(ctx context.Context, key string, field string, value *string) error {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil
	}

	if key == "" {
		return fmt.Errorf("缓存键为空")
	}
	if field == "" {
		return fmt.Errorf("字段为空")
	}
	if value == nil {
		return fmt.Errorf("缓存值为空")
	}

	if err := c.RDB.HSet(ctx, c.buildKey(key), field, *value).Err(); err != nil {
		return err
	}

	return c.RDB.Expire(ctx, c.buildKey(key), c.TTL).Err()
}

// SetAllStruct 将完整的 Introspection 数据结构存储到缓存
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回 nil。
// TTL 会根据令牌的 ExpiresIn 字段动态计算，最大不超过 15 分钟。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//   - introspection: 要缓存的令牌自省结果对象。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) SetAllStruct(ctx context.Context, key string, introspection *bSdkModels.CacheBusinessIntrospection) error {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil
	}

	if key == "" {
		return fmt.Errorf("缓存键为空")
	}
	if introspection == nil {
		return fmt.Errorf("缓存值为空")
	}

	if err := c.RDB.HSet(ctx, c.buildKey(key), introspection).Err(); err != nil {
		return err
	}

	ttl := c.calculateTTL(introspection.ExpiresIn)
	return c.RDB.Expire(ctx, c.buildKey(key), ttl).Err()
}

// GetAll 从缓存中获取所有字段和值
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//
// 返回值:
//   - map[string]string: 所有字段和值的映射。
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) GetAll(ctx context.Context, key string) (map[string]string, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键为空")
	}

	return c.RDB.HGetAll(ctx, c.buildKey(key)).Result()
}

// SetAll 批量设置多个字段的值
//
// 该方法会检查缓存开关环境变量，如果缓存未启用则直接返回 nil。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//   - fields: 字段名到值指针的映射。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) SetAll(ctx context.Context, key string, fields map[string]*string) error {
	if !xEnv.GetEnvBool(bSdkConst.EnvSsoBusinessCache, false) {
		return nil
	}

	if key == "" {
		return fmt.Errorf("缓存键为空")
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

	if err := c.RDB.HSet(ctx, c.buildKey(key), values).Err(); err != nil {
		return err
	}
	return c.RDB.Expire(ctx, c.buildKey(key), c.TTL).Err()
}

// Exists 检查指定字段是否存在
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//   - field: 要检查的字段名。
//
// 返回值:
//   - bool: 字段是否存在。
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) Exists(ctx context.Context, key string, field string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键为空")
	}
	if field == "" {
		return false, fmt.Errorf("字段为空")
	}

	return c.RDB.HExists(ctx, c.buildKey(key), field).Result()
}

// Remove 从缓存中移除指定的字段
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//   - fields: 要移除的字段名列表。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) Remove(ctx context.Context, key string, fields ...string) error {
	if key == "" {
		return fmt.Errorf("缓存键为空")
	}
	if len(fields) == 0 {
		return nil
	}

	return c.RDB.HDel(ctx, c.buildKey(key), fields...).Err()
}

// Delete 删除指定令牌的缓存数据
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文。
//   - key: 缓存键（已组合的键）。
//
// 返回值:
//   - error: 操作过程中发生的错误。
func (c *IntrospectionCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("缓存键为空")
	}

	return c.RDB.Del(ctx, c.buildKey(key)).Err()
}

// calculateTTL 计算动态 TTL
//
// 根据令牌的 ExpiresIn 字段计算缓存的 TTL，最大不超过 15 分钟。
//
// 参数:
//   - expiresIn: 令牌的剩余有效时间（秒）。
//
// 返回值:
//   - time.Duration: 计算后的 TTL。
func (c *IntrospectionCache) calculateTTL(expiresIn int64) time.Duration {
	if expiresIn <= 0 {
		return c.TTL
	}

	tokenTTL := time.Duration(expiresIn) * time.Second
	if tokenTTL > c.TTL {
		return c.TTL
	}

	return tokenTTL
}

// buildKey 构建 Redis 缓存键
//
// 参数:
//   - key: 缓存键（已组合的键，格式为 tokenType:token）。
//
// 返回值:
//   - string: 格式化后的缓存键。
func (c *IntrospectionCache) buildKey(key string) string {
	return bSdkConst.RedisBusinessIntrospection.Get(key).String()
}
