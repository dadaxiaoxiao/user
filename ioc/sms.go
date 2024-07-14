package ioc

import (
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	pkgratelimit "github.com/dadaxiaoxiao/go-pkg/ratelimit"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
	"github.com/dadaxiaoxiao/user/internal/service/sms/failover"
	"github.com/dadaxiaoxiao/user/internal/service/sms/memory"
	"github.com/dadaxiaoxiao/user/internal/service/sms/metrics"
	"github.com/dadaxiaoxiao/user/internal/service/sms/opentelemetry"
	"github.com/dadaxiaoxiao/user/internal/service/sms/ratelimit"
	"github.com/dadaxiaoxiao/user/internal/service/sms/tencentcloud"
	"github.com/redis/go-redis/v9"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentcloudSms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
	"os"
	"time"
)

// InitSmsService 初始化短信服务
func InitSmsService(redisClient redis.Cmdable) sms.Service {
	smssvcs := []sms.Service{
		initLimitSMSService(redisClient, initTencentSms()),                         //初始化限流器 腾讯云短信服务
		initLimitSMSService(redisClient, initPrometheusDecorator(initMemorySms())), //初始化限流器 本地短信服务
	}
	return initOTELSMSService(initFailoverSMSService(smssvcs))
}

// initTencentSms 初始化腾讯云短信服务
func initTencentSms() sms.Service {
	/*
	 * 腾讯云账户密钥对secretId，secretKey
	 * 因为安全问题，这里采用的是从环境变量读取的方式，需要在环境变量中先设置这两个值
	 */
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("获取系统环境变量 SMS_SECRET_ID 失败 ")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("获取系统环境变量 SMS_SECRET_KEY 失败 ")
	}
	// 实例一个认证对象
	credential := common.NewCredential(secretId, secretKey)
	// 实列一个客户端配置对象
	cpf := profile.NewClientProfile()
	client, err := tencentcloudSms.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		panic(err)
	}
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	l := accesslog.NewZapLogger(log)
	return tencentcloud.NewService(client, "1400855644", "码农小叶个人公众号", l)
}

// initMemorySms 初始化本地短信服务
func initMemorySms() sms.Service {
	return memory.NewService()
}

// initLimitSMSService 初始化限流器短信服务  smssvc 被修饰的短信服务
func initLimitSMSService(redisClient redis.Cmdable, smssvc sms.Service) sms.Service {
	limiter := pkgratelimit.NewRedisSlideWindowLimiter(redisClient,
		pkgratelimit.WithInterval(time.Second),
		pkgratelimit.WithRate(1000))

	service := ratelimit.NewRatelimitSMSService(smssvc, limiter)
	return service
}

// 初始化轮询短信服务
func initFailoverSMSService(svcs []sms.Service) sms.Service {
	failoversvc := failover.NewFailoverSMSService(svcs)
	return failoversvc
}

func initPrometheusDecorator(smssvc sms.Service) sms.Service {
	svc := metrics.NewPrometheusDecorator(smssvc,
		"qinye_yiyi",
		"demo",
		"sms_resp_time",
		"my_instance_1")
	return svc
}

// initOTELSMSService
func initOTELSMSService(smssvc sms.Service) sms.Service {
	svc := opentelemetry.NewService(smssvc)
	return svc
}
