// Package routes for API endpoints
package routes

import (
	"net/http"

	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GenerateVoucherHandler generates a voucher by admin
func (r *Router) GetQuotaHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	quota, err := r.db.GetUserQuota(userID)
	if err == gorm.ErrRecordNotFound {
		writeNotFoundResponse(w, "user quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Quota is found", quota)
}
