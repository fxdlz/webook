package dao

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"reflect"
	"testing"
)

func TestGORMReqRetryDAO_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		sqlmock func(t *testing.T) *sql.DB
		ctx     context.Context
		id      string
		wantErr error
	}{
		{
			name: "删除成功",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				mock.ExpectExec("DELETE FROM `reqretries` WHERE `reqretries`.`id` = ?").WithArgs("212131").WillReturnResult(sqlmock.NewResult(1, 1))
				return db
			},
			ctx:     context.Background(),
			id:      "212131",
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.sqlmock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}),
				&gorm.Config{
					DisableAutomaticPing:   true,
					SkipDefaultTransaction: true,
				})
			dao := NewGORMReqRetryDAO(db)
			err = dao.Delete(tc.ctx, tc.id)
			assert.NoError(t, err)
		})
	}
}

func TestGORMReqRetryDAO_FindById(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx context.Context
		Id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Reqretry
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao := &GORMReqRetryDAO{
				db: tt.fields.db,
			}
			got, err := dao.FindById(tt.args.ctx, tt.args.Id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGORMReqRetryDAO_Insert(t *testing.T) {
	testCases := []struct {
		name     string
		sqlmock  func(t *testing.T) *sql.DB
		ctx      context.Context
		reqretry Reqretry
		wantErr  error
	}{
		{
			name: "插入成功",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.sqlmock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}),
				&gorm.Config{
					DisableAutomaticPing:   true,
					SkipDefaultTransaction: true,
				})
			dao := NewGORMReqRetryDAO(db)
			err = dao.Insert(tc.ctx, tc.reqretry)
			assert.NoError(t, err)
		})
	}
}

func TestGORMReqRetryDAO_Update(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx context.Context
		r   Reqretry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao := &GORMReqRetryDAO{
				db: tt.fields.db,
			}
			if err := dao.Update(tt.args.ctx, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewGORMReqRetryDAO(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name string
		args args
		want *GORMReqRetryDAO
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGORMReqRetryDAO(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGORMReqRetryDAO() = %v, want %v", got, tt.want)
			}
		})
	}
}
