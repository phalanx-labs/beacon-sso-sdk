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

// OAuthTokenCache OAuth 2.0 令牌缓存管理器
//
// 该类型封装了与 Redis 的交互，用于缓存已换取的 OAuth 访问令牌信息。
// 通过控制键值对的生命周期（TTL）来减少对 OAuth2 平台的重复验证请求。
type OAuthTokenCache xCache.Cache

// NewOAuthTokenCache 创建并初始化一个 OAuth 令牌缓存管理器实例
//
// 参数:
//   - rdb: 已初始化的 Redis 客户端连接，用于底层数据交互。
//
// 返回值:
//   - *OAuthTokenCache: 配置完成的缓存管理器指针，默认 TTL 为 30 天。
func NewOAuthTokenCache(rdb *redis.Client) *OAuthTokenCache {
	return &OAuthTokenCache{
		RDB: rdb,
		TTL: time.Hour * 24 * 30,
	}
}

func (c *OAuthTokenCache) Get(ctx context.Context, key string, field string) (*string, bool, error) {
	if key == "" {
		return nil, false, fmt.Errorf("令牌为空")
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

func (c *OAuthTokenCache) Set(ctx context.Context, key string, field string, value *string) error {
	if key == "" {
		return fmt.Errorf("令牌为空")
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

func (c *OAuthTokenCache) GetAllStruct(ctx context.Context, key string) (*bSdkModels.CacheOAuthToken, error) {
	if key == "" {
		return nil, fmt.Errorf("令牌为空")
	}

	result, err := c.RDB.HGetAll(ctx, c.buildKey(key)).Result()
	if err != nil {
		return nil, err
	}
	return &bSdkModels.CacheOAuthToken{
		AccessToken:  result["access_token"],
		TokenType:    result["token_type"],
		RefreshToken: result["refresh_token"],
		Expiry:       result["expiry"],
	}, nil
}

func (c *OAuthTokenCache) GetAll(ctx context.Context, key string) (map[string]string, error) {
	if key == "" {
		return nil, fmt.Errorf("令牌为空")
	}

	return c.RDB.HGetAll(ctx, c.buildKey(key)).Result()
}

func (c *OAuthTokenCache) SetAll(ctx context.Context, key string, fields map[string]*string) error {
	if key == "" {
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

	if err := c.RDB.HSet(ctx, c.buildKey(key), values).Err(); err != nil {
		return err
	}
	return c.RDB.Expire(ctx, c.buildKey(key), c.TTL).Err()
}

func (c *OAuthTokenCache) SetAllStruct(ctx context.Context, key string, fields *bSdkModels.CacheOAuthToken) error {
	if key == "" {
		return fmt.Errorf("令牌为空")
	}
	if fields == nil {
		return fmt.Errorf("缓存值为空")
	}

	if err := c.RDB.HSet(ctx, c.buildKey(key), fields).Err(); err != nil {
		return err
	}
	return c.RDB.Expire(ctx, c.buildKey(key), c.TTL).Err()
}

func (c *OAuthTokenCache) Exists(ctx context.Context, key string, field string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("令牌为空")
	}
	if field == "" {
		return false, fmt.Errorf("字段为空")
	}

	return c.RDB.HExists(ctx, c.buildKey(key), field).Result()
}

func (c *OAuthTokenCache) Remove(ctx context.Context, key string, fields ...string) error {
	if key == "" {
		return fmt.Errorf("令牌为空")
	}
	if len(fields) == 0 {
		return nil
	}

	return c.RDB.HDel(ctx, c.buildKey(key), fields...).Err()
}

func (c *OAuthTokenCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("令牌为空")
	}

	return c.RDB.Del(ctx, c.buildKey(key)).Err()
}

func (c *OAuthTokenCache) buildKey(token string) string {
	return bSdkConst.RedisOAuthToken.Get(token).String()
}
