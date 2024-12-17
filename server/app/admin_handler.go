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

// EmailUser struct for data needed when admin sends new email to a user
type EmailUser struct {
	Subject string `json:"subject"  binding:"required"`
	Body    string `json:"body" binding:"required"`
	Email   string `json:"email" binding:"required" validate:"mail"`
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
	Launched bool `json:"launched" binding:"required"`
}

// SetPricesInput struct for setting prices as admins
type SetPricesInput struct {
	Small    float64 `json:"small"`
	Medium   float64 `json:"medium"`
	Large    float64 `json:"large"`
	PublicIP float64 `json:"public_ip"`
}

type ListDeploymentsResponse struct {
	VMs []models.VM         `json:"vms"`
	K8S []models.K8sCluster `json:"k8s"`
}

// GetAllUsersHandler returns all users
// Example endpoint: List all users
// @Summary List all users
// @Description List all users in the system
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.User
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/all [get]
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

// GetAllInvoicesHandler returns all invoices
// Example endpoint: List all invoices
// @Summary List all invoices
// @Description List all invoices in the system
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.Invoice
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /invoice/all [get]
func (a *App) GetAllInvoicesHandler(req *http.Request) (interface{}, Response) {
	invoices, err := a.db.ListInvoices()
	if err == gorm.ErrRecordNotFound || len(invoices) == 0 {
		return ResponseMsg{
			Message: "Invoices are not found",
			Data:    invoices,
		}, Ok()
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Invoices are found",
		Data:    invoices,
	}, Ok()
}

// SetPricesHandler set prices for vms and public ip
// Example endpoint: Set prices
// @Summary Set prices
// @Description Set vms and public ips prices prices
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param prices body SetPricesInput true "Prices to be set"
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Router /set_prices [put]
func (a *App) SetPricesHandler(req *http.Request) (interface{}, Response) {
	var input SetPricesInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read input data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid input data"))
	}

	if input.Small != 0 {
		a.config.PricesPerMonth.SmallVM = input.Small
	}

	if input.Medium != 0 {
		a.config.PricesPerMonth.MediumVM = input.Medium
	}

	if input.Large != 0 {
		a.config.PricesPerMonth.LargeVM = input.Large
	}

	if input.PublicIP != 0 {
		a.config.PricesPerMonth.PublicIP = input.PublicIP
	}

	return ResponseMsg{
		Message: "New prices are set",
		Data:    nil,
	}, Ok()
}

// GetDlsCountHandler returns deployments count
// Example endpoint: Get users' deployments count
// @Summary Get users' deployments count
// @Description Get users' deployments count in the system
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} models.DeploymentsCount
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /deployments/count [get]
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
// Example endpoint: Get main TF account balance
// @Summary Get main TF account balance
// @Description Get main TF account balance
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} float64
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /balance [get]
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

// DeleteAllDeploymentsHandler deletes all users' deployments
// Example endpoint: Deletes all users' deployments
// @Summary Deletes all users' deployments
// @Description Deletes all users' deployments
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /deployments [delete]
func (a *App) DeleteAllDeploymentsHandler(req *http.Request) (interface{}, Response) {
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
		vms, err := a.db.GetAllVms(user.ID.String())
		if err == gorm.ErrRecordNotFound || len(vms) == 0 {
			log.Error().Err(err).Str("userID", user.ID.String()).Msg("Virtual machines are not found")
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

		err = a.db.DeleteAllVms(user.ID.String())
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		// k8s clusters
		clusters, err := a.db.GetAllK8s(user.ID.String())
		if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
			log.Error().Err(err).Str("userID", user.ID.String()).Msg("Kubernetes clusters are not found")
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

		err = a.db.DeleteAllK8s(user.ID.String())
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	return ResponseMsg{
		Message: "Deployments are deleted successfully",
	}, Ok()
}

// ListDeploymentsHandler lists all users' deployments
// Example endpoint: List all users' deployments
// @Summary List all users' deployments
// @Description List all users' deployments
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} ListDeploymentsResponse
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /deployments [get]
func (a *App) ListDeploymentsHandler(req *http.Request) (interface{}, Response) {
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
		vms, err := a.db.GetAllVms(user.ID.String())
		if err == gorm.ErrRecordNotFound || len(vms) == 0 {
			log.Error().Err(err).Str("userID", user.ID.String()).Msg("Virtual machines are not found")
			continue
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		allVMs = append(allVMs, vms...)

		// k8s clusters
		clusters, err := a.db.GetAllK8s(user.ID.String())
		if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
			log.Error().Err(err).Str("userID", user.ID.String()).Msg("Kubernetes clusters are not found")
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
		Data:    ListDeploymentsResponse{VMs: allVMs, K8S: allClusters},
	}, Ok()
}

// UpdateMaintenanceHandler updates maintenance flag
// Example endpoint: Updates maintenance flag
// @Summary Updates maintenance flag
// @Description Updates maintenance flag
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param maintenance body UpdateMaintenanceInput true "Maintenance value to be set"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /maintenance [put]
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

// SetAdminHandler sets a user as an admin
// Example endpoint: Sets a user as an admin
// @Summary Sets a user as an admin
// @Description Sets a user as an admin
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param setAdmin body SetAdminInput true "User to be set as admin"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /set_admin [put]
func (a *App) SetAdminHandler(req *http.Request) (interface{}, Response) {
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

// CreateNewAnnouncementHandler creates a new administrator announcement and sends it to all users as an email and notification
// Example endpoint: Creates a new administrator announcement and sends it to all users as an email and notification
// @Summary Creates a new administrator announcement and sends it to all users as an email and notification
// @Description Creates a new administrator announcement and sends it to all users as an email and notification
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param announcement body AdminAnnouncement true "announcement to be created"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /announcement [post]
func (a *App) CreateNewAnnouncementHandler(req *http.Request) (interface{}, Response) {
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
		subject, body := internal.AdminAnnouncementMailContent(adminAnnouncement.Subject, adminAnnouncement.Body, a.config.Server.Host, user.Name())

		err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, user.Email, subject, body)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		notification := models.Notification{UserID: user.ID.String(), Msg: fmt.Sprintf("Announcement: %s", adminAnnouncement.Body)}
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

// SendEmailHandler creates a new administrator email and sends it to a specific user as an email and notification
// Example endpoint: Creates a new administrator email and sends it to a specific user as an email and notification
// @Summary Creates a new administrator email and sends it to a specific user as an email and notification
// @Description Creates a new administrator email and sends it to a specific user as an email and notification
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param email body EmailUser true "email to be sent"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /announcement [post]
func (a *App) SendEmailHandler(req *http.Request) (interface{}, Response) {
	var emailUser EmailUser
	err := json.NewDecoder(req.Body).Decode(&emailUser)

	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read email data"))
	}

	err = validator.Validate(emailUser)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid email data"))
	}

	user, err := a.db.GetUserByEmail(emailUser.Email)
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("user is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to get user"))
	}

	subject, body := internal.AdminMailContent(fmt.Sprintf("Hey! 📢 %s", emailUser.Subject), emailUser.Body, a.config.Server.Host, user.Name())

	err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, user.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	notification := models.Notification{UserID: user.ID.String(), Msg: fmt.Sprintf("Email: %s", emailUser.Body)}
	err = a.db.CreateNotification(&notification)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "new email is sent successfully",
	}, Created()
}

// UpdateNextLaunchHandler updates next launch flag
// Example endpoint: Updates next launch flag
// @Summary Updates next launch flag
// @Description Updates next launch flag
// @Tags Admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param nextlaunch body UpdateNextLaunchInput true "Next launch value to be set"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /nextlaunch [put]
func (a *App) UpdateNextLaunchHandler(req *http.Request) (interface{}, Response) {
	var input UpdateNextLaunchInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read NextLaunch update data"))
	}

	err = a.db.UpdateNextLaunch(input.Launched)
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
