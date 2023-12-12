// Package app for c4s backend app
package app

import (
	"errors"
	"net/http"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetQuotaHandler gets quota
func (a *App) GetQuotaHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	quota, err := a.db.GetUserQuota(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user quota is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Quota is found",
		Data:    quota,
	}, Ok()
}
