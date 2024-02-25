package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"
)

func TestCacheArticleRepository_SyncV1(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO)
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "新建同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				authorDao := daomocks.NewMockArticleAuthorDAO(ctrl)
				readerDao := daomocks.NewMockArticleReaderDAO(ctrl)
				authorDao.EXPECT().Create(gomock.Any(), dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 1,
				}).Return(int64(3), nil)
				readerDao.EXPECT().Upsert(gomock.Any(), dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 1,
				}).Return(nil)
				return authorDao, readerDao
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 1,
				},
			},
			wantId: 3,
		},
		{
			name: "修改同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				authorDao := daomocks.NewMockArticleAuthorDAO(ctrl)
				readerDao := daomocks.NewMockArticleReaderDAO(ctrl)
				authorDao.EXPECT().Update(gomock.Any(), dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 1,
				}).Return(nil)
				readerDao.EXPECT().Upsert(gomock.Any(), dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 1,
				}).Return(nil)
				return authorDao, readerDao
			},
			art: domain.Article{
				Id:      3,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 1,
				},
			},
			wantId: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authorDAO, readerDAO := tc.mock(ctrl)
			repo := NewCacheArticleRepositoryV2(readerDAO, authorDAO)
			id, err := repo.SyncV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
