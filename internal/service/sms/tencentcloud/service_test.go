package tencentcloud

import (
	"context"
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"testing"
)

func TestSend(t *testing.T) {
	/*
	 * 腾讯云账户密钥对secretId，secretKey
	 * 因为安全问题，这里采用的是从环境变量读取的方式，需要在环境变量中先设置这两个值
	 */
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		t.Fatal()
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		t.Fatal()
	}
	// 实例一个认证对象
	credential := common.NewCredential(secretId, secretKey)
	// 实列一个客户端配置对象
	cpf := profile.NewClientProfile()
	client, err := sms.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		t.Fatal(err)
	}
	logger := accesslog.NewNopLogger()
	if err != nil {
		t.Fatal(err)
	}
	s := NewService(client, "1400855644", "码农小叶个人公众号", logger)

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name: "发送验证码",
			// 模板id
			tplId: "1932694",
			// 验证码
			params: []string{"123456"},
			// 改成你的手机号码
			numbers: []string{"17875513413"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, er)
		})
	}

}
