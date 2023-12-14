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
	"webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		dao.NewGORMUserDAO, cache.NewRedisUserCache, cache.NewRedisCodeCache,
		repository.NewCacheUserRepository, repository.NewCacheCodeRepository,
		ioc.InitSMSService,
		service.NewCacheUserService, service.NewCacheCodeService,
		web.NewUserHandler,
		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
