// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/dadaxiaoxiao/go-pkg/customserver"
	"github.com/dadaxiaoxiao/user/internal/repository"
	"github.com/dadaxiaoxiao/user/internal/repository/cache"
	"github.com/dadaxiaoxiao/user/internal/repository/dao"
	"github.com/dadaxiaoxiao/user/internal/service"
	"github.com/dadaxiaoxiao/user/internal/web"
	"github.com/dadaxiaoxiao/user/internal/web/jwt"
	"github.com/dadaxiaoxiao/user/ioc"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitApp() *customserver.App {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	logger := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, logger)
	db := ioc.InitDB(logger)
	userDao := dao.NewGORMUserDAO(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository, logger)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smsService := ioc.InitSmsService(cmdable)
	codeService := service.NewSMSCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler, logger)
	wechatService := ioc.InitWechatService()
	wechatHandlerConfig := ioc.InitWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, wechatHandlerConfig, handler)
	server := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler)
	app := &customserver.App{
		GinServer: server,
	}
	return app
}

// wire.go:

var thirdProvider = wire.NewSet(ioc.InitDB, ioc.InitEtcd, ioc.InitLogger, ioc.InitRedis, jwt.NewRedisJWTHandler)

var userHdlProvider = wire.NewSet(dao.NewGORMUserDAO, cache.NewRedisUserCache, cache.NewRedisCodeCache, repository.NewCachedUserRepository, repository.NewCachedCodeRepository, ioc.InitSmsService, service.NewUserService, service.NewSMSCodeService, web.NewUserHandler)

var oauth2WechatHdlProvider = wire.NewSet(ioc.InitWechatService, ioc.InitWechatHandlerConfig, web.NewOAuth2WechatHandler)
