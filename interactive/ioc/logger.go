package ioc

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	cfg := zap.NewDevelopmentConfig()
	viper.UnmarshalKey("log", &cfg)
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
