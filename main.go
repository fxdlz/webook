package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"webook/internal/middleware"
)

func main() {
	//tools.Mt.DeleteAll()
	//for i := 0; i < 1000; i++ {
	//	tools.Mt.InsertUserN(1000)
	//}
	server := InitWebServer()
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
