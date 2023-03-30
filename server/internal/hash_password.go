// Package internal for internal details
package internal

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashAndSaltPassword hashes password of user
func HashAndSaltPassword(password string, salt string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(salt+password), bcrypt.MinCost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword checks if given password is same as hashed one
func VerifyPassword(hashedPassword string, password string, salt string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(salt+password))
	return err == nil
}
