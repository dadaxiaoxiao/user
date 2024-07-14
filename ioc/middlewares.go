package ioc

import (
	"context"
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/dadaxiaoxiao/go-pkg/ginx"
	midlogger "github.com/dadaxiaoxiao/go-pkg/ginx/middlerwares/logger"
	"github.com/dadaxiaoxiao/go-pkg/ginx/middlerwares/metric"
	midratelimit "github.com/dadaxiaoxiao/go-pkg/ginx/middlerwares/ratelimitx"
	"github.com/dadaxiaoxiao/go-pkg/ratelimit"
	myjwt "github.com/dadaxiaoxiao/user/internal/web/jwt"
	"github.com/dadaxiaoxiao/user/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"
)

// InitGinMiddlewares 初始化中间件
func InitGinMiddlewares(redisClient redis.Cmdable, wtHdl myjwt.Handler, log accesslog.Logger) []gin.HandlerFunc {
	initCodeCounter()
	return []gin.HandlerFunc{
		corsMiddleware(),
		jwtTokenMiddleware(wtHdl),
		rateLimitMiddleware(redisClient),
		loggerMiddleware(log),
		metricMiddleware(),
		otelMiddleware(),
	}
}

// corsMiddleware   跨越注册
func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true, // 允许带上用户认证
		AllowOriginFunc: func(origin string) bool { // 允许哪些源
			if strings.HasPrefix(origin, "http://localhost") {
				// 开发环境
				return true
			}
			return strings.Contains(origin, "your-company.com")
		},
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"}, // 指示哪些头可以安全地暴露给CORS的API
		MaxAge:        12 * time.Hour,
	})
}

// rateLimitMiddleware  限流中间件
func rateLimitMiddleware(redisClient redis.Cmdable) gin.HandlerFunc {
	limiter := ratelimit.NewRedisSlideWindowLimiter(redisClient, ratelimit.WithInterval(time.Second), ratelimit.WithRate(100))
	return midratelimit.NewBuilder(limiter).Build()

}

// jwtTokenMiddleware JWT token 中间件
func jwtTokenMiddleware(wtHdl myjwt.Handler) gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder(wtHdl).
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/oauth2/wechat/authurl").
		IgnorePaths("/oauth2/wechat/callback").
		IgnorePaths("/users/refresh_token").
		IgnorePaths("/test/metric").
		Build()
}

// loggerMiddleware 初始化log 中间件
func loggerMiddleware(log accesslog.Logger) gin.HandlerFunc {
	ml := midlogger.NewBuilder(func(ctx context.Context, al *midlogger.AccessLog) {
		log.Debug("HTTP请求", accesslog.Field{Key: "al", Value: al})
	})
	// 动态开关，结合监听配置文件
	type Config struct {
		Logreq  bool `yaml:"logreq"`
		Logresp bool `yaml:"logresp"`
	}
	var config = Config{
		Logreq:  false,
		Logresp: false,
	}
	err := viper.UnmarshalKey("web", &config)
	if err != nil {
		panic(err)
	}
	if config.Logreq {
		ml.AllowReqBody()
	}
	if config.Logresp {
		ml.AllowRespBody()
	}

	return ml.Builder()
}

func metricMiddleware() gin.HandlerFunc {
	return metric.NewBuilder(
		"qinye_yiyi",
		"demo",
		"gin_http",
		"统计 GIN 的 HTTP 接口",
		"my_instance_1",
	).Build()
}

func initCodeCounter() {
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "qinye_yiyi",
		Subsystem: "demo",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
}

func otelMiddleware() gin.HandlerFunc {
	return otelgin.Middleware("webook")
}
