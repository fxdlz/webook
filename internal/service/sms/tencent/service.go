package tencent

import (
	"context"
	"fmt"
	ekit "github.com/gotomicro/ekit"
	"github.com/gotomicro/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
}

func NewService(c *sms.Client, appId string, signName string) *Service {
	return &Service{
		client:   c,
		appId:    &appId,
		signName: &signName,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	request := sms.NewSendSmsRequest()
	request.SmsSdkAppId = s.appId
	request.SignName = s.signName
	request.TemplateId = ekit.ToPtr[string](tplId)
	request.TemplateParamSet = toStringPtrSlice(args)
	request.PhoneNumberSet = toStringPtrSlice(numbers)
	// 通过client对象调用想要访问的接口，需要传入请求对象
	response, err := s.client.SendSms(request)
	// 处理异常
	if err != nil {
		return err
	}
	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			continue
		}
		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送失败，code: %s, message: %s\n", *status.Code, *status.Message)
		}
	}
	return nil
}

func toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
