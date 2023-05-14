// Package routes for API endpoints
package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ListNotificationsHandler lists notifications for a user
func (r *Router) ListNotificationsHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	notifications, err := r.db.ListNotifications(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) || len(notifications) == 0 {
		writeMsgResponse(req, w, "You don't have any notifications yet", notifications)
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "You have notifications", notifications)
}

// UpdateNotificationsHandler updates notifications for a user
func (r *Router) UpdateNotificationsHandler(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read notification id")
		return
	}

	err = r.db.UpdateNotification(id, true)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Notifications is updated", "")
}
