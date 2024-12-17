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

// K8sDeployInput deploy k8s cluster input
type K8sDeployInput struct {
	MasterName      string        `json:"master_name" validate:"min=3,max=20"`
	MasterResources string        `json:"resources"`
	MasterPublic    bool          `json:"public"`
	MasterRegion    string        `json:"region"`
	Workers         []WorkerInput `json:"workers"`
}

// WorkerInput deploy k8s worker input
type WorkerInput struct {
	Name      string `json:"name" validate:"min=3,max=20"`
	Resources string `json:"resources"`
}

// K8sDeployHandler deploy k8s handler
// Example endpoint: Deploy kubernetes
// @Summary Deploy kubernetes
// @Description Deploy kubernetes
// @Tags Kubernetes
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param kubernetes body K8sDeployInput true "Kubernetes deployment input"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /k8s [post]
func (a *App) K8sDeployHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	var input K8sDeployInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read k8s data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid kubernetes data"))
	}

	cru, mru, sru, _, err := deployer.CalcNodeResources(input.MasterResources, input.MasterPublic)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	master := models.Master{
		CRU:       cru,
		MRU:       mru,
		SRU:       sru,
		Public:    input.MasterPublic,
		Name:      input.MasterName,
		Resources: input.MasterResources,
		Region:    input.MasterRegion,
	}

	workers := []models.Worker{}
	for _, worker := range input.Workers {
		cru, mru, sru, _, err := deployer.CalcNodeResources(worker.Resources, false)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		workerModel := models.Worker{
			Name:      worker.Name,
			CRU:       cru,
			MRU:       mru,
			SRU:       sru,
			Public:    input.MasterPublic,
			Resources: worker.Resources,
			Region:    input.MasterRegion,
		}
		workers = append(workers, workerModel)
	}

	k8sCluster := models.K8sCluster{
		UserID:  userID,
		Master:  master,
		Workers: workers,
	}

	// check if user can deploy? cards verification or voucher balance exists
	k8sPrice, err := a.deployer.CanDeployK8s(user.ID.String(), k8sCluster)
	if errors.Is(err, deployer.ErrCannotDeploy) {
		return nil, BadRequest(err)
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	k8sCluster.PricePerMonth = k8sPrice

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		return nil, BadRequest(errors.New("ssh key is required"))
	}

	// unique names
	available, err := a.db.AvailableK8sName(input.MasterName)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !available {
		return nil, BadRequest(errors.New("kubernetes master name is not available, please choose a different name"))
	}

	k8sCluster.State = models.StateInProgress
	err = a.db.CreateK8s(&k8sCluster)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.deployer.Redis.PushK8sRequest(streams.K8sDeployRequest{User: user, Cluster: k8sCluster, AdminSSHKey: a.config.AdminSSHKey})
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Kubernetes cluster request is being deployed, you'll receive a confirmation notification soon",
		Data:    nil,
	}, Created()
}

// ValidateK8sNameHandler validates a cluster name
// Example endpoint: Validate kubernetes name
// @Summary Validate kubernetes name
// @Description Validate kubernetes name
// @Tags Kubernetes
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param name path string true "Kubernetes name"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /k8s/validate/{name} [get]
func (a *App) ValidateK8sNameHandler(req *http.Request) (interface{}, Response) {
	name := mux.Vars(req)["name"]

	err := validator.Validate(name)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid k8s data"))
	}

	// unique names
	available, err := a.db.AvailableK8sName(name)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !available {
		return nil, BadRequest(errors.New("kubernetes cluster name is not available, please choose a different name"))
	}

	return ResponseMsg{
		Message: "kubernetes cluster name is available",
		Data:    nil,
	}, Ok()
}

// K8sGetHandler gets a cluster for a user
// Example endpoint: Get kubernetes deployment using ID
// @Summary Get kubernetes deployment using ID
// @Description Get kubernetes deployment using ID
// @Tags Kubernetes
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Kubernetes cluster ID"
// @Success 200 {object} models.K8sCluster
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /k8s/{id} [get]
func (a *App) K8sGetHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read cluster id"))
	}

	cluster, err := a.db.GetK8s(id)
	if err == gorm.ErrRecordNotFound || cluster.UserID != userID {
		return nil, NotFound(errors.New("kubernetes cluster is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if userID != cluster.UserID {
		return nil, NotFound(errors.New("cluster is not found"))
	}

	return ResponseMsg{
		Message: "Kubernetes cluster is found",
		Data:    cluster,
	}, Ok()
}

// K8sGetAllHandler gets all clusters for a user
// Example endpoint: Get user's kubernetes deployments
// @Summary Get user's kubernetes deployments
// @Description Get user's kubernetes deployments
// @Tags Kubernetes
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.K8sCluster
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /k8s [get]
func (a *App) K8sGetAllHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	clusters, err := a.db.GetAllK8s(userID)
	if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
		return ResponseMsg{
			Message: "Kubernetes clusters are not found",
			Data:    clusters,
		}, Ok()
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Kubernetes clusters are found",
		Data:    clusters,
	}, Ok()
}

// K8sDeleteHandler deletes a cluster for a user
// Example endpoint: Delete kubernetes deployment using ID
// @Summary Delete kubernetes deployment using ID
// @Description Delete kubernetes deployment using ID
// @Tags Kubernetes
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Kubernetes cluster ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /k8s/{id} [delete]
func (a *App) K8sDeleteHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		return nil, BadRequest(errors.New("failed to read cluster id"))
	}

	cluster, err := a.db.GetK8s(id)
	if err == gorm.ErrRecordNotFound || cluster.UserID != userID {
		return nil, NotFound(errors.New("kubernetes cluster is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if userID != cluster.UserID {
		return nil, NotFound(errors.New("cluster is not found"))
	}

	err = a.deployer.CancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract), "k8s", cluster.Master.Name)
	if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.db.DeleteK8s(id)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// metrics
	middlewares.Deletions.WithLabelValues(userID, "k8s").Inc()

	return ResponseMsg{
		Message: "kubernetes cluster is deleted successfully",
		Data:    nil,
	}, Ok()
}

// K8sDeleteAllHandler deletes all clusters for a user
// Example endpoint: Delete all user's kubernetes deployments
// @Summary Delete all user's kubernetes deployments
// @Description Delete all user's kubernetes deployments
// @Tags Kubernetes
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /k8s [delete]
func (a *App) K8sDeleteAllHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	clusters, err := a.db.GetAllK8s(userID)
	if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
		return ResponseMsg{
			Message: "Kubernetes clusters are not found",
			Data:    nil,
		}, Ok()
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, cluster := range clusters {
		err = a.deployer.CancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract), "k8s", cluster.Master.Name)
		if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	err = a.db.DeleteAllK8s(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, c := range clusters {
		middlewares.Deletions.WithLabelValues(c.UserID, "k8s").Inc()
	}

	return ResponseMsg{
		Message: "All kubernetes clusters are deleted successfully",
		Data:    nil,
	}, Ok()
}
