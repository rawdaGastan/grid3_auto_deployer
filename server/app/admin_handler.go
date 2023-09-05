// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

// AdminAnnouncement struct for data needed when admin sends new announcement
type AdminAnnouncement struct {
	Subject string `json:"subject"  binding:"required"`
	Body    string `json:"announcement" binding:"required"`
}

// UpdateMaintenanceInput struct for data needed when user update maintenance
type UpdateMaintenanceInput struct {
	ON bool `json:"on" binding:"required"`
}

// GetAllUsersHandler returns all users
func (a *App) GetAllUsersHandler(req *http.Request) (interface{}, Response) {
	users, err := a.db.ListAllUsers()
	if err == gorm.ErrRecordNotFound || len(users) == 0 {
		return ResponseMsg{
			Message: "Users are not found",
			Data:    users,
		}, Ok()
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Users are found",
		Data:    users,
	}, Ok()
}

// GetDlsCountHandler returns deployments count
func (a *App) GetDlsCountHandler(req *http.Request) (interface{}, Response) {
	count, err := a.db.CountAllDeployments()
	if err == gorm.ErrRecordNotFound {
		return ResponseMsg{
			Message: "Deployments count is not found",
			Data:    count,
		}, Ok()
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Deployments count is returned successfully",
		Data:    count,
	}, Ok()
}

// GetBalanceHandler return account balance information
func (a *App) GetBalanceHandler(req *http.Request) (interface{}, Response) {
	balance, err := a.deployer.GetBalance()
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Balance is found",
		Data:    balance,
	}, Ok()
}

// UpdateMaintenanceHandler updates maintenance flag
func (a *App) UpdateMaintenanceHandler(req *http.Request) (interface{}, Response) {
	var input UpdateMaintenanceInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read maintenance update data"))
	}

	err = a.db.UpdateMaintenance(input.ON)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("maintenance is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Maintenance is updated successfully",
		Data:    nil,
	}, Ok()
}

// GetMaintenanceHandler updates maintenance flag
func (a *App) GetMaintenanceHandler(req *http.Request) (interface{}, Response) {
	maintenance, err := a.db.GetMaintenance()
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("maintenance is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: fmt.Sprintf("Maintenance is set with %v", maintenance.Active),
		Data:    maintenance,
	}, Ok()
}

// NotifyAdmins is used to notify admins that there are new vouchers requests
func (a *App) notifyAdmins() {
	ticker := time.NewTicker(time.Hour * time.Duration(a.config.NotifyAdminsIntervalHours))

	for range ticker.C {
		// get admins
		admins, err := a.db.ListAdmins()
		if err != nil {
			log.Error().Err(err).Send()
		}

		// check pending voucher requests
		pending, err := a.db.GetAllPendingVouchers()
		if err != nil {
			log.Error().Err(err).Send()
		}

		if len(pending) > 0 {
			subject, body := internal.NotifyAdminsMailContent(len(pending), a.config.Server.Host)

			for _, admin := range admins {
				err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, admin.Email, subject, body)
				if err != nil {
					log.Error().Err(err).Send()
				}
			}
		}

		// check account balance
		balance, err := a.deployer.GetBalance()
		if err != nil {
			log.Error().Err(err).Send()
		}

		if int(balance) < a.config.BalanceThreshold {
			subject, body := internal.NotifyAdminsMailLowBalanceContent(balance, a.config.Server.Host)

			for _, admin := range admins {
				err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, admin.Email, subject, body)
				if err != nil {
					log.Error().Err(err).Send()
				}
			}
		}
	}
}

// CreateNewAnnouncement creates a new admin announcement and sends it to all users as an email and notification
func (a *App) CreateNewAnnouncement(req *http.Request) (interface{}, Response) {
	var adminAnnouncement AdminAnnouncement
	err := json.NewDecoder(req.Body).Decode(&adminAnnouncement)

	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read announcement data"))
	}

	err = validator.Validate(adminAnnouncement)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid announcement data"))
	}

	users, err := a.db.ListAllUsers()
	if err == gorm.ErrRecordNotFound || len(users) == 0 {
		return ResponseMsg{
			Message: "Users are not found",
			Data:    users,
		}, Ok()
	}
	subject, body := internal.AdminAnnouncementMailContent(adminAnnouncement.Subject, adminAnnouncement.Body, a.config.Server.Host)
	for _, user := range users {
		err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, user.Email, subject, body)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
		notification := models.Notification{UserID: user.UserID, Msg: body}
		err = a.db.CreateNotification(&notification)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}
	return ResponseMsg{
		Message: "new announcement created successfully",
	}, Created()
}
