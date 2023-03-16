// Package routes for API endpoints
package routes

import (
	"net/http"

	"github.com/rawdaGastan/cloud4students/middlewares"
)

// GenerateVoucherHandler generates a voucher by admin
func (r *Router) GetQuotaHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	quota, err := r.db.GetUserQuota(userID)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}

	writeMsgResponse(w, "Quota is found", quota)
}
