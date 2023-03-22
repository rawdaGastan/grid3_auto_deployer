// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/models"
)

// GenerateVoucherInput struct for data needed when user creates account
type GenerateVoucherInput struct {
	Length int `json:"length" binding:"required"`
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
		writeErrResponse(w, err.Error())
		return
	}

	voucher := internal.GenerateRandomVoucher(input.Length)

	v := models.Voucher{
		Voucher: voucher,
		VMs:     input.VMs,
	}

	err = r.db.CreateVoucher(&v)
	if err != nil {
		writeErrResponse(w, err.Error())
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
		writeErrResponse(w, err.Error())
		return
	}

	writeMsgResponse(w, "List of all vouchers", vouchers)
}
// ApproveVoucher approves a voucher by admin
func (r *Router) ApproveVoucher(w http.ResponseWriter, req *http.Request) {
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
	// get user id from url
	id := mux.Vars(req)["id"]
	voucher, err := r.db.ActivateVoucher(id)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}
	// TODO: send confirmation email to user via third party
	fmt.Printf("voucher: %v\n", voucher)
	writeMsgResponse(w, "Confirmation mail's sent to the user", "")
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

	vouchers, err := r.db.ListAllVouchers()
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	for _, v := range vouchers {
		voucher, err := r.db.ActivateVoucher(v.Voucher)
		if err != nil {
			writeErrResponse(w, err.Error())
			return
		}
		fmt.Printf("voucher: %v\n", voucher)
		// TODO: send confirmation email to user via third party
	}

	writeMsgResponse(w, "All vouchers are approved", vouchers)
}
