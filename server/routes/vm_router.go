// Package routes for API endpoints
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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
	}

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
	}

	var input models.DeployVMInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "failed to read vm data")
	}

	id := len(r.vmRequestResponse) + 1
	err = r.redis.PushVMRequest(streams.VMDeployRequest{ID: id, User: user, Input: input})
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// wait for request
	key := fmt.Sprintf("%s %d", input.Name, id)
	r.mutex.Lock()
	r.vmRequestResponse[key] = streams.ErrResponse{}
	resCode := r.vmRequestResponse[key].Code
	r.mutex.Unlock()

	for resCode == nil {
		r.mutex.Lock()
		resCode = r.vmRequestResponse[key].Code
		r.mutex.Unlock()
	}

	res := r.vmRequestResponse[key]
	if res.Err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, *res.Code, res.Err.Error())
		return
	}

	writeMsgResponse(req, w, "Virtual machine is deployed successfully", "")
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

	err = r.cancelDeployment(vm.ContractID, vm.NetworkContractID)
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
		err = r.cancelDeployment(vm.ContractID, vm.NetworkContractID)
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

func (r *Router) deployVMRequest(ctx context.Context, user models.User, input models.DeployVMInput) (int, error) {
	err := validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusBadRequest, errors.New("invalid vm data")
	}

	// check quota of user
	quota, err := r.db.GetUserQuota(user.ID.String())
	if err == gorm.ErrRecordNotFound {
		return http.StatusNotFound, errors.New("user quota is not found")
	}
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	neededQuota, err := validateVMQuota(input, quota.Vms, quota.PublicIPs)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		return http.StatusBadRequest, errors.New("ssh key is required")
	}

	// unique names
	available, err := r.db.AvailableVMName(input.Name)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	if !available {
		return http.StatusBadRequest, errors.New("vm name is not available, please choose a different name")
	}

	vm, contractID, networkContractID, diskSize, err := r.deployVM(ctx, input, user.SSHKey)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	userVM := models.VM{
		UserID:            user.ID.String(),
		Name:              vm.Name,
		YggIP:             vm.YggIP,
		Resources:         input.Resources,
		Public:            input.Public,
		PublicIP:          vm.ComputedIP,
		SRU:               diskSize,
		CRU:               uint64(vm.CPU),
		MRU:               uint64(vm.Memory),
		ContractID:        contractID,
		NetworkContractID: networkContractID,
	}

	err = r.db.CreateVM(&userVM)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	publicIPsQuota := quota.PublicIPs
	if input.Public {
		publicIPsQuota -= publicQuota
	}
	// update quota of user
	err = r.db.UpdateUserQuota(user.ID.String(), quota.Vms-neededQuota, publicIPsQuota)
	if err == gorm.ErrRecordNotFound {
		return http.StatusNotFound, errors.New("User quota is not found")
	}
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	middlewares.Deployments.WithLabelValues(user.ID.String(), input.Resources, "vm").Inc()
	return 0, nil
}
