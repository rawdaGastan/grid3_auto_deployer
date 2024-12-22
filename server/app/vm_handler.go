// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/codescalers/cloud4students/deployer"
	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

// DeployVMInput struct takes input of vm from user
type DeployVMInput struct {
	Name      string `json:"name" binding:"required" validate:"min=3,max=20"`
	Resources string `json:"resources" binding:"required"`
	Public    bool   `json:"public"`
	Region    string `json:"region"`
}

// DeployVMHandler creates vm for user and deploy it
// Example endpoint: Deploy virtual machine
// @Summary Deploy virtual machine
// @Description Deploy virtual machine
// @Tags VM
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param vm body DeployVMInput true "virtual machine deployment input"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /vm [post]
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

	var input DeployVMInput
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

	cru, mru, sru, _, err := deployer.CalcNodeResources(input.Resources, input.Public)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	vm := models.VM{
		UserID:    userID,
		Name:      input.Name,
		Resources: input.Resources,
		Public:    input.Public,
		SRU:       sru,
		CRU:       cru,
		MRU:       mru * 1024,
		Region:    input.Region,
	}

	vmPrice, err := a.deployer.CanDeployVM(user.ID.String(), vm)
	if errors.Is(err, deployer.ErrCannotDeploy) {
		return nil, BadRequest(err)
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	vm.PricePerMonth = vmPrice

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

	vm.State = models.StateInProgress
	err = a.db.CreateVM(&vm)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.deployer.Redis.PushVMRequest(streams.VMDeployRequest{User: user, VM: vm, AdminSSHKey: a.config.AdminSSHKey})
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
// Example endpoint: Validate virtual machine name
// @Summary Validate virtual machine name
// @Description Validate virtual machine name
// @Tags VM
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param name path string true "Virtual machine name"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /vm/validate/{name} [get]
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
// Example endpoint: Get virtual machine deployment using ID
// @Summary Get virtual machine deployment using ID
// @Description Get virtual machine deployment using ID
// @Tags VM
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Virtual machine ID"
// @Success 200 {object} models.VM
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /vm/{id} [get]
func (a *App) GetVMHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		return nil, BadRequest(errors.New("failed to read vm id"))
	}

	vm, err := a.db.GetVMByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("virtual machine is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if vm.UserID != userID {
		return nil, NotFound(errors.New("virtual machine is not found"))
	}

	return ResponseMsg{
		Message: "Virtual machine is found",
		Data:    vm,
	}, Ok()
}

// ListVMsHandler returns all vms of user
// Example endpoint: Get user's virtual machine deployments
// @Summary Get user's virtual machine deployments
// @Description Get user's virtual machine deployments
// @Tags VM
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.VM
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /vm [get]
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
// Example endpoint: Delete virtual machine deployment using ID
// @Summary Delete virtual machine deployment using ID
// @Description Delete virtual machine deployment using ID
// @Tags VM
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Virtual machine ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /vm/{id} [delete]
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

	if vm.UserID != userID {
		return nil, NotFound(errors.New("virtual machine is not found"))
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
// Example endpoint: Delete all user's virtual machine deployments
// @Summary Delete all user's virtual machine deployments
// @Description Delete all user's virtual machine deployments
// @Tags VM
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /vm [delete]
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

// ListRegionsHandler returns all supported regions
// Example endpoint: List all supported regions
// @Summary List all supported regions
// @Description List all supported regions
// @Tags Region
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []string
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /region [get]
func (a *App) ListRegionsHandler(req *http.Request) (interface{}, Response) {
	stats, err := a.deployer.TFPluginClient.GridProxyClient.Stats(req.Context(), types.StatsFilter{})
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	graphql, err := internal.NewGraphQl(a.config.Account.Network)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	var countries []string
	for country := range stats.NodesDistribution {
		countries = append(countries, country)
	}

	regions, err := graphql.ListRegions(countries)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Regions are found",
		Data:    regions,
	}, Ok()
}
