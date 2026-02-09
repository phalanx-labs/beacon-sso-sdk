package bSdkConst

import (
	"fmt"

	xEnv "github.com/bamboo-services/bamboo-base-go/env"
)

type RedisKey string

const (
	RedisOAuthState RedisKey = "oauth:state:%s" // OAuth state 缓存键
	RedisOAuthToken RedisKey = "oauth:token:%s" // OAuth token 缓存键
)

// Get 返回一个格式化后的 `RedisKey`，根据输入参数对原始键进行格式化并生成新的键。
func (k RedisKey) Get(args ...interface{}) RedisKey {
	validKey := xEnv.GetEnvString(xEnv.NoSqlPrefix, "bss:") + string(k)
	return RedisKey(fmt.Sprintf(validKey, args...))
}

// String 返回 `RedisKey` 的字符串表示形式，主要用于将自定义键类型转换为其底层字符串值。
func (k RedisKey) String() string {
	return string(k)
}
