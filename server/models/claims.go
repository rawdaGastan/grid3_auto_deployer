// package models for database models
package models

import (
	"github.com/golang-jwt/jwt/v4"
)

// Claims struct that will be encoded to a JWT.
type Claims struct {
	UserID string `json:"userID"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
