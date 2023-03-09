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
		writeNotFoundResponse(w, err)
		return
	}

	var input DeployVMInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// check quota of user
	quota, err := r.db.GetUserQuota(userID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	neededQuota, err := validateVMQuota(input.Resources, quota.Vms)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		writeErrResponse(w, fmt.Errorf("ssh key is required"))
		return
	}

	vm, contractID, networkContractID, diskSize, err := r.deployVM(input.Name, input.Resources, user.SSHKey)
	if err != nil {
		writeErrResponse(w, err)
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
		writeErrResponse(w, err)
		return
	}

	// update quota of user
	err = r.db.UpdateUserQuota(userID, quota.Vms-neededQuota, quota.K8s)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	writeMsgResponse(w, "VM is deployed successfully", map[string]int{"ID": userVM.ID})
}

// GetVMHandler returns vm by its id
func (r *Router) GetVMHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	vm, err := r.db.GetVMByID(id)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if vm.UserID != userID {
		writeNotFoundResponse(w, fmt.Errorf("vm not found"))
		return
	}

	writeMsgResponse(w, "VM is found", vm)
}

// ListVMsHandler returns all vms of user
func (r *Router) ListVMsHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	vms, err := r.db.GetAllVms(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	writeMsgResponse(w, "VMs are found", vms)
}

// DeleteVM deletes vm by its id
func (r *Router) DeleteVM(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	vm, err := r.db.GetVMByID(id)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if vm.UserID != userID {
		writeNotFoundResponse(w, fmt.Errorf("vm not found"))
		return
	}

	err = r.cancelDeployment(vm.ContractID, vm.NetworkContractID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = r.db.DeleteVMByID(id)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	writeMsgResponse(w, "VM is deleted successfully", "")
}

// DeleteAllVMs deletes all vms of user
func (r *Router) DeleteAllVMs(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	vms, err := r.db.GetAllVms(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	for _, vm := range vms {
		err = r.cancelDeployment(vm.ContractID, vm.NetworkContractID)
		if err != nil {
			writeErrResponse(w, err)
			return
		}
	}

	err = r.db.DeleteAllVms(userID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	writeMsgResponse(w, "All vms are deleted successfully", "")
}
