package failover

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
	"webook/internal/service/sms"
	smsmocks "webook/internal/service/sms/mocks"
)

func TestCodeFailOverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name            string
		mock            func(ctrl *gomock.Controller) ([]sms.Service, repository.ReqRetryRepository)
		idx             int32
		retryNum        int8
		interval        int32
		reqErrCount     int32
		reqAllCount     int32
		wantErr         error
		wantReqErrCount int32
		wantReqAllCount int32
	}{
		{
			name: "请求成功",
			mock: func(ctrl *gomock.Controller) ([]sms.Service, repository.ReqRetryRepository) {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				repo := repomocks.NewMockReqRetryRepository(ctrl)
				//repo.EXPECT().FindById(gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0}, repo
			},
			idx:             0,
			retryNum:        5,
			interval:        5,
			reqErrCount:     0,
			reqAllCount:     0,
			wantErr:         nil,
			wantReqErrCount: 0,
			wantReqAllCount: 0,
		},
		{
			name: "请求失败",
			mock: func(ctrl *gomock.Controller) ([]sms.Service, repository.ReqRetryRepository) {
				data, _ := json.Marshal(SMSReq{
					tplId:   "tplId",
					args:    []string{"args"},
					numbers: []string{"args"},
				})
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("服务商崩溃"))
				repo := repomocks.NewMockReqRetryRepository(ctrl)
				repo.EXPECT().FindById(gomock.Any(), gomock.Any()).Return(domain.ReqRetry{
					Id:  "1",
					Req: string(data),
				}, nil).AnyTimes()
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				repo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return []sms.Service{svc0}, repo
			},
			idx:             0,
			retryNum:        5,
			interval:        5,
			reqErrCount:     10,
			reqAllCount:     20,
			wantErr:         errors.New("服务商崩溃"),
			wantReqErrCount: 1,
			wantReqAllCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := NewCodeFailOverSMSService(tc.mock(ctrl))
			err := svc.Send(context.Background(), "1", []string{"121"}, "wq")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantReqAllCount, svc.reqAllCount)
			assert.Equal(t, tc.wantReqErrCount, svc.reqErrCount)
		})
	}

}
