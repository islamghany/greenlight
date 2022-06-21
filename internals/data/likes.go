package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"islamghany.greenlight/internals/validator"
)

var (
	ErrDuplicateLike = errors.New("you already liked this movie")
)

type Like struct {
	ID      int64 `json:"id"`
	MovieID int64 `json:"movie_id"`
	UserID  int64 `json:"user_id"`
}

func ValidateLikeInput(v *validator.Validator, movieID int64) {

	v.Check(movieID != 0, "user_id", "must be provided")
	v.Check(movieID > 0, "user_id", "invalid movie id")
}

type LikeModel struct {
	DB  *sql.DB
	RDB *redis.Client
}

type LikeResponse struct {
	MovieID            int64 `json:"movie_id"`
	IsCurrentUserLiked int32 `json:"isCurrentUserLiked"`
	Likes              int64 `json:"likes"`
}

func (m LikeModel) GetMoiveLike(movieID, userID int64) (*LikeResponse, error) {
	query := `
	select count(*) as likes, sum(
			CASE
			WHEN l.user_id = $2 then 1
			else 0
			end
		) as currentUserLiked 
	from likes l 
	join movies m on m.id = l.movie_id
	where l.movie_id = $1
	group by l.id, m.id;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var l LikeResponse
	err := m.DB.QueryRowContext(ctx, query, movieID, userID).Scan(&l.Likes, &l.IsCurrentUserLiked)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	l.MovieID = movieID
	return &l, nil
}

func (m LikeModel) Insert(movieID, userID int64) error {
	query := `
		INSERT INTO likes (user_id,movie_id)
		VALUES ($1,$2);
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, movieID)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint"):
			return ErrDuplicateLike
		case strings.HasPrefix(err.Error(), "pq: insert or update on table"):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (m LikeModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM likes
		WHERE ID = $1;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return nil
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// func (m LikeModel) CacheAddLikeIfMovieExist(movieID int64) error {

// 	ok, err := m.RDB.HExists(ctx, MoviesKey(movieID), "id").Result()
// 	if err != nil {
// 		return err
// 	}
// 	if ok == true {
// 		err := m.RDB.HIncrBy(ctx, MoviesKey(movieID), "likes", 1).Err()
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
