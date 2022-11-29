package cache

import "github.com/go-redis/redis/v8"

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{
		rdb: rdb,
	}
}
