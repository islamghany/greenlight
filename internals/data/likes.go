package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

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

func ValidateLikeInput(v *validator.Validator, userID, movieID int64) {
	v.Check(userID != 0, "user_id", "must be provided")
	v.Check(userID > 0, "user_id", "invalid user id")
	v.Check(movieID != 0, "user_id", "must be provided")
	v.Check(movieID > 0, "user_id", "invalid movie id")
}

type LikeModel struct {
	DB *sql.DB
}

func (m LikeModel) Insert(userID, movieID int64) error {
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
