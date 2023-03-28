// Package internal for internal details
package internal

import (
	"time"

	"github.com/codescalers/cloud4students/models"
	"github.com/golang-jwt/jwt/v4"
)

// CreateJWT create token for user
func CreateJWT(userID string, email string, secret string, timeout int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(timeout) * time.Minute)
	claims := &models.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil

}
