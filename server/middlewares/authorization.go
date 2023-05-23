// Package middlewares for middleware between api and backend
package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/codescalers/cloud4students/internal"
)

// UserIDKey key saved in request context
type UserIDKey string

// Authorization to authorize users in requests
func Authorization(secret string, timeout int) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqToken := r.Header.Get("Authorization")
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) != 2 {
				writeErrResponse(r, w, http.StatusUnauthorized, "user is not authorized")
				return
			}
			reqToken = splitToken[1]

			claims, err := internal.ValidateJWTToken(reqToken, secret, timeout)
			if err != nil {
				writeErrResponse(r, w, http.StatusUnauthorized, "user is not authorized")
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey("UserID"), claims.UserID)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
