package main

func main() {
	//tools.Mt.DeleteAll()
	//for i := 0; i < 1000; i++ {
	//	tools.Mt.InsertUserN(1000)
	//}
	server := InitWebServer()
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: config.Config.Redis.Addr,
	//})
	//codeSvc := initCodeSvc(redisClient)
	//db := initDB()
	//server := initWebServer(redisClient)
	//initUserHdl(db, redisClient, codeSvc, server)
	server.Run(":8080")
}

//func initUserHdl(db *gorm.DB, redisClient redis.Cmdable,
//	codeSvc *service.CodeService,
//	server *gin.Engine) {
//	dao := dao.NewUserDAO(db)
//	cache := cache.NewUserCache(redisClient)
//	repo := repository.NewUserRepository(dao, cache)
//	svc := service.NewUserService(repo)
//	u := web.NewUserHandler(svc, codeSvc)
//	u.RegisterRoutes(server)
//}

//func initCodeSvc(redisClient redis.Cmdable) *service.CodeService {
//	cache := cache.NewCodeCache(redisClient)
//	crepo := repository.NewCodeRepository(cache)
//	return service.NewCodeService(crepo, &local.Service{})
//}

//func initDB() *gorm.DB {
//	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
//	if err != nil {
//		panic("数据库连接初始化失败")
//	}
//	dao.InitTables(db)
//	return db
//}

//func initWebServer(redisClient redis.Cmdable) *gin.Engine {
//	server := gin.Default()
//	server.Use(cors.New(cors.Config{
//		AllowHeaders: []string{
//			"Content-Type",
//			"Accept",
//			"Authorization",
//			"Referer",
//			"Sec-Ch-Ua",
//			"Sec-Ch-Ua-Mobile",
//			"Sec-Ch-Ua-Platform",
//			"User-Agent",
//			"Cookie",
//		},
//		ExposeHeaders: []string{
//			"x-jwt-token",
//		},
//		AllowCredentials: true,
//		AllowOriginFunc: func(origin string) bool {
//			return strings.Contains(origin, "localhost")
//		},
//		MaxAge: 12 * time.Hour,
//	}))
//
//	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
//	//useSession(server)
//	useJWT(server)
//
//	return server
//}

//func useSession(server *gin.Engine) {
//	//基于cookie存储
//	//store := cookie.NewStore([]byte("secret"))
//
//	//基于内存存储
//	store := memstore.NewStore([]byte(""))
//
//	//store, err := redis.NewStore(16,
//	//	"tcp",
//	//	"localhost:6379",
//	//	"",
//	//	[]byte("tD1vD9qI5bF9fX8fH5nJ6yH4kM2dD6uM"),
//	//	[]byte("lD4qN1mC6eH2kK9bF8fF3oF1zT8qM3pC"))
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	server.Use(sessions.Sessions("ssid", store))
//
//	login := middleware.LoginMiddlewareBuilder{}
//	server.Use(login.CheckLogin())
//}

//func useJWT(server *gin.Engine) {
//	login := &middleware.LoginJWTMiddlewareBuilder{}
//	server.Use(login.CheckLogin())
//}
