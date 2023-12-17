package ioc

import "github.com/coocood/freecache"

func InitLocalMem() *freecache.Cache {
	return freecache.NewCache(100 * 1024 * 1024)
}
