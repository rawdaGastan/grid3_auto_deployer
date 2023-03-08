// Package routes for API endpoints
package routes

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rawdaGastan/cloud4students/models"
)

func (r *Router) validateToken(admin bool, token, secret string) (models.Claims, error) {
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

	if time.Until(claims.ExpiresAt.Time) > time.Duration(r.config.Token.Timeout)*time.Minute {
		return models.Claims{}, fmt.Errorf("token '%s' is expired", token)
	}

	if admin {
		user, err := r.db.GetUserByID(claims.UserID)
		if err != nil {
			return models.Claims{}, err
		}

		if !user.Admin {
			return models.Claims{}, fmt.Errorf("user '%s' doesn't have an admin access", user.Name)
		}
	}

	return *claims, nil
}
