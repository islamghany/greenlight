package data

import (
	"database/sql"
	"errors"

	"github.com/go-redis/redis/v8"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Movies      MovieModel
	Permissions PermissionModel
	Likes       LikeModel
}

func NewModels(db *sql.DB, rdb *redis.Client) Models {
	return Models{
		Movies:      MovieModel{DB: db, RDB: rdb},
		Permissions: PermissionModel{DB: db},
		Likes:       LikeModel{DB: db, RDB: rdb},
	}
}
