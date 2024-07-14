package sms

import "context"

// Service 发送短信的抽象
// 屏蔽不同供应商之间的区别
//
//go:generate mockgen.exe -source=./types.go  -package=smsmocks -destination=mocks/svc.mock.go
type Service interface {
	// Send biz 很含糊的业务
	Send(ctx context.Context, biz string, args []string, phones ...string) error
}
