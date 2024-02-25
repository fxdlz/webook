package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/internal/integration/startup"
	"webook/internal/repository/dao"
	ijwt "webook/internal/web/jwt"
)

type ArticleMongoDBHandlerSuite struct {
	suite.Suite
	db      *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
	server  *gin.Engine
}

func (s *ArticleMongoDBHandlerSuite) SetupSuite() {
	s.db = startup.InitMongoDB()
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	col := s.db.Collection("articles")
	liveCol := s.db.Collection("published_articles")
	hdl := startup.InitArticleHandler(dao.NewMongoDBArticleDAO(node, col, liveCol))

	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", ijwt.UserClaims{
			Uid: 1,
		})
	})
	hdl.RegisterRoutes(server)
	s.server = server
}

func (s *ArticleMongoDBHandlerSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.liveCol.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *ArticleMongoDBHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name       string
		before     func(t *testing.T)
		after      func(t *testing.T)
		art        Article
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name:   "新建帖子",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				//验证保存到了数据库中
				var art = dao.Article{}
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 1}}).Decode(&art)
				assert.NoError(t, err)

				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)

				art.Ctime = 0
				art.Utime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "这是标题",
					Content:  "这是内容",
					AuthorId: 1,
					Status:   1,
				}, art)
			},
			art: Article{
				Title:   "这是标题",
				Content: "这是内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		//{
		//	name: "修改帖子",
		//	before: func(t *testing.T) {
		//		err := s.db.Create(dao.Article{
		//			Id:       2,
		//			Title:    "我的标题",
		//			Content:  "我的内容",
		//			Status:   2,
		//			AuthorId: 1,
		//			Ctime:    456,
		//			Utime:    789,
		//		}).Error
		//		assert.NoError(t, err)
		//	},
		//	after: func(t *testing.T) {
		//		//验证保存到了数据库中
		//		var art = dao.Article{}
		//		err := s.db.Where("id=?", 2).First(&art).Error
		//		assert.NoError(t, err)
		//		assert.True(t, art.Utime > 789)
		//		art.Utime = 0
		//		assert.Equal(t, dao.Article{
		//			Id:       2,
		//			Title:    "新的标题",
		//			Content:  "新的内容",
		//			Status:   1,
		//			AuthorId: 1,
		//			Ctime:    456,
		//		}, art)
		//	},
		//	art: Article{
		//		Id:      2,
		//		Title:   "新的标题",
		//		Content: "新的内容",
		//	},
		//	wantCode: http.StatusOK,
		//	wantResult: Result[int64]{
		//		Data: 2,
		//	},
		//},
		//{
		//	name: "修改帖子-别人的帖子",
		//	before: func(t *testing.T) {
		//		err := s.db.Create(dao.Article{
		//			Id:       3,
		//			Title:    "我的标题",
		//			Content:  "我的内容",
		//			Status:   1,
		//			AuthorId: 2,
		//			Ctime:    456,
		//			Utime:    789,
		//		}).Error
		//		assert.NoError(t, err)
		//	},
		//	after: func(t *testing.T) {
		//		//验证保存到了数据库中
		//		var art = dao.Article{}
		//		err := s.db.Where("id=?", 3).First(&art).Error
		//		assert.NoError(t, err)
		//		assert.Equal(t, dao.Article{
		//			Id:       3,
		//			Title:    "我的标题",
		//			Content:  "我的内容",
		//			Status:   1,
		//			AuthorId: 2,
		//			Ctime:    456,
		//			Utime:    789,
		//		}, art)
		//	},
		//	art: Article{
		//		Id:      3,
		//		Title:   "新的标题",
		//		Content: "新的内容",
		//	},
		//	wantCode: http.StatusOK,
		//	wantResult: Result[int64]{
		//		Data: 3,
		//	},
		//},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			// 准备Req和记录的 recorder
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit",
				bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()

			// 执行
			s.server.ServeHTTP(recorder, req)
			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			if tc.wantResult.Data > 0 {
				assert.True(t, res.Data > 0)
			}
		})
	}

}

func TestMongoDBArticleHandler(t *testing.T) {
	suite.Run(t, &ArticleMongoDBHandlerSuite{})
}
