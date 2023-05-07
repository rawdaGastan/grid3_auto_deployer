// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codescalers/cloud4students/internal"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UpdateMaintenanceInput struct for data needed when user update maintenance
type UpdateMaintenanceInput struct {
	ON bool `json:"on" binding:"required"`
}

// GetAllUsersHandler returns all users
func (r *Router) GetAllUsersHandler(w http.ResponseWriter, req *http.Request) {
	users, err := r.db.ListAllUsers()
	if err == gorm.ErrRecordNotFound || len(users) == 0 {
		writeMsgResponse(req, w, "Users are not found", users)
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Users are found", users)
}

// UpdateMaintenanceHandler updates maintenance flag
func (r *Router) UpdateMaintenanceHandler(w http.ResponseWriter, req *http.Request) {
	var input UpdateMaintenanceInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read maintenance update data")
		return
	}

	err = r.db.UpdateMaintenance(input.ON)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "maintenance is not found")
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Maintenance is updated successfully", "")
}

// GetMaintenanceHandler updates maintenance flag
func (r *Router) GetMaintenanceHandler(w http.ResponseWriter, req *http.Request) {
	maintenance, err := r.db.GetMaintenance()
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "maintenance is not found")
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, fmt.Sprintf("Maintenance is set with %v", maintenance.Active), maintenance)
}

// NotifyAdmins is used to notify admins that there are new vouchers requests
func (r *Router) NotifyAdmins() {
	ticker := time.NewTicker(time.Hour * time.Duration(r.config.NotifyAdminsIntervalHours))

	for range ticker.C {
		pending, err := r.db.GetAllPendingVouchers()
		if err != nil {
			log.Error().Err(err).Send()
		}

		if len(pending) > 0 {
			subject, body := internal.NotifyAdminsMailContent(len(pending))

			admins, err := r.db.ListAdmins()
			if err != nil {
				log.Error().Err(err).Send()
			}

			for _, admin := range admins {
				err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.SendGridKey, admin.Email, subject, body)
				if err != nil {
					log.Error().Err(err).Send()
				}
			}
		}
	}
}
