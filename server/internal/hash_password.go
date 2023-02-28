package internal

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes password of user 
func HashPassword(password string) (string, error) { //TODO: add salt of password (more encryption)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword checks if given password is same as hashed one
func VerifyPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
