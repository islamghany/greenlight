package token

import (
	"time"
)

// Maker is an interface to manage the token life cycle.
type Maker interface {
	// CreateToken create a new token for a specific username and duration.
	CreateToken(userID int64, ttl time.Duration) (string, *Payload, error)

	// VerifyToken Checks if the provided token is valid.
	VerifyToken(token string) (*Payload, error)
}
