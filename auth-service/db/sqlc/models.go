// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package db

import (
	"time"
)

type User struct {
	ID                int64     `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	HashedPassword    []byte    `json:"hashed_password"`
	Activated         bool      `json:"activated"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	Version           int32     `json:"version"`
}
