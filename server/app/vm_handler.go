// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
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
func (a *App) DeployVMHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	var input models.DeployVMInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read vm data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid vm data"))
	}

	// check quota of user
	quota, err := a.db.GetUserQuota(user.ID.String())
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user quota is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	_, err = deployer.ValidateVMQuota(input, quota.Vms, quota.PublicIPs)
	if err != nil {
		return nil, BadRequest(errors.New(err.Error()))
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		return nil, BadRequest(errors.New("ssh key is required"))
	}

	// unique names
	available, err := a.db.AvailableVMName(input.Name)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !available {
		return nil, BadRequest(errors.New("virtual machine name is not available, please choose a different name"))
	}

	err = a.deployer.Redis.PushVMRequest(streams.VMDeployRequest{User: user, Input: input, AdminSSHKey: a.config.AdminSSHKey})
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Virtual machine request is being deployed, you'll receive a confirmation notification soon",
		Data:    nil,
	}, Created()
}

// ValidateVMNameHandler validates a vm name
func (a *App) ValidateVMNameHandler(req *http.Request) (interface{}, Response) {
	name := mux.Vars(req)["name"]

	err := validator.Validate(name)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid vm data"))
	}

	// unique names
	available, err := a.db.AvailableVMName(name)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !available {
		return nil, BadRequest(errors.New("virtual machine name is not available, please choose a different name"))
	}

	return ResponseMsg{
		Message: "Virtual machine name is available",
		Data:    nil,
	}, Ok()
}

// GetVMHandler returns vm by its id
func (a *App) GetVMHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		return nil, BadRequest(errors.New("failed to read vm id"))
	}

	vm, err := a.db.GetVMByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("virtual machine not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if vm.UserID != userID {
		return nil, NotFound(errors.New("virtual machine not found"))
	}

	return ResponseMsg{
		Message: "Virtual machine is found",
		Data:    vm,
	}, Ok()
}

// ListVMsHandler returns all vms of user
func (a *App) ListVMsHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	vms, err := a.db.GetAllVms(userID)
	if err == gorm.ErrRecordNotFound || len(vms) == 0 {
		return ResponseMsg{
			Message: "no virtual machines found",
			Data:    vms,
		}, Ok()
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Virtual machines are found",
		Data:    vms,
	}, Ok()
}

// DeleteVMHandler deletes vm by its id
func (a *App) DeleteVMHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read vm id"))
	}

	vm, err := a.db.GetVMByID(id)
	if err == gorm.ErrRecordNotFound || vm.UserID != userID {
		return nil, NotFound(errors.New("virtual machine is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.deployer.CancelDeployment(vm.ContractID, vm.NetworkContractID, "vm", vm.Name)
	if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.db.DeleteVMByID(id)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	middlewares.Deletions.WithLabelValues(userID, "vms").Inc()
	return ResponseMsg{
		Message: "Virtual machine is deleted successfully",
		Data:    nil,
	}, Ok()
}

// DeleteAllVMsHandler deletes all vms of user
func (a *App) DeleteAllVMsHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	vms, err := a.db.GetAllVms(userID)
	if err == gorm.ErrRecordNotFound || len(vms) == 0 {
		return ResponseMsg{
			Message: "Virtual machines are not found",
			Data:    nil,
		}, Ok()
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, vm := range vms {
		err = a.deployer.CancelDeployment(vm.ContractID, vm.NetworkContractID, "vm", vm.Name)
		if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	err = a.db.DeleteAllVms(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// metrics
	for _, vm := range vms {
		middlewares.Deletions.WithLabelValues(vm.UserID, "vms").Inc()
	}

	middlewares.Deletions.WithLabelValues(userID, "vms").Inc()
	return ResponseMsg{
		Message: "All virtual machines are deleted successfully",
		Data:    nil,
	}, Ok()
}
