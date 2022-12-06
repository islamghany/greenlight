package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ctx             = context.TODO()
	ErrNoCacheFound = errors.New("find: not found")
)

func UsersKey(id int64) string {
	return fmt.Sprint("users#", id)
}

type Cache struct {
	RDB *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{
		RDB: rdb,
	}
}

func Set(rdb *redis.Client, key, value string, ttl time.Duration) error {

	return rdb.Set(ctx, key, value, ttl).Err()

}

func Get(rdb *redis.Client, key string) (*string, error) {
	res, err := rdb.Get(ctx, key).Result()

	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("find: redis error: %w", err)
	}

	if err == redis.Nil {
		return nil, fmt.Errorf("find: not found")
	}
	return &res, nil
}
