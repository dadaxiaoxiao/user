package tencentcloud

import (
	"context"
	"fmt"
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	mysms "github.com/dadaxiaoxiao/user/internal/service/sms"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111" // 引入sms
)

type Service struct {
	// 短信 SdkAppId
	appId *string

	// 短信签名内容
	signName *string

	// 短信客户端
	client *sms.Client

	log accesslog.Logger
}

func NewService(client *sms.Client, appId string, signName string, log accesslog.Logger) mysms.Service {
	return &Service{
		client:   client,
		appId:    ekit.ToPtr[string](appId),
		signName: ekit.ToPtr[string](signName),
		log:      log,
	}
}

// Send 发送短信
func (s *Service) Send(ctx context.Context, templateId string, args []string, phones ...string) error {
	request := sms.NewSendSmsRequest()
	// 短信 SdkAppId
	request.SmsSdkAppId = s.appId
	// 签名内容
	request.SignName = s.signName
	// 模板id
	request.TemplateId = ekit.ToPtr[string](templateId)
	// 模板参数
	request.TemplateParamSet = s.toStringPtrSlice(args)
	// 发送号码
	request.PhoneNumberSet = s.toStringPtrSlice(phones)

	response, err := s.client.SendSms(request)
	s.log.Debug("调用腾讯短信服务",
		accesslog.Any("req", request),
		accesslog.Any("resp", response),
		accesslog.Any("error", err))

	if err != nil {
		return err
	}
	for _, status := range response.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败 %s, %s ", *status.Code, *status.Message)
		}
	}
	return nil
}

// toStringPtrSlice 返回指针类型的切片
func (s *Service) toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
