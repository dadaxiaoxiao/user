//go:build wireinject

package main

import (
	"github.com/dadaxiaoxiao/go-pkg/customserver"
	"github.com/dadaxiaoxiao/user/internal/repository"
	"github.com/dadaxiaoxiao/user/internal/repository/cache"
	"github.com/dadaxiaoxiao/user/internal/repository/dao"
	"github.com/dadaxiaoxiao/user/internal/service"
	"github.com/dadaxiaoxiao/user/internal/web"
	myjwt "github.com/dadaxiaoxiao/user/internal/web/jwt"
	"github.com/dadaxiaoxiao/user/ioc"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitEtcd,
	ioc.InitLogger,
	ioc.InitRedis,
	myjwt.NewRedisJWTHandler,
)

var userHdlProvider = wire.NewSet(
	dao.NewGORMUserDAO,
	cache.NewRedisUserCache,
	cache.NewRedisCodeCache,
	repository.NewCachedUserRepository,
	repository.NewCachedCodeRepository,
	ioc.InitSmsService,
	service.NewUserService,
	service.NewSMSCodeService,
	web.NewUserHandler,
)

var oauth2WechatHdlProvider = wire.NewSet(
	ioc.InitWechatService,
	ioc.InitWechatHandlerConfig,
	web.NewOAuth2WechatHandler,
)

func InitApp() *customserver.App {
	wire.Build(
		thirdProvider,
		ioc.InitGinMiddlewares,
		userHdlProvider,
		oauth2WechatHdlProvider,
		ioc.InitWebServer,
		// 组装 *App
		wire.Struct(new(customserver.App), "GinServer"),
	)
	return new(customserver.App)
}
