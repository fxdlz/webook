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

type UserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (c *UserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func (c *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := c.key(id)
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	u := domain.User{}
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (c *UserCache) Set(ctx context.Context, u dao.User) error {
	key := c.key(u.Id)
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	err = c.cmd.Set(ctx, key, data, c.expiration).Err()
	return err
}

func NewUserCache(cmd redis.Cmdable) *UserCache {
	return &UserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}
