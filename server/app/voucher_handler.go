// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

// GenerateVoucherInput struct for data needed when user generate vouchers
type GenerateVoucherInput struct {
	Length  int    `json:"length" binding:"required" validate:"min=3,max=20"`
	Balance uint64 `json:"balance" binding:"required"`
}

// UpdateVoucherInput struct for data needed when user update voucher
type UpdateVoucherInput struct {
	Approved bool `json:"approved" binding:"required"`
}

// GenerateVoucherHandler generates a voucher by admin
// Example endpoint: Generates a new voucher
// @Summary Generates a new voucher
// @Description Generates a new voucher
// @Tags Voucher (only admins)
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param voucher body GenerateVoucherInput true "Voucher details"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /voucher [post]
func (a *App) GenerateVoucherHandler(req *http.Request) (interface{}, Response) {
	var input GenerateVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read voucher data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid voucher data"))
	}
	voucher := internal.GenerateRandomVoucher(input.Length)

	v := models.Voucher{
		Voucher:  voucher,
		Balance:  input.Balance,
		Approved: true,
	}

	err = a.db.CreateVoucher(&v)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Voucher is generated successfully",
		Data:    map[string]string{"voucher": voucher},
	}, Created()
}

// ListVouchersHandler lists all vouchers by admin
// Example endpoint: Lists users' vouchers
// @Summary Lists users' vouchers
// @Description Lists users' vouchers
// @Tags Voucher (only admins)
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.Voucher
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /voucher [get]
func (a *App) ListVouchersHandler(req *http.Request) (interface{}, Response) {
	vouchers, err := a.db.ListAllVouchers()
	if err == gorm.ErrRecordNotFound || len(vouchers) == 0 {
		return ResponseMsg{
			Message: "Vouchers are not found",
			Data:    vouchers,
		}, Ok()
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "List of all vouchers",
		Data:    vouchers,
	}, Ok()
}

// UpdateVoucherHandler approves/rejects a voucher by admin
// Example endpoint: Update (approve-reject) a voucher
// @Summary Update (approve-reject) a voucher
// @Description Update (approve-reject) a voucher
// @Tags Voucher (only admins)
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Voucher ID"
// @Param state body UpdateVoucherInput true "Voucher approval state"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /voucher/{id} [put]
func (a *App) UpdateVoucherHandler(req *http.Request) (interface{}, Response) {
	var input UpdateVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read voucher update data"))
	}

	// get voucher id from url
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		return nil, BadRequest(errors.New("failed to read voucher id"))
	}

	voucher, err := a.db.GetVoucherByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("voucher is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if voucher.Approved && input.Approved {
		return nil, BadRequest(errors.New("voucher is already approved"))
	}

	if voucher.Rejected && !input.Approved {
		return nil, BadRequest(errors.New("voucher is already rejected"))
	}

	updatedVoucher, err := a.db.UpdateVoucher(id, input.Approved)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	user, err := a.db.GetUserByID(updatedVoucher.UserID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	var subject, body string
	if input.Approved {
		subject, body = internal.ApprovedVoucherMailContent(updatedVoucher.Voucher, user.Name(), a.config.Server.Host)
	} else {
		subject, body = internal.RejectedVoucherMailContent(user.Name(), a.config.Server.Host)
	}

	err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, user.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Update mail has been sent to the user",
		Data:    nil,
	}, Ok()
}

// ApproveAllVouchersHandler approves all vouchers by admin
// Example endpoint: Approve all vouchers
// @Summary Approve all vouchers
// @Description Approve all vouchers
// @Tags Voucher (only admins)
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /voucher [put]
func (a *App) ApproveAllVouchersHandler(req *http.Request) (interface{}, Response) {
	vouchers, err := a.db.ListAllVouchers()
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, v := range vouchers {
		if v.Approved || v.Rejected {
			continue
		}

		_, err := a.db.UpdateVoucher(v.ID, true)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		user, err := a.db.GetUserByID(v.UserID)
		if err == gorm.ErrRecordNotFound {
			continue
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		subject, body := internal.ApprovedVoucherMailContent(v.Voucher, user.Name(), a.config.Server.Host)
		err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, user.Email, subject, body)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	return ResponseMsg{
		Message: "All vouchers are approved and confirmation mails has been sent to the users",
		Data:    nil,
	}, Ok()
}

// ResetUsersVoucherBalanceHandler resets all users voucher balance
// Example endpoint: Resets all users voucher balance
// @Summary Resets all users voucher balance
// @Description Resets all users voucher balance
// @Tags Voucher (only admins)
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} float64
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /voucher/reset [put]
func (a *App) ResetUsersVoucherBalanceHandler(req *http.Request) (interface{}, Response) {
	users, err := a.db.ListAllUsers()
	if err == gorm.ErrRecordNotFound || len(users) == 0 {
		return ResponseMsg{
			Message: "Users are not found",
		}, Ok()
	}

	for _, user := range users {
		user.VoucherBalance = 0
		err = a.db.UpdateUserByID(user)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	return ResponseMsg{
		Message: "Voucher balance is reset successfully",
	}, Ok()
}
