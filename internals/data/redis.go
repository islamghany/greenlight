package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ctx = context.TODO()
)

// ************** USERS

func UsersKey(id int64) string {
	return fmt.Sprint("users#", id)
}

// ************** Movives
func MoviesKey(id int64) string {
	return fmt.Sprint("movies#", id)
}

// *************** Helpers

func PipeSet(rdb *redis.Client, keys []string, values [][]byte, ttl time.Duration) error {
	if len(keys) != len(values) {
		return errors.New("the two arrays have different lengthes")
	}
	pipe := rdb.Pipeline()

	for i, key := range keys {
		err := pipe.Set(ctx, key, values[i], ttl).Err()
		if err != nil {
			return err
		}
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func PipeGet(rdb *redis.Client, keys []string) (*map[string]string, error) {
	pipe := rdb.Pipeline()

	m := map[string]*redis.StringCmd{}
	for _, key := range keys {
		m[key] = pipe.Get(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		if err == redis.Nil {
			return nil, ErrRecordNotFound
		} else {
			return nil, err
		}
	}
	result2 := map[string]string{}
	for k, v := range m {
		res, err := v.Result()
		if res == "" || err == redis.Nil {
			result2[k] = ""
		} else {
			result2[k] = res
		}
	}
	return &result2, nil
}
