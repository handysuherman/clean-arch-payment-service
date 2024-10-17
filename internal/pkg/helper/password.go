package helper

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword securely hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hashed password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword compares a plaintext password with its hashed counterpart.
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
