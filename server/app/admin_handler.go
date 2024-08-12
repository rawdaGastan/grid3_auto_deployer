// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
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

// SetAdminInput struct for setting users as admins
type SetAdminInput struct {
	Email string `json:"email" binding:"required"`
	Admin bool   `json:"admin" binding:"required"`
}

// UpdateNextLaunchInput struct for data needed when updating next launch state
type UpdateNextLaunchInput struct {
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

func (a *App) ResetUsersQuota(req *http.Request) (interface{}, Response) {
	users, err := a.db.ListAllUsers()
	if err == gorm.ErrRecordNotFound || len(users) == 0 {
		return ResponseMsg{
			Message: "Users are not found",
		}, Ok()
	}

	for _, user := range users {
		err = a.db.UpdateUserQuota(user.UserID, 0, 0)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	return ResponseMsg{
		Message: "Quota is reset successfully",
	}, Ok()
}

// DeleteAllDeployments deletes all deployments
func (a *App) DeleteAllDeployments(req *http.Request) (interface{}, Response) {
	users, err := a.db.ListAllUsers()
	if err == gorm.ErrRecordNotFound || len(users) == 0 {
		return ResponseMsg{
			Message: "Users are not found",
		}, Ok()
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, user := range users {
		// vms
		vms, err := a.db.GetAllVms(user.UserID)
		if err == gorm.ErrRecordNotFound || len(vms) == 0 {
			log.Error().Err(err).Str("userID", user.UserID).Msg("Virtual machines are not found")
			continue
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		for _, vm := range vms {
			err = a.deployer.CancelDeployment(vm.ContractID, vm.NetworkContractID, "vm", vm.Name)
			if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
				log.Error().Err(err).Send()
				return nil, InternalServerError(errors.New(internalServerErrorMsg))
			}
		}

		err = a.db.DeleteAllVms(user.UserID)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		// k8s clusters
		clusters, err := a.db.GetAllK8s(user.UserID)
		if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
			log.Error().Err(err).Str("userID", user.UserID).Msg("Kubernetes clusters are not found")
			continue
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		for _, cluster := range clusters {
			err = a.deployer.CancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract), "k8s", cluster.Master.Name)
			if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
				log.Error().Err(err).Send()
				return nil, InternalServerError(errors.New(internalServerErrorMsg))
			}
		}

		err = a.db.DeleteAllK8s(user.UserID)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	return ResponseMsg{
		Message: "Deployments are deleted successfully",
	}, Ok()
}

// ListDeployments lists all deployments
func (a *App) ListDeployments(req *http.Request) (interface{}, Response) {
	users, err := a.db.ListAllUsers()
	if err == gorm.ErrRecordNotFound || len(users) == 0 {
		return ResponseMsg{
			Message: "Users are not found",
		}, Ok()
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	var allVMs []models.VM
	var allClusters []models.K8sCluster

	for _, user := range users {
		// vms
		vms, err := a.db.GetAllVms(user.UserID)
		if err == gorm.ErrRecordNotFound || len(vms) == 0 {
			log.Error().Err(err).Str("userID", user.UserID).Msg("Virtual machines are not found")
			continue
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		allVMs = append(allVMs, vms...)

		// k8s clusters
		clusters, err := a.db.GetAllK8s(user.UserID)
		if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
			log.Error().Err(err).Str("userID", user.UserID).Msg("Kubernetes clusters are not found")
			continue
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		allClusters = append(allClusters, clusters...)
	}

	return ResponseMsg{
		Message: "Deployments are listed successfully",
		Data:    map[string]interface{}{"vms": allVMs, "k8s": allClusters},
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

// SetAdmin sets a user as an admin
func (a *App) SetAdmin(req *http.Request) (interface{}, Response) {
	input := SetAdminInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read data"))
	}

	user, err := a.db.GetUserByEmail(input.Email)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if user.Admin && input.Admin {
		return ResponseMsg{
			Message: "User is already an admin",
		}, Ok()
	}

	if !user.Admin && !input.Admin {
		return ResponseMsg{
			Message: "User is not already an admin",
		}, Ok()
	}

	err = a.db.UpdateAdminUserByID(user.ID.String(), input.Admin)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "User is updated successfully",
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

// CreateNewAnnouncement creates a new administrator announcement and sends it to all users as an email and notification
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

	for _, user := range users {
		subject, body := internal.AdminAnnouncementMailContent(adminAnnouncement.Subject, adminAnnouncement.Body, a.config.Server.Host, user.Name)

		err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, user.Email, subject, body)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		notification := models.Notification{UserID: user.UserID, Msg: fmt.Sprintf("Announcement: %s", adminAnnouncement.Body)}
		err = a.db.CreateNotification(&notification)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	return ResponseMsg{
		Message: "new announcement is sent successfully",
	}, Created()
}

// UpdateNextLaunchHandler updates next launch flag
func (a *App) UpdateNextLaunchHandler(req *http.Request) (interface{}, Response) {
	var input UpdateNextLaunchInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read NextLaunch update data"))
	}

	err = a.db.UpdateNextLaunch(input.ON)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("next launch is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Next Launch is updated successfully",
		Data:    nil,
	}, Ok()
}

// GetNextLaunchHandler returns next launch state
func (a *App) GetNextLaunchHandler(req *http.Request) (interface{}, Response) {
	nextlaunch, err := a.db.GetNextLaunch()

	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("next launch is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: fmt.Sprintf("Next Launch is Launched with state: %v", nextlaunch.Launched),
		Data:    nextlaunch,
	}, Ok()
}
