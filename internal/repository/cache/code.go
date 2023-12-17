package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/coocood/freecache"
	"github.com/redis/go-redis/v9"
	"strconv"
	"sync"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode        string
	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("发送太频繁")
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	key(biz, phone string) string
	Verify(ctx context.Context, biz, phone, expectedCode string) (bool, error)
}

type LocalCodeCache struct {
	lock  sync.Mutex
	cache *freecache.Cache
}

func NewLocalCodeCache(cache *freecache.Cache) CodeCache {
	return &LocalCodeCache{
		cache: cache,
	}
}

type RedisCodeCache struct {
	cmd redis.Cmdable
}

func NewRedisCodeCache(cmd redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		cmd: cmd,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.cmd.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case -2:
		return errors.New("验证码存在，但没有过期时间！")
	case -1:
		return ErrCodeSendTooMany
	default:
		return nil
	}
}

func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, expectedCode string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, expectedCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case -2:
		return false, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	default:
		return true, nil
	}
}

func (c *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	key := c.key(biz, phone)
	cntKey := key + ":cnt"
	c.lock.Lock()
	defer c.lock.Unlock()
	ttl, err := c.cache.TTL([]byte(key))
	if err != nil || ttl < 540 {
		err = c.cache.Set([]byte(key), []byte(code), 600)
		if err != nil {
			return err
		}
		err = c.cache.Set([]byte(cntKey), []byte("3"), 600)
		if err != nil {
			return err
		}
		return nil
	} else {
		return ErrCodeSendTooMany
	}

}

func (c *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (c *LocalCodeCache) Verify(ctx context.Context, biz, phone, expectedCode string) (bool, error) {
	key := c.key(biz, phone)
	cntKey := key + ":cnt"
	c.lock.Lock()
	defer c.lock.Unlock()
	cnt, err := c.cache.Get([]byte(cntKey))
	if err != nil {
		return false, err
	}
	n, err := strconv.Atoi(string(cnt))
	if err != nil {
		return false, err
	}
	if n <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	code, err := c.cache.Get([]byte(key))
	if err != nil {
		return false, nil
	}
	if string(code) == expectedCode {
		c.cache.Set([]byte(cntKey), []byte("0"), 0)
	} else {
		ttl, err := c.cache.TTL([]byte(key))
		if err != nil {
			return false, nil
		}
		c.cache.Set([]byte(cntKey), []byte(strconv.Itoa(n-1)), int(ttl))
		return false, nil
	}
	return true, nil
}
