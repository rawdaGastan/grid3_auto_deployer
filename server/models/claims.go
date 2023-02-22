package models

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// struct that will be encoded to a JWT.
type Claims struct {
	UserID uuid.UUID `json:"userID"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}
