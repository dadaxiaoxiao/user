package ioc

import (
	"github.com/dadaxiaoxiao/go-pkg/ginx"
	"github.com/dadaxiaoxiao/user/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// InitWebServer 初始化 web 服务
func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler) *ginx.Server {

	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("http", &cfg)
	if err != nil {
		panic(err)
	}

	server := gin.Default()
	// 注册中间件
	server.Use(mdls...)
	// 注册路由
	userHdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	return &ginx.Server{
		Engine: server,
		Addr:   cfg.Addr,
	}
}
