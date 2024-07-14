package wechat

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_service_VerifyCode(t *testing.T) {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("获取系统环境变量 WECHAT_APP_ID 失败 ")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("获取系统环境变量 WECHAT_APP_SECRET 失败 ")
	}

	svc := Newservice(appId, appSecret)
	res, err := svc.VerifyCode(context.Background(), "011lbkFa1oKKeG0GhiHa1A9afp3lbkFx")
	require.NoError(t, err)
	t.Log(res)
}
