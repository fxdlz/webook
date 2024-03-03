package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"log"
	"webook/internal/middleware"
)

func main() {
	//tools.Mt.DeleteAll()
	//for i := 0; i < 1000; i++ {
	//	tools.Mt.InsertUserN(1000)
	//}
	initViper()
	//initViperRemote()
	//initViperWatch()
	app := InitApp()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	server := app.server
	server.Run(":8080")
}

func useSession(server *gin.Engine) {
	//基于cookie存储
	//store := cookie.NewStore([]byte("secret"))

	//基于内存存储
	store := memstore.NewStore([]byte(""))

	//store, err := redis.NewStore(16,
	//	"tcp",
	//	"localhost:6379",
	//	"",
	//	[]byte("tD1vD9qI5bF9fX8fH5nJ6yH4kM2dD6uM"),
	//	[]byte("lD4qN1mC6eH2kK9bF8fF3oF1zT8qM3pC"))
	//if err != nil {
	//	panic(err)
	//}

	server.Use(sessions.Sessions("ssid", store))

	login := middleware.LoginMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

func initViper() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("dev")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initViperWatch() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("dev")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println(viper.GetString("test.key"))
	})
	viper.WatchConfig()
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
