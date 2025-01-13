// Package app for c4s backend app
package app

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
// Example endpoint: Lists user's notifications
// @Summary Lists user's notifications
// @Description Lists user's notifications
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.Notification
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /notification [get]
func (a *App) ListNotificationsHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	notifications, err := a.db.ListNotifications(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) || len(notifications) == 0 {
		return ResponseMsg{
			Message: "You don't have any notifications yet",
			Data:    notifications,
		}, Ok()
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "You have notifications",
		Data:    notifications,
	}, Ok()
}

// UpdateNotificationsHandler updates notifications for a user
// Example endpoint: Set user's notifications as seen
// @Summary Set user's notifications as seen
// @Description Set user's notifications as seen
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /notification/{id} [put]
func (a *App) UpdateNotificationsHandler(req *http.Request) (interface{}, Response) {
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read notification id"))
	}

	err = a.db.UpdateNotification(id, true)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Notifications are updated",
		Data:    nil,
	}, Ok()
}

// SeenNotificationsHandler updates notifications for a user to be seen
// Example endpoint: Set user's notifications as seen
// @Summary Set user's notifications as seen
// @Description Set user's notifications as seen
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /notification [put]
func (a *App) SeenNotificationsHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	err := a.db.UpdateUserNotification(userID, true)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Notifications are seen",
		Data:    nil,
	}, Ok()
}
