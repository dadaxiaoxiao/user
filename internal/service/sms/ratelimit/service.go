package ratelimit

import (
	"context"
	"fmt"
	"github.com/dadaxiaoxiao/go-pkg/ratelimit"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
)

var errLimited = fmt.Errorf("触发了限流")

// RatelimitSMSService 限流器短信服务
type RatelimitSMSService struct {
	svc   sms.Service       // 短信服务
	limit ratelimit.Limiter //限流器
}

// NewRatelimitSMSService 新建 RatelimitSMSService
func NewRatelimitSMSService(svc sms.Service, limit ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:   svc,
		limit: limit,
	}
}

// Send 修饰器摸实现 限流器+ 短信发送
func (r *RatelimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := r.limit.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候，
		// 可以不限：你的下游很强，业务可用性要求很高，尽量容错策略
		// 包一下这个错误
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if limited {
		// 触发限流
		return errLimited
	}
	err = r.svc.Send(ctx, tpl, args, numbers...)
	return err
}
