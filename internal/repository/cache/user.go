package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dadaxiaoxiao/user/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

//go:generate mockgen.exe -source=./user.go -package=cachemocks -destination=mocks/user.mock.go UserCache
type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
	Delete(ctx context.Context, id int64) error
}

// RedisUserCache 用户缓存
type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

// NewRedisUserCache  新建实现UserCache 接口的实例
func NewRedisUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

// Get 获取缓存
func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, nil
}

func (cache *RedisUserCache) Delete(ctx context.Context, id int64) error {
	return cache.client.Del(ctx, cache.key(id)).Err()
}

// Set 设置缓存
func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
