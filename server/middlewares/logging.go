// Package middlewares for middleware between api and backend
package middlewares

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// LoggingMW logs all information of every request
func LoggingMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/// structured logs since u using zero logs

		log.Info().Timestamp().Str("method", r.Method).Str("uri", r.RequestURI).Msg("")
		h.ServeHTTP(w, r)
	})
}
