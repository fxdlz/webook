package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	key(id int64) string
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u dao.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (c *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func (c *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := c.key(id)
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	u := domain.User{}
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (c *RedisUserCache) Set(ctx context.Context, u dao.User) error {
	key := c.key(u.Id)
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	err = c.cmd.Set(ctx, key, data, c.expiration).Err()
	return err
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}
