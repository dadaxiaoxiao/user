package async

import (
	"context"
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/dadaxiaoxiao/user/internal/domain"
	"github.com/dadaxiaoxiao/user/internal/repository"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
	"time"
)

// AsyncSMSService 异步短信服务
type AsyncSMSService struct {
	svc  sms.Service
	repo repository.AsyncSmsRepository
	l    accesslog.Logger
}

func NewAsyncSMSService(svc sms.Service, repo repository.AsyncSmsRepository, l accesslog.Logger) sms.Service {
	res := &AsyncSMSService{
		svc:  svc,
		repo: repo,
		l:    l,
	}
	go func() {
		res.StartAsyncCycle()
	}()

	return res
}

// StartAsyncCycle 异步发送消息
// 原理：这是最简单的抢占式调度
func (s *AsyncSMSService) StartAsyncCycle() {
	// 防止在运行测试的时候，会出现偶发性的失败
	time.Sleep(time.Second * 3)
	for {
		s.AsyncSend()
	}
}

// AsyncSend 异步发送
func (s *AsyncSMSService) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 抢占一个异步发送的消息，确保在非常多个实例
	// 比如 k8s 部署了三个 pod，一个请求，只有一个实例能拿到
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	case nil:
		// 执行发送
		// 这个也可以做成配置的
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 发送
		err = s.svc.Send(ctx, as.TplId, as.Args, as.Numbers...)
		if err != nil {
			// 打 log
			s.l.Error("执行异步发送短信失败",
				accesslog.Error(err),
				accesslog.Int64("id", as.Id))
		}
		res := err == nil
		//  回调通知 repository 这一次的执行结果
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			s.l.Error("执行异步发送短信成功，但是标记数据库失败",
				accesslog.Error(err),
				accesslog.Bool("res", res),
				accesslog.Int64("id", as.Id))
		}

	case repository.ErrWaitingSMSNotFound:
		// 睡一秒。这个你可以自己决定
		time.Sleep(time.Second)
	default:
		// 正常来说应该是数据库那边出了问题，
		// 但是为了尽量运行，还是要继续的
		// 你可以稍微睡眠，也可以不睡眠
		// 睡眠的话可以帮你规避掉短时间的网络抖动问题
		s.l.Error("抢占异步发送短信任务失败",
			accesslog.Error(err))
		time.Sleep(time.Second)
	}
}

// Send 短信发送
func (s *AsyncSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 需要同步转异步
	if s.needAsync() {
		// 请求转储到数据
		return s.repo.Add(ctx, domain.AsyncSms{
			TplId:    tplId,
			Args:     args,
			Numbers:  numbers,
			RetryMax: 3,
		})
	}
	return s.svc.Send(ctx, tplId, args, numbers...)
}

// needAsync
// 判定服务商已经崩溃
// 需要同步转异步，考系统容错问题
func (s *AsyncSMSService) needAsync() bool {
	//  各种判定要不要触发异步的方案
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率
	return true
}
