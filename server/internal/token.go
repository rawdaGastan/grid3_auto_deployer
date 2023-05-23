// Package internal for internal details
package internal

import (
	"fmt"
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

// ValidateJWTToken validates if the token is valid
func ValidateJWTToken(token, secret string, timeout int) (models.Claims, error) {
	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return models.Claims{}, err
	}
	if !tkn.Valid {
		return models.Claims{}, fmt.Errorf("token '%s' is invalid", token)
	}

	if time.Until(claims.ExpiresAt.Time) > time.Duration(timeout)*time.Minute {
		return models.Claims{}, fmt.Errorf("token '%s' is expired", token)
	}

	return *claims, nil
}
