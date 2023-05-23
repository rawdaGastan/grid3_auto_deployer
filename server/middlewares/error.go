// Package middlewares for middleware between api and backend
package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ErrorMsg holds errors
type ErrorMsg struct {
	Error string `json:"err"`
}

// writeErrResponse write error messages in api
func writeErrResponse(r *http.Request, w http.ResponseWriter, statusCode int, errStr string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(ErrorMsg{Error: errStr}); err != nil {
		log.Error().Err(err).Msg("failed to encode response object")
	}

	Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(statusCode)).Inc()
}
