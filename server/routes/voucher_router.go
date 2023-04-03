// Package routes for API endpoints
package routes

import (
	"encoding/json"
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
	Length    int `json:"length" binding:"required" validate:"min=3,max=20"`
	VMs       int `json:"vms" binding:"required"`
	PublicIPs int `json:"public_ips" binding:"required"`
}

// UpdateVoucherInput struct for data needed when user update voucher
type UpdateVoucherInput struct {
	Approved bool `json:"approved" binding:"required"`
}

// GenerateVoucherHandler generates a voucher by admin
func (r *Router) GenerateVoucherHandler(w http.ResponseWriter, req *http.Request) {
	/*userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if !user.Admin {
		writeErrResponse(req, w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}
	*/

	var input GenerateVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read voucher data")
		return
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Invalid voucher data")
		return
	}
	voucher := internal.GenerateRandomVoucher(input.Length)

	v := models.Voucher{
		Voucher:   voucher,
		VMs:       input.VMs,
		PublicIPs: input.PublicIPs,
	}

	err = r.db.CreateVoucher(&v)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	_, err = r.db.UpdateVoucher(v.ID, true)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Voucher is generated successfully", map[string]string{"voucher": voucher})
}

// ListVouchersHandler lists all vouchers by admin
func (r *Router) ListVouchersHandler(w http.ResponseWriter, req *http.Request) {
	/*userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if !user.Admin {
		writeErrResponse(req, w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}
	*/

	vouchers, err := r.db.ListAllVouchers()
	if err == gorm.ErrRecordNotFound || len(vouchers) == 0 {
		writeMsgResponse(req, w, "Vouchers are not found", vouchers)
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "List of all vouchers", vouchers)
}

// UpdateVoucherHandler approves/rejects a voucher by admin
func (r *Router) UpdateVoucherHandler(w http.ResponseWriter, req *http.Request) {
	/*userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}


	if !user.Admin {
		writeErrResponse(req, w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}
	*/

	var input UpdateVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read voucher update data")
		return
	}

	// get voucher id from url
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read voucher id")
		return
	}

	voucher, err := r.db.GetVoucherByID(id)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "Voucher is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if voucher.Approved && input.Approved {
		writeErrResponse(req, w, http.StatusBadRequest, "Voucher is already approved")
		return
	}

	if !voucher.Approved && !input.Approved {
		writeErrResponse(req, w, http.StatusBadRequest, "Voucher is already rejected")
		return
	}

	updatedVoucher, err := r.db.UpdateVoucher(id, input.Approved)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	user, err := r.db.GetUserByID(updatedVoucher.UserID)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	var subject, body string
	if input.Approved {
		subject, body = internal.ApprovedVoucherMailContent(updatedVoucher.Voucher, user.Name)
	} else {
		subject, body = internal.RejectedVoucherMailContent(user.Name)
	}

	err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.SendGridKey, user.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	writeMsgResponse(req, w, "Update mail has been sent to the user", "")
}

// ApproveAllVouchers approves all vouchers by admin
func (r *Router) ApproveAllVouchers(w http.ResponseWriter, req *http.Request) {
	/*userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if !user.Admin {
		writeErrResponse(req, w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}
	*/

	vouchers, err := r.db.ApproveAllVouchers()
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	for _, v := range vouchers {
		user, err := r.db.GetUserByID(v.UserID)
		if err == gorm.ErrRecordNotFound {
			continue
		}
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}

		subject, body := internal.ApprovedVoucherMailContent(v.Voucher, user.Name)
		err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.SendGridKey, user.Email, subject, body)
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}
	}

	writeMsgResponse(req, w, "All vouchers are approved and confirmation mails has been sent to the user", "")
}
