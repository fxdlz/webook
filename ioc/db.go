package ioc

import (
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
	"webook/internal/repository/dao"
	"webook/pkg/gormx"
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
	db, err := gorm.Open(mysql.Open(cfg.DNS), &gorm.Config{
		//Logger: glogger.New(gormLoggerFunc(l.Info), glogger.Config{
		//	SlowThreshold: 0,
		//	LogLevel:      glogger.Info,
		//}),
	})
	if err != nil {
		panic("数据库连接初始化失败")
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"threads_running"},
			},
		},
	}))

	if err != nil {
		panic(err)
	}

	cb := gormx.NewCallBacks(prometheus2.SummaryOpts{
		Namespace: "fxlz",
		Subsystem: "webook",
		Name:      "gorm_db",
		Help:      "统计GORM的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})

	err = db.Use(cb)

	if err != nil {
		panic(err)
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
