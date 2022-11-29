package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return hashedPassword, nil
}

// CheckPassword checks if the provided password is correct or not
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
