// Package middlewares for middleware between api and backend
package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"gorm.io/gorm"
)

// UserIDKey key saved in request context
type UserIDKey string

// Authorization to authorize users in requests
func Authorization(db models.DB, secret string, timeout int) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqToken := r.Header.Get("Authorization")
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) != 2 {
				writeErrResponse(r, w, http.StatusUnauthorized, "user is not authorized, problem with token parsing")
				return
			}

			if strings.TrimSpace(splitToken[1]) == "" {
				writeErrResponse(r, w, http.StatusUnauthorized, "user is not authorized, problem with token parsing")
				return
			}
			reqToken = splitToken[1]

			claims, err := internal.ValidateJWTToken(reqToken, secret, timeout)
			if err != nil {
				writeErrResponse(r, w, http.StatusUnauthorized, "user is not authorized, token is not valid")
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey("UserID"), claims.UserID)

			user, err := db.GetUserByID(claims.UserID)
			if err == gorm.ErrRecordNotFound {
				writeErrResponse(r, w, http.StatusNotFound, "user: "+ user.Email +", is not found")
				return
			}
			if err != nil {
				writeErrResponse(r, w, http.StatusInternalServerError, "internal server error")
				return
			}
			if !user.Verified {
				writeErrResponse(r, w, http.StatusBadRequest, "email: "+ user.Email +" is not verified yet, please check the verification email in your inbox")
				return
			}

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
