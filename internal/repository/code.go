package repository

import (
	"context"
	"github.com/dadaxiaoxiao/user/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CodeRepository interface {
	Store(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

// NewCachedCodeRepository  使用了缓存的 CodeRepository 实现
func NewCachedCodeRepository(cache cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: cache,
	}
}

// Store 设置验证码缓存
func (r *CachedCodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return r.cache.Set(ctx, biz, phone, code)
}

// Verify 验证验证码
func (r *CachedCodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return r.cache.Verify(ctx, biz, phone, inputCode)
}
