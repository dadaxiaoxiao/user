package wechat

import (
	"context"
	"github.com/dadaxiaoxiao/user/internal/domain"
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}
