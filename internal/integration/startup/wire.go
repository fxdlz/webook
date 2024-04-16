//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/jwt"
	"webook/ioc"
)

var thirdPartySet = wire.NewSet(
	InitLogger,
	InitDB, InitRedis, ioc.InitLocalMem)

var userSvcProvider = wire.NewSet(
	dao.NewGORMUserDAO,
	cache.NewRedisUserCache,
	repository.NewCacheUserRepository,
	service.NewCacheUserService)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		dao.NewArticleGORMDAO,
		dao.NewGORMUserDAO, cache.NewRedisUserCache, cache.NewLocalCodeCache, cache.NewArticleRedisCache,
		repository.NewCacheArticleRepository,
		repository.NewCacheUserRepository, repository.NewCacheCodeRepository,
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewArticleService,
		service.NewCacheUserService, service.NewCacheCodeService,
		jwt.NewRedisJWTHandler,
		web.NewArticleHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(thirdPartySet,
		interactiveSvcSet,
		userSvcProvider,
		cache.NewArticleRedisCache,
		repository.NewCacheArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdPartySet, interactiveSvcSet)
	return service2.NewInteractiveService(nil)
}
