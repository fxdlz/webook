//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/interactive/events"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
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
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	cache2.NewInteractiveRedisCache,
	dao2.NewGORMInteractiveDAO,
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitApp() *App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitDB, ioc.InitRedis, ioc.InitLocalMem,
		ioc.InitEtcd,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		article.NewSaramaSyncProducer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		ioc.InitRlockClient,
		dao.NewArticleGORMDAO,
		dao.NewGORMUserDAO, cache.NewRedisUserCache, cache.NewLocalCodeCache, cache.NewArticleRedisCache,
		repository.NewCacheArticleRepository,
		repository.NewCacheUserRepository, repository.NewCacheCodeRepository,
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewArticleService,
		service.NewCacheUserService, service.NewCacheCodeService,
		interactiveSvcSet,
		ioc.InitIntrClientV1,
		rankingSvcSet,
		ioc.InitRankingJob,
		ioc.InitJobs,
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
