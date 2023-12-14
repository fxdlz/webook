package dao

import (
	"context"
	"database/sql"
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
	var res User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&res).Error
	return res, err
}

func (dao *UserDAO) Update(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Model(&u).Updates(map[string]interface{}{"nickname": u.Nickname, "birthday": u.Birthday, "profile": u.Profile}).Error
	return err
}

func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var res User
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&res).Error
	return res, err
}

func (dao *UserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var res User
	err := dao.db.WithContext(ctx).Where("phone=?", phone).First(&res).Error
	return res, err
}

type User struct {
	Id       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Password string
	Nickname string `gorm:"type=varchar(10)"`
	Birthday string
	Profile  string         `gorm:"type=varchar(300)"`
	Phone    sql.NullString `gorm:"unique"`
	Ctime    int64
	Utime    int64
}
