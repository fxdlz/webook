package web

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/internal/domain"
	"webook/internal/service"
	svcmocks "webook/internal/service/mocks"
	ijwt "webook/internal/web/jwt"
	"webook/pkg/logger"
)

type Req struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.ArticleService
		reqBody  string
		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并发表帖子",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Save(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 1,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: func() string {
				req := Req{
					Title:   "我的标题",
					Content: "我的内容",
				}
				res, _ := json.Marshal(req)
				return string(res)
			}(),
			wantCode: http.StatusOK,
			wantRes: Result{
				Data: float64(1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 准备Req和记录的 recorder
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish",
				bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			ctrl := gomock.NewController(t)
			svc := tc.mock(ctrl)
			hdl := NewArticleHandler(svc, logger.NewNopLogger())
			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", ijwt.UserClaims{
					Uid: 1,
				})
			})
			hdl.RegisterRoutes(server)
			// 执行
			server.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusOK {
				return
			}
			var res Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
