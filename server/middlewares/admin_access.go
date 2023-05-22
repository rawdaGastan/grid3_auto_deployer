// Package middlewares for middleware between api and backend
package middlewares

import (
	"fmt"
	"net/http"

	"github.com/codescalers/cloud4students/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// AdminAccess to authorize admins in requests
func AdminAccess(db models.DB) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(UserIDKey("UserID")).(string)
			user, err := db.GetUserByID(userID)
			if err == gorm.ErrRecordNotFound {
				writeErrResponse(r, w, http.StatusNotFound, "user is not found")
				return
			}
			if err != nil {
				log.Error().Err(err).Send()
				writeErrResponse(r, w, http.StatusInternalServerError, "something went wrong")
				return
			}

			if !user.Admin {
				writeErrResponse(r, w, http.StatusUnauthorized, fmt.Sprintf("user '%s' doesn't have an admin access", user.Name))
				return
			}
		})
	}
}
