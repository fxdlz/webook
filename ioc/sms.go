package ioc

import (
	"webook/internal/service/sms"
	"webook/internal/service/sms/local"
)

func InitSMSService() sms.Service {
	return local.NewLocalSMSService()
}
