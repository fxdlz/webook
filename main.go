package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
	"webook/internal/middleware"
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/pkg/ginx/middleware/ratelimit"
)

func main() {
	db := initDB()
	server := initWebServer()
	initUserHdl(db, server)
	server.Run(":8080")
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	dao := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(dao)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	u.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(192.168.1.17:13316)/webook"))
	if err != nil {
		panic("数据库连接初始化失败")
	}
	dao.InitTables(db)
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowHeaders: []string{
			"Content-Type",
			"Accept",
			"Authorization",
			"Referer",
			"Sec-Ch-Ua",
			"Sec-Ch-Ua-Mobile",
			"Sec-Ch-Ua-Platform",
			"User-Agent",
			"Cookie",
		},
		ExposeHeaders: []string{
			"x-jwt-token",
		},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return strings.Contains(origin, "localhost")
		},
		MaxAge: 12 * time.Hour,
	}))

	redisClient := redis.NewClient(&redis.Options{
		Addr: "192.168.1.17:6379",
	})
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 1).Build())
	//useSession(server)
	useJWT(server)

	return server
}

func useSession(server *gin.Engine) {
	//基于cookie存储
	//store := cookie.NewStore([]byte("secret"))

	//基于内存存储
	store := memstore.NewStore([]byte(""))

	//store, err := redis.NewStore(16,
	//	"tcp",
	//	"192.168.1.17:6379",
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

func useJWT(server *gin.Engine) {
	login := middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}
