//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/interactive/events"
	"webook/interactive/grpc"
	"webook/interactive/ioc"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitSaramaClient,
)

var interactiveSvcSet = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	cache2.NewInteractiveRedisCache,
	dao2.NewGORMInteractiveDAO,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		ioc.InitConsumers,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
