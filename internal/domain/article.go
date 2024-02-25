package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Status  ArticleStatus
	Author  Author
	Ctime   time.Time
	Utime   time.Time
}

type ArticleStatus uint8

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s Article) Abstract() string {
	str := []rune(s.Content)
	if len(str) > 128 {
		str = str[:128]
	}
	return string(str)
}

const (
	ArticleStatusUnknown = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

type Author struct {
	Id   int64
	Name string
}
