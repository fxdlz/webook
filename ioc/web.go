package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"
	"webook/internal/middleware"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/pkg/ginx"
	"webook/pkg/ginx/middleware/prometheus"
	"webook/pkg/logger"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler,
	wechatHdl *web.OAuth2WechatHandler,
	articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	articleHdl.RegisterRoutes(server)
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddleWares(redisClient redis.Cmdable, handler ijwt.Handler, log logger.LoggerV1) []gin.HandlerFunc {
	pb := &prometheus.Builder{
		Namespace: "fxlz",
		Subsystem: "webook",
		Name:      "resp_time",
		Help:      "请求相应时间监测",
	}
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "fxlz",
		Subsystem: "webook",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowHeaders: []string{
				"Content-Type",
				"Accept",
				"Authorization",
				"Referer",
				"Sec-Ch-Ua",
				"Sec-Ch-Ua-Mobile",
				"Sec-Ch-Ua-Platform",
				"User-Agent",
				"Cookie",
			},
			ExposeHeaders: []string{
				"x-jwt-token", "x-refresh-token",
			},
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				return strings.Contains(origin, "localhost")
			},
			MaxAge: 12 * time.Hour,
		}),
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),
		otelgin.Middleware("webook"),
		//ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 100)).Build(),
		//(&middleware.LoginJWTMiddlewareBuilder{}).CheckLogin(),
		//sessions.Sessions("ssid", cookie.NewStore([]byte(""))),
		//sessions.Sessions("ssid", memstore.NewStore([]byte(""))),
		(middleware.NewLoginJWTMiddlewareBuilder(handler)).CheckLogin(),
		//(middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
		//	log.Debug("", logger.Field{
		//		Key: "AccessLog",
		//		Val: al,
		//	})
		//})).Build(),
	}
}
