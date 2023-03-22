// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DeployVMInput struct takes input of vm from user
type DeployVMInput struct {
	Name      string `json:"name" binding:"required"`
	Resources string `json:"resources" binding:"required"`
}

// DeployVMHandler creates vm for user and deploy it
func (r *Router) DeployVMHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		writeNotFoundResponse(w, "User not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	var input DeployVMInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, "Faile to read vm data")
		return
	}

	// check quota of user
	quota, err := r.db.GetUserQuota(userID)
	if err == gorm.ErrRecordNotFound {
		writeNotFoundResponse(w, "User quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	neededQuota, err := validateVMQuota(input.Resources, quota.Vms)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		writeErrResponse(w, "ssh key is required")
		return
	}

	vm, contractID, networkContractID, diskSize, err := r.deployVM(input.Name, input.Resources, user.SSHKey)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	userVM := models.VM{
		UserID:            userID,
		Name:              vm.Name,
		IP:                vm.YggIP,
		Resources:         input.Resources,
		SRU:               diskSize,
		CRU:               uint64(vm.CPU),
		MRU:               uint64(vm.Memory),
		ContractID:        contractID,
		NetworkContractID: networkContractID,
	}

	err = r.db.CreateVM(&userVM)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	// update quota of user
	err = r.db.UpdateUserQuota(userID, quota.Vms-neededQuota, quota.K8s)
	if err == gorm.ErrRecordNotFound {
		writeNotFoundResponse(w, "User quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Virtual machine is deployed successfully", map[string]int{"ID": userVM.ID})
}

// GetVMHandler returns vm by its id
func (r *Router) GetVMHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, "Failed to parse vm id")
		return
	}

	vm, err := r.db.GetVMByID(id)
	if err == gorm.ErrRecordNotFound {
		writeNotFoundResponse(w, "Virtual machine not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	if vm.UserID != userID {
		writeNotFoundResponse(w, "Virtual machine not found")
		return
	}

	writeMsgResponse(w, "Virtual machine found", vm)
}

// ListVMsHandler returns all vms of user
func (r *Router) ListVMsHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	vms, err := r.db.GetAllVms(userID)
	if err == gorm.ErrRecordNotFound || len(vms) == 0 {
		writeMsgResponse(w, "Virtual machines not found", nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Virtual machines found", vms)
}

// DeleteVM deletes vm by its id
func (r *Router) DeleteVM(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, "Failed to parse vm id")
		return
	}

	vm, err := r.db.GetVMByID(id)
	if err == gorm.ErrRecordNotFound {
		writeNotFoundResponse(w, "VM not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	if vm.UserID != userID {
		writeNotFoundResponse(w, "Virtual machine not found")
		return
	}

	err = r.cancelDeployment(vm.ContractID, vm.NetworkContractID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	err = r.db.DeleteVMByID(id)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Virtual machine is deleted successfully", "")
}

// DeleteAllVMs deletes all vms of user
func (r *Router) DeleteAllVMs(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	vms, err := r.db.GetAllVms(userID)
	if err == gorm.ErrRecordNotFound || len(vms) == 0 {
		writeMsgResponse(w, "Virtual machines not found", nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	for _, vm := range vms {
		err = r.cancelDeployment(vm.ContractID, vm.NetworkContractID)
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(w, internalServerErrorMsg)
			return
		}
	}

	err = r.db.DeleteAllVms(userID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "All Virtual machines are deleted successfully", "")
}
