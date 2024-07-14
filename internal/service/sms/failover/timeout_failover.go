package failover

import (
	"context"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
	"sync/atomic"
)

// TimeoutFailoverSMSService 超时轮询短信服务
type TimeoutFailoverSMSService struct {
	svcs []sms.Service
	idx  int32
	// 连续超时次数
	cnt int32
	// 连续超时次数阈值
	threshold int32
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	cnt := atomic.LoadInt32(&t.cnt)
	// svc 下标
	idx := atomic.LoadInt32(&t.idx)
	if cnt >= t.threshold {
		// 触发了阈值，计算新的下标
		newIdx := (idx + 1) % int32(len(t.svcs)) // 防止溢出
		// CAS 操作失败，说明有并发切换成功了
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 切换成功，重新计算连续超时次数
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = atomic.LoadInt32(&t.idx)
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, args, numbers...)
	switch err {
	case nil:
		// 发送成功，重置计算器
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1) // 计数+1
		return err
	default:
		// 如果是别的异常的话，我们保持不动
		return err
	}
}