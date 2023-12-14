package tools

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/rand"
	"webook/internal/repository/dao"
)

type MysqlTool struct {
	db *gorm.DB
}

var Mt *MysqlTool

func init() {
	Mt = &MysqlTool{
		db: initDB(),
	}
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic("数据库连接初始化失败")
	}
	return db
}

func (mt *MysqlTool) InsertUserN(limit int) {
	users := make([]dao.User, limit)
	for i := 0; i < limit; i++ {
		//passwd, err := bcrypt.GenerateFromPassword([]byte(randStr(int(rand.Int31()%5+6))+"a!"), bcrypt.DefaultCost)
		//if err != nil {
		//	panic("加密异常")
		//}
		h := md5.New()
		h.Write([]byte(randStr(int(rand.Int31()%5+6)) + "a!"))
		passwd := fmt.Sprintf("%x", md5.Sum(nil))

		users[i] = dao.User{
			Email: sql.NullString{
				String: randStr(int(rand.Int31()%5+8)) + "@" + "qq.com",
				Valid:  true,
			},
			Password: string(passwd[:]),
		}
	}
	mt.db.Create(&users)
}

func (mt *MysqlTool) DeleteAll() {
	mt.db.Where("1=1").Delete(&dao.User{})
}
