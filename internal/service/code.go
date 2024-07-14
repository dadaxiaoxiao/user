package service

import (
	"context"
	"fmt"
	"github.com/dadaxiaoxiao/user/internal/repository"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
	"math/rand"
)

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

const codeTplId = "1932694"

type SMSCodeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

// NewSMSCodeService 新建 code server 实例
func NewSMSCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &SMSCodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send 生成一个随机验证码，并发送
func (svc *SMSCodeService) Send(ctx context.Context,
// 区别业务场景
	biz string,
	phone string) error {
	// 随机生成验证码
	code := svc.generateCode()
	// 验证码写入缓存
	// codeRepository
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	//发送验证码
	// smsService
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	return err
}

// Verify 验证验证码
func (svc *SMSCodeService) Verify(ctx context.Context,
	biz string,
	phone string,
	inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *SMSCodeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%6d", num)
}
