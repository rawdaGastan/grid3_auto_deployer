// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rawdaGastan/cloud4students/models"
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
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}

	var input DeployVMInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	// check quota of user
	quota, err := r.db.GetUserQuota(userID)
	if err != nil {
		writeErrResponse(w, err.Error())
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
		writeErrResponse(w, err.Error())
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

	err = r.db.CreateVM(userVM)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	// update quota of user
	err = r.db.UpdateUserQuota(userID, quota.Vms-neededQuota, quota.K8s)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	writeMsgResponse(w, "Virtual machine is deployed successfully", map[string]int{"ID": userVM.ID})
}

// GetVMHandler returns vm by its id
func (r *Router) GetVMHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	vm, err := r.db.GetVMByID(id)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}

	if vm.UserID != userID {
		writeNotFoundResponse(w, fmt.Sprintf("Virtual machine with ID %v is not found", id))
		return
	}

	writeMsgResponse(w, "Virtual machine is found", vm)
}

// ListVMsHandler returns all vms of user
func (r *Router) ListVMsHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	vms, err := r.db.GetAllVms(userID)
	if err != nil {
		writeMsgResponse(w, "Virtual machines are not found", vms)
		return
	}

	if len(vms) > 0 {
		writeMsgResponse(w, "Virtual machines are not found", vms)
		return
	}

	writeMsgResponse(w, "Virtual machines are found", vms)
}

// DeleteVM deletes vm by its id
func (r *Router) DeleteVM(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	vm, err := r.db.GetVMByID(id)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}

	if vm.UserID != userID {
		writeNotFoundResponse(w, "Virtual machine is not found")
		return
	}

	err = r.cancelDeployment(vm.ContractID, vm.NetworkContractID)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	err = r.db.DeleteVMByID(id)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	writeMsgResponse(w, "Virtual machine is deleted successfully", "")
}

// DeleteAllVMs deletes all vms of user
func (r *Router) DeleteAllVMs(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	vms, err := r.db.GetAllVms(userID)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}

	for _, vm := range vms {
		err = r.cancelDeployment(vm.ContractID, vm.NetworkContractID)
		if err != nil {
			writeErrResponse(w, err.Error())
			return
		}
	}

	err = r.db.DeleteAllVms(userID)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	writeMsgResponse(w, "All Virtual machines are deleted successfully", "")
}
