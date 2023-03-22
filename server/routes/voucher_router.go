// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"net/http"

	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rs/zerolog/log"
)

// GenerateVoucherInput struct for data needed when user creates account
type GenerateVoucherInput struct {
	Length int `json:"length" binding:"required"`
	VMs    int `json:"vms" binding:"required"`
	K8s    int `json:"k8s" binding:"required"`
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
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	voucher := internal.GenerateRandomVoucher(input.Length)

	v := models.Voucher{
		Voucher: voucher,
		K8s:     input.K8s,
		VMs:     input.VMs,
	}

	err = r.db.CreateVoucher(&v)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Voucher is generated successfully", map[string]string{"voucher": voucher})
}
