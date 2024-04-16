package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	dao2 "webook/interactive/repository/dao"
	"webook/pkg/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DNS string `yaml:"dns"`
	}
	cfg := Config{
		DNS: "root:root@tcp(localhost:13316)/webook",
	}
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DNS), &gorm.Config{})
	if err != nil {
		panic("数据库连接初始化失败")
	}

	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{
		Key: "args",
		Val: args,
	})
}
