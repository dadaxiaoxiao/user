package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type RedisCodeCache struct {
	client redis.Cmdable
}

// 编译器会在编译的时候,把set_code.lua 的代码放进去 luaSetCode 变量
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

func NewRedisCodeCache(client redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

// Set 设置验证码 缓存
// ctx 链式调用 , biz 业务场景 ，phone 手机号，code 验证码
func (cache *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := cache.client.Eval(ctx, luaSetCode, []string{cache.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1:
		return ErrCodeSendTooMany
	default:
		return errors.New("系统错误")
	}
}

// Verify 验证验证码
// ctx 链式调用 , biz 业务场景 ，phone 手机号，code 验证码
// res bool 验证结果，error 错误结果
func (cache *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := cache.client.Eval(ctx, luaVerifyCode, []string{cache.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		// 验证码正确
		return true, nil
	case -1:
		// 验证码次数太多
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		// 验证码验证错误
		return false, nil
	}
	return false, errors.New("系统错误")
}

// Key 获取缓存key
func (cache *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
