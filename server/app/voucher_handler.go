// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Length                 int `json:"length" binding:"required" validate:"min=3,max=20"`
	VMs                    int `json:"vms" binding:"required"`
	PublicIPs              int `json:"public_ips" binding:"required"`
	VoucherDurationInMonth int `json:"voucher_duration_in_month" binding:"required"`
}

// UpdateVoucherInput struct for data needed when user update voucher
type UpdateVoucherInput struct {
	Approved bool `json:"approved" binding:"required"`
}

// GenerateVoucherHandler generates a voucher by admin
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

	if input.VoucherDurationInMonth > a.config.VouchersMaxDuration {
		return nil, BadRequest(fmt.Errorf("invalid voucher duration, max duration is %d", a.config.VouchersMaxDuration))
	}

	v := models.Voucher{
		Voucher:                voucher,
		VMs:                    input.VMs,
		PublicIPs:              input.PublicIPs,
		Approved:               true,
		VoucherDurationInMonth: input.VoucherDurationInMonth,
	}

	err = a.db.CreateVoucher(&v)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	_, err = a.db.UpdateVoucher(v.ID, true)
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
		subject, body = internal.ApprovedVoucherMailContent(updatedVoucher.Voucher, user.Name, a.config.Server.Host)
	} else {
		subject, body = internal.RejectedVoucherMailContent(user.Name, a.config.Server.Host)
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

		subject, body := internal.ApprovedVoucherMailContent(v.Voucher, user.Name, a.config.Server.Host)
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
