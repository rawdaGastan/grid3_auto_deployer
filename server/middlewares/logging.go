// Package middlewares for middleware between api and backend
package middlewares

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// LoggingMW logs all information of every request
func LoggingMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Timestamp().Str("method", r.Method).Str("uri", r.RequestURI).Send()
		h.ServeHTTP(w, r)
	})
}
