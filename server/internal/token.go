// Package internal for internal details
package internal

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rawdaGastan/cloud4students/models"
)

// CreateJWT create token for user
func CreateJWT(u *models.User, secret string, timeout int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(timeout) * time.Minute)
	claims := &models.Claims{
		UserID: u.ID,
		Email:  u.Email,
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
