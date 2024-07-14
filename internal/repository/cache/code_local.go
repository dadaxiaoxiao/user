package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

// LocalCodeCache 本地 Code 缓存
type LocalCodeCache struct {
	cache *lru.Cache
	// 用于保护缓存的互斥锁
	lock       sync.Mutex
	expiration time.Duration
}

func NewLocalCodeCache(cache *lru.Cache, expiration time.Duration) CodeCache {
	return &LocalCodeCache{
		cache:      cache,
		expiration: expiration,
	}
}

// Set 设置验证码 缓存
func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	key := l.key(biz, phone)
	now := time.Now()
	val, ok := l.cache.Get(key)
	if !ok {
		// 没有缓存，没有验证码
		// 设置验证码缓存
		l.cache.Add(key, codeItem{
			code:   code,
			cnt:    3,
			expire: now.Add(l.expiration),
		})
		return nil
	}

	// 类型断言
	item, ok := val.(codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return errors.New("系统错误")
	}

	// 判断过期时间
	if item.expire.Sub(now) > time.Minute*9 {
		return ErrCodeSendTooMany
	}

	// 重新设置验证码
	l.cache.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(l.expiration),
	})
	return nil
}

// Verify 验证验证码
func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	now := time.Now()
	key := l.key(biz, phone)
	val, ok := l.cache.Get(key)
	if !ok {
		// 缓存不存在
		return false, ErrKeyNotExist
	}
	// 类型断言
	item, ok := val.(codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return false, errors.New("系统错误")
	}
	// 判断验证码过期时间
	if item.expire.Sub(now) <= time.Minute*0 {
		return false, ErrCodeVerifyTooManyTimes
	}
	// 判断验证次数
	if item.cnt <= 0 {
		return false, ErrCodeVerifyTooManyTimes
	}
	// 判断验证码
	item.cnt--
	// 更新缓存
	l.cache.Add(key, item)
	return item.code == inputCode, nil
}

// Key 获取缓存key
func (l *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type codeItem struct {
	code string
	// 可验证次数
	cnt int
	// 过期时间
	expire time.Time
}
