// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

// GenerateVoucherInput struct for data needed when user creates account
type GenerateVoucherInput struct {
	Length int `json:"length" binding:"required" validate:"min=3,max=20"`
	VMs    int `json:"vms" binding:"required"`
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
		writeErrResponse(w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}
	*/

	var input GenerateVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusBadRequest, "Failed to read voucher data")
		return
	}

	voucher := internal.GenerateRandomVoucher(input.Length)
	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusBadRequest, "Invalid voucher data")
		return
	}

	v := models.Voucher{
		Voucher: voucher,
		VMs:     input.VMs,
	}

	err = r.db.CreateVoucher(&v)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	_, err = r.db.ApproveVoucher(v.ID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Voucher is generated successfully", map[string]string{"voucher": voucher})
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
		writeErrResponse(w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}
	*/

	vouchers, err := r.db.ListAllVouchers()
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "List of all vouchers", vouchers)
}

// ApproveVoucherHandler approves a voucher by admin
func (r *Router) ApproveVoucherHandler(w http.ResponseWriter, req *http.Request) {
	/*userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}


	if !user.Admin {
		writeErrResponse(w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}
	*/

	// get voucher id from url
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, "Failed to read vm id")
		return
	}

	voucher, err := r.db.GetVoucherByID(id)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if voucher.Approved {
		writeErrResponse(w, http.StatusBadRequest, "Voucher is already approved")
		return
	}

	approvedVoucher, err := r.db.ApproveVoucher(id)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	user, err := r.db.GetUserByID(approvedVoucher.UserID)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	subject, body := internal.ApprovedVoucherMailContent(approvedVoucher.Voucher, user.Name)
	err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.SendGridKey, user.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	writeMsgResponse(w, "Confirmation mail has been sent to the user", "")
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
		writeErrResponse(w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}
	*/

	vouchers, err := r.db.ApproveAllVouchers()
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	for _, v := range vouchers {
		user, err := r.db.GetUserByID(v.UserID)
		if err != nil {
			writeErrResponse(w, http.StatusNotFound, fmt.Sprintf("User '%s' is not found", v.UserID))
			return
		}

		subject, body := internal.ApprovedVoucherMailContent(v.Voucher, user.Name)
		err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.SendGridKey, user.Email, subject, body)
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}
	}

	writeMsgResponse(w, "All vouchers are approved and confirmation mails has been sent to the user", "")
}
