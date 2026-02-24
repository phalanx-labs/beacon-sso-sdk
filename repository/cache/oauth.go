package bSdkCache

import (
	"context"
	"errors"
	"fmt"
	"time"

	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	bSdkConst "github.com/phalanx-labs/beacon-sso-sdk/constant"
	bSdkModels "github.com/phalanx-labs/beacon-sso-sdk/models"
	"github.com/redis/go-redis/v9"
)

// OAuthCache OAuth 2.0 认证流程的缓存管理器
//
// 该类型封装了与 Redis 的交互，用于临时存储 OAuth 上下文信息（如 State 和 PKCE Verifier）。
// 它通过控制键值对的生命周期（TTL）来确保认证状态的有效性和安全性。
//
// 注意: 该实现非并发安全，不建议在多 goroutine 中共享同一实例操作。
type OAuthCache xCache.Cache

// NewOAuthCache 创建并初始化一个 OAuth 2.0 缓存管理器实例
//
// 该函数封装了 Redis 客户端的注入逻辑，并为新实例设置默认的键值对生存时间（TTL）。
// 它被设计为在仓储层或服务层初始化时调用，以便为 OAuth 流程中的临时数据（如
// state 和 code_verifier）提供快速的存储支持。
//
// 参数:
//   - rdb: 已初始化的 Redis 客户端连接，用于底层数据交互。
//
// 返回值:
//   - *OAuthCache: 配置完成的缓存管理器指针，默认 TTL 为 30 分钟。
func NewOAuthCache(rdb *redis.Client) *OAuthCache {
	return &OAuthCache{
		RDB: rdb,
		TTL: time.Minute * 15,
	}
}

func (c *OAuthCache) Get(ctx context.Context, key string, field string) (*string, bool, error) {
	if key == "" {
		return nil, false, fmt.Errorf("状态为空")
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

func (c *OAuthCache) Set(ctx context.Context, key string, field string, value *string) error {
	if key == "" {
		return fmt.Errorf("状态为空")
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

func (c *OAuthCache) GetAllStruct(ctx context.Context, key string) (*bSdkModels.CacheOAuth, error) {
	if key == "" {
		return nil, fmt.Errorf("状态为空")
	}

	result, err := c.RDB.HGetAll(ctx, c.buildKey(key)).Result()
	if err != nil {
		return nil, err
	}
	return &bSdkModels.CacheOAuth{
		State:    result["state"],
		Verifier: result["verifier"],
	}, nil
}

func (c *OAuthCache) GetAll(ctx context.Context, key string) (map[string]string, error) {
	if key == "" {
		return nil, fmt.Errorf("状态为空")
	}

	return c.RDB.HGetAll(ctx, c.buildKey(key)).Result()
}

func (c *OAuthCache) SetAll(ctx context.Context, key string, fields map[string]*string) error {
	if key == "" {
		return fmt.Errorf("状态为空")
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

func (c *OAuthCache) SetAllStruct(ctx context.Context, key string, fields *bSdkModels.CacheOAuth) error {
	if key == "" {
		return fmt.Errorf("状态为空")
	}
	if fields == nil {
		return fmt.Errorf("缓存值为空")
	}

	if err := c.RDB.HSet(ctx, c.buildKey(key), fields).Err(); err != nil {
		return err
	}
	return c.RDB.Expire(ctx, c.buildKey(key), c.TTL).Err()
}

func (c *OAuthCache) Exists(ctx context.Context, key string, field string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("状态为空")
	}
	if field == "" {
		return false, fmt.Errorf("字段为空")
	}

	return c.RDB.HExists(ctx, c.buildKey(key), field).Result()
}

func (c *OAuthCache) Remove(ctx context.Context, key string, fields ...string) error {
	if key == "" {
		return fmt.Errorf("状态为空")
	}
	if len(fields) == 0 {
		return nil
	}

	return c.RDB.HDel(ctx, c.buildKey(key), fields...).Err()
}

func (c *OAuthCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("状态为空")
	}

	return c.RDB.Del(ctx, c.buildKey(key)).Err()
}

func (c *OAuthCache) buildKey(state string) string {
	return bSdkConst.RedisOAuthState.Get(state).String()
}
