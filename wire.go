//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/jwt"
	"webook/ioc"
)

var interactiveSvcSet = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	cache.NewInteractiveRedisCache,
	dao.NewGORMInteractiveDAO,
)

func InitApp() *App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitDB, ioc.InitRedis, ioc.InitLocalMem,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventConsumer,
		ioc.InitConsumer,
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
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
