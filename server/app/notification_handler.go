// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

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

// sseNotificationsHandler to stream notifications
// Example endpoint: Stream user's notifications
// @Summary Stream user's notifications
// @Description Stream user's notifications
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.Notification
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /notification [get]
func (a *App) sseNotificationsHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Flush the headers immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error().Msg("Streaming unsupported")
		internalServerError(w)
		return
	}

	// Sending notifications every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			notifications, err := a.db.GetNewNotifications(userID)
			if err != nil {
				log.Error().Err(err).Send()
				internalServerError(w)
				return
			}

			// Send each notification as a separate SSE message
			for _, notification := range notifications {
				if _, err := w.Write([]byte(notification.Msg)); err != nil {
					log.Error().Err(err).Send()
					internalServerError(w)
					return
				}
				flusher.Flush() // Ensure the event is sent immediately
			}

		case <-req.Context().Done():
			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

func internalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	object := struct {
		Error string `json:"err"`
	}{
		Error: "Internal server error",
	}

	if err := json.NewEncoder(w).Encode(object); err != nil {
		log.Error().Err(err).Msg("failed to encode return object")
	}
}
