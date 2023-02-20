package models

import "github.com/golang-jwt/jwt"

type Token struct {
	UserID uint64 `json:"userID" binding:"required"`
	Email  string `json:"email" binding:"required"`
	*jwt.StandardClaims
}
