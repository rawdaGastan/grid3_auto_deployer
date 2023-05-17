// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/codescalers/cloud4students/deployer"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

// DeployVMHandler creates vm for user and deploy it
func (r *Router) DeployVMHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusNotFound, "user is not found")
		return
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	var input models.DeployVMInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "failed to read vm data")
		return
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "invalid vm data")
		return
	}

	// check quota of user
	quota, err := r.db.GetUserQuota(user.ID.String())
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "user quota is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	_, err = deployer.ValidateVMQuota(input, quota.Vms, quota.PublicIPs)
	if err != nil {
		writeErrResponse(req, w, http.StatusBadRequest, err.Error())
		return
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		writeErrResponse(req, w, http.StatusBadRequest, "ssh key is required")
		return
	}

	// unique names
	available, err := r.db.AvailableVMName(input.Name)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if !available {
		writeErrResponse(req, w, http.StatusBadRequest, "vm name is not available, please choose a different name")
		return
	}

	err = r.deployer.Redis.PushVMRequest(streams.VMDeployRequest{User: user, Input: input})
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Virtual machine request is being deployed, you'll receive a confirmation notification soon", "")
}

// GetVMHandler returns vm by its id
func (r *Router) GetVMHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read vm id")
		return
	}

	vm, err := r.db.GetVMByID(id)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "Virtual machine not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if vm.UserID != userID {
		writeErrResponse(req, w, http.StatusNotFound, "Virtual machine not found")
		return
	}

	writeMsgResponse(req, w, "Virtual machine found", vm)
}

// ListVMsHandler returns all vms of user
func (r *Router) ListVMsHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	vms, err := r.db.GetAllVms(userID)
	if err == gorm.ErrRecordNotFound || len(vms) == 0 {
		writeMsgResponse(req, w, "Virtual machines not found", vms)
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Virtual machines found", vms)
}

// DeleteVM deletes vm by its id
func (r *Router) DeleteVM(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read vm id")
		return
	}

	vm, err := r.db.GetVMByID(id)
	if err == gorm.ErrRecordNotFound || vm.UserID != userID {
		writeErrResponse(req, w, http.StatusNotFound, "VM not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.deployer.CancelDeployment(vm.ContractID, vm.NetworkContractID)
	if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.db.DeleteVMByID(id)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	middlewares.Deletions.WithLabelValues(userID, "vms").Inc()
	writeMsgResponse(req, w, "Virtual machine is deleted successfully", "")
}

// DeleteAllVMs deletes all vms of user
func (r *Router) DeleteAllVMs(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	vms, err := r.db.GetAllVms(userID)
	if err == gorm.ErrRecordNotFound || len(vms) == 0 {
		writeMsgResponse(req, w, "Virtual machines not found", nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	for _, vm := range vms {
		err = r.deployer.CancelDeployment(vm.ContractID, vm.NetworkContractID)
		if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}
	}

	err = r.db.DeleteAllVms(userID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// metrics
	for _, vm := range vms {
		middlewares.Deletions.WithLabelValues(vm.UserID, "vms").Inc()
	}

	writeMsgResponse(req, w, "All Virtual machines are deleted successfully", "")
}
