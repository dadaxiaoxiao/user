package failover

import (
	"context"
	"errors"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
	"sync/atomic"
)

// FailoverSMSService 轮询短信服务
type FailoverSMSService struct {
	svcs []sms.Service
	idx  uint64
}

// NewFailoverSMSService 新建 轮询短信服务
func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

/*
// Send 修饰器摸实现 轮询发送短信
func (f *FailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	for _, svc := range r.svcs {
		err := svc.Send(ctx, tpl, args, numbers...)
		if err == nil {
			// 发送成功，退出轮询
			return nil
		}
		// 需要打印错误日志
	}
	return errors.New("发送失败，所有短信服务商都尝试过了")
}
*/

// Send 修饰器摸实现 轮询发送短信
func (f *FailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 索引 原子操作+1
	idx := atomic.AddUint64(&f.idx, 1) // 只是在起始位置做了轮询
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tpl, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			// 设置的超时时间到了
			// 调用者主动取消了
			return err

		}
	}
	return errors.New("发送失败，所有短信服务商都尝试过了")
}