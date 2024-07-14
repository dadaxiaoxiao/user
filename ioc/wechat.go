package ioc

import (
	"github.com/dadaxiaoxiao/user/internal/service/oauth2/wechat"
	"github.com/dadaxiaoxiao/user/internal/web"
	"os"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("获取系统环境变量 WECHAT_APP_ID 失败 ")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("获取系统环境变量 WECHAT_APP_SECRET 失败 ")
	}
	return wechat.Newservice(appId, appSecret)
}

func InitWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}
