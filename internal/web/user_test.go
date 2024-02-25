package web

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/internal/domain"
	"webook/internal/service"
	svcmocks "webook/internal/service/mocks"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
		wantBody   string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "23424132@qq.com",
					Password: "213123#wsaads",
				}).Return(nil)
				return userSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email": "23424132@qq.com",
"password": "213123#wsaads",
"confirmPassword": "213123#wsaads"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "邮箱不正确",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email": "23424132",
"password": "213123#wsaads",
"confirmPassword": "213123#wsaads"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "邮箱不正确",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userSvc, codeSvc := tc.mock(ctrl)
			handler := NewUserHandler(userSvc, codeSvc, nil)
			server := gin.Default()
			handler.RegisterRoutes(server)
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}
