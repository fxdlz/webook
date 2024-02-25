package startup

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/internal/repository/dao"
	"webook/pkg/logger"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic("数据库连接初始化失败")
	}
	err = dao.InitTables(db)
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
