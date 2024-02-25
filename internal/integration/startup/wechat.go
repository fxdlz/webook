package startup

import (
	"os"
	"webook/internal/service/oauth2/wechat"
)

func InitWechatService() wechat.Service {
	appID, err := os.LookupEnv("WECHAT_APP_ID")
	if err {
		panic("找不到环境变量WECHAT_APP_ID")
	}
	appSecret, err := os.LookupEnv("WECHAT_APP_SECRET")
	if err {
		panic("找不到环境变量WECHAT_APP_SECRET")
	}
	return wechat.NewService(appID, appSecret)
}
