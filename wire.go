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
		dao.NewUserDAO, cache.NewUserCache, cache.NewCodeCache,
		repository.NewUserRepository, repository.NewCodeRepository,
		ioc.InitSMSService,
		service.NewUserService, service.NewCodeService,
		web.NewUserHandler,
		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
