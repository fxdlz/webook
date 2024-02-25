//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/jwt"
	"webook/ioc"
)

var interactiveSvcSet = wire.NewSet(service.NewInteractiveService, repository.NewCachedInteractiveRepository, cache.NewInteractiveRedisCache, dao.NewGORMInteractiveDAO)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitLogger,
		ioc.InitDB, ioc.InitRedis, ioc.InitLocalMem,
		dao.NewArticleGORMDAO,
		dao.NewGORMUserDAO, cache.NewRedisUserCache, cache.NewLocalCodeCache, cache.NewArticleRedisCache,
		repository.NewCacheArticleRepository,
		repository.NewCacheUserRepository, repository.NewCacheCodeRepository,
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewArticleService,
		service.NewCacheUserService, service.NewCacheCodeService,
		interactiveSvcSet,
		web.NewArticleHandler,
		jwt.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
