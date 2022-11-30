package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExpiryToken  = errors.New("Token has expired")
	ErrInvalidToken = errors.New("Invalid token")
)

// Payload contains the payload data for the token.
type Payload struct {
	ID        uuid.UUID `json:"id"`
	UserID    int64     `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(userID int64, ttl time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}
	return &Payload{
		ID:        tokenID,
		UserID:    userID,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(ttl),
	}, nil
}

func (p *Payload) Valid() error {

	if time.Now().After(p.ExpiredAt) {
		return ErrExpiryToken
	}

	return nil
}
