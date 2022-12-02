package api

import (
	"auth-service/db/cache"
	db "auth-service/db/sqlc"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func (server *Server) CacheUserbyID(userID int64, user map[string]interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	bUser, err := json.Marshal(user)
	if err != nil {
		return err
	}
	err = server.cache.RDB.Set(ctx, cache.UsersKey(userID), bUser, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) getReadyUser(id int64) (*db.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := server.cache.RDB.Get(ctx, cache.UsersKey(id)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	var user struct {
		User db.User `json:"user"`
	}
	if err != redis.Nil {
		err = json.Unmarshal([]byte(res), &user)
		if err != nil {
			return nil, err
		}
		return &user.User, nil
	}

	user.User, err = server.store.GetUserByID(ctx, id)

	if err != nil {
		return nil, err
	}
	server.background(func() {
		err := server.CacheUserbyID(id, envelope{"user": user.User})

		log.Println(err)
	})

	return &user.User, nil
}
