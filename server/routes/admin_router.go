// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UpdateMaintenanceInput struct for data needed when user update maintenance
type UpdateMaintenanceInput struct {
	ON bool `json:"on" binding:"required"`
}

// GetAllUsersHandler returns all users
func (r *Router) GetAllUsersHandler(w http.ResponseWriter, req *http.Request) {
	users, err := r.db.ListAllUsers()
	if err == gorm.ErrRecordNotFound || len(users) == 0 {
		writeMsgResponse(req, w, "Users are not found", users)
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Users are found", users)
}

// GetBalance return account balance information
func (r *Router) GetBalance(w http.ResponseWriter, req *http.Request) {
	balance, err := r.deployer.GetBalance()
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Balance is found", balance)
}

// UpdateMaintenanceHandler updates maintenance flag
func (r *Router) UpdateMaintenanceHandler(w http.ResponseWriter, req *http.Request) {
	var input UpdateMaintenanceInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read maintenance update data")
		return
	}

	err = r.db.UpdateMaintenance(input.ON)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "maintenance is not found")
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Maintenance is updated successfully", "")
}

// GetMaintenanceHandler updates maintenance flag
func (r *Router) GetMaintenanceHandler(w http.ResponseWriter, req *http.Request) {
	maintenance, err := r.db.GetMaintenance()
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "maintenance is not found")
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, fmt.Sprintf("Maintenance is set with %v", maintenance.Active), maintenance)
}
