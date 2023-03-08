// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/models"
)

// GenerateVoucherInput struct for data needed when user creates account
type GenerateVoucherInput struct {
	Length int `json:"length" binding:"required"`
	VMs    int `json:"vms" binding:"required"`
	K8s    int `json:"k8s" binding:"required"`
}

// GenerateVoucherHandler generates a voucher by admin
func (r *Router) GenerateVoucherHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("UserID").(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if !user.Admin {
		writeErrResponse(w, fmt.Errorf("user '%s' doesn't have an admin access", user.Name))
		return
	}

	var input GenerateVoucherInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	voucher := internal.GenerateRandomVoucher(input.Length)

	v := models.Voucher{
		Voucher: voucher,
		K8s:     input.K8s,
		VMs:     input.VMs,
	}

	err = r.db.CreateVoucher(v)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	writeMsgResponse(w, "voucher is created successfully", map[string]string{"voucher": voucher})
}
