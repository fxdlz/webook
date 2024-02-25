package domain

type User struct {
	Id         int64
	Email      string
	Password   string
	Nickname   string
	Birthday   string
	Profile    string
	Phone      string
	WechatInfo WechatInfo
}
