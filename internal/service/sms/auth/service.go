package auth

import (
	"context"
	"errors"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

type AuthSMSService struct {
	svc sms.Service
	key string
}

func NewAuthSMSService(svc sms.Service, key string) *AuthSMSService {
	return &AuthSMSService{
		svc: svc,
		key: key,
	}
}

// Send 发送，其中 biz 必须是线下申请的一个代表业务方的 token
func (r *AuthSMSService) Send(ctx context.Context, biz string, args []string, phones ...string) error {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(biz, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.key), nil
	})
	if err != nil {
		return err
	}
	if token == nil || !token.Valid {
		return errors.New("token 不合法")
	}
	return r.svc.Send(ctx, claims.Tpl, args, phones...)
}

type Claims struct {
	jwt.RegisteredClaims        // 使用了组合
	Tpl                  string // 短信模板
}
