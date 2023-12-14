package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/middleware"
	"webook/internal/web"
	"webook/pkg/ginx/middleware/ratelimit"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddleWares(redisClient redis.Cmdable) []gin.HandlerFunc {
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
				"x-jwt-token",
			},
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				return strings.Contains(origin, "localhost")
			},
			MaxAge: 12 * time.Hour,
		}),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
		(&middleware.LoginJWTMiddlewareBuilder{}).CheckLogin(),
	}
}
