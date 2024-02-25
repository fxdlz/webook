package ioc

import (
	"github.com/redis/go-redis/v9"
	"time"
	"webook/config"
	"webook/internal/service/sms"
	"webook/internal/service/sms/local"
	"webook/internal/service/sms/ratelimit"
	"webook/pkg/limiter"
)

func InitSMSService() sms.Service {
	//return local.NewLocalSMSService()
	return ratelimit.NewRateLimitSMSService(local.NewLocalSMSService(), limiter.NewRedisSlidingWindowLimiter(redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	}), time.Minute*2, 5))
}
