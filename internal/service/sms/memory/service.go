package memory

import (
	"context"
	"fmt"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
)

type Service struct {
}

func NewService() sms.Service {
	return &Service{}
}

// Send 发送验证码
func (s *Service) Send(ctx context.Context, tplId string, args []string, phones ...string) error {
	fmt.Println(args)
	return nil
}
