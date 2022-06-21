package data

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func SerializeGenres(g []string) string {
	values := ""

	for i, val := range g {
		if i == len(g)-1 {
			values = fmt.Sprint(values, val)
		} else {
			values = fmt.Sprint(values, val, ",")
		}
	}

	return values
}
func DeserializeGenres(val string) []string {

	return strings.Split(val, ",")
}

func SerializeRuntime(r int32) string {
	return fmt.Sprintf("%d mins", r)
}
func DeserializeRuntime(val string) int32 {

	parts := strings.Split(val, " ")

	i, _ := strconv.ParseInt(parts[0], 10, 32)
	return int32(i)
}

// *************** Helpers

func PipeSet(rdb *redis.Client, keys []string, values []map[string]interface{}, ttl time.Duration) error {
	if len(keys) != len(values) {
		return errors.New("the two arrays have different lengthes")
	}

	pipe := rdb.Pipeline()

	for i, key := range keys {

		err := pipe.HSet(ctx, key, values[i]).Err()
		if err != nil {
			return err
		}

		err = rdb.Expire(ctx, key, ttl).Err()
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

// func PipeGet(rdb *redis.Client, keys []string) (*Movie, error) {
// 	pipe := rdb.Pipeline()

// 	m := map[string]*redis.StringCmd{}
// 	for _, key := range keys {
// 		 pipe.HGetAll(ctx, key)
// 	}
// 	_, err := pipe.Exec(ctx)
// 	if err != nil {
// 		if err == redis.Nil {
// 			return nil, ErrRecordNotFound
// 		} else {
// 			return nil, err
// 		}
// 	}
// 	result2 := map[string]string{}
// 	for k, v := range m {
// 		res, err := v.Result()
// 		if res == "" || err == redis.Nil {
// 			result2[k] = ""
// 		} else {
// 			result2[k] = res
// 		}
// 	}
// 	return &result2, nil
// }
