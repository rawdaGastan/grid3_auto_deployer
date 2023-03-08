// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	_, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	var input GenerateVoucherInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		r.WriteErrResponse(w, err)
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
		r.WriteErrResponse(w, err)
		return
	}

	r.WriteMsgResponse(w, "voucher is created successfully", map[string]string{"voucher": voucher})
}
