package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const uniqueIndexErrno uint16 = 1062
		if me.Number == uniqueIndexErrno {
			return ErrDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	u := User{}
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

func (dao *UserDAO) Update(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Model(&u).Updates(map[string]interface{}{"nickname": u.Nickname, "birthday": u.Birthday, "profile": u.Profile}).Error
	return err
}

func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	u := User{}
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&u).Error
	return u, err
}

type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string
	Nickname string `gorm:"type=varchar(10)"`
	Birthday string
	Profile  string `gorm:"type=varchar(300)"`
	Ctime    int64
	Utime    int64
}
