package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
)

// InitRlockClient 初始化分布式锁客户端
func InitRlockClient(client redis.Cmdable) *rlock.Client {
	return rlock.NewClient(client)
}
