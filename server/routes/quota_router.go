// Package routes for API endpoints
package routes

import (
	"net/http"

	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetQuotaHandler gets quota
func (r *Router) GetQuotaHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	quota, err := r.db.GetUserQuota(userID)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(w, http.StatusNotFound, "user quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Quota is found", quota)
}
