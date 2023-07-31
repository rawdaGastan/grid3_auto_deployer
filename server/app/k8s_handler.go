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

// K8sDeployHandler deploy k8s handler
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

	var k8sDeployInput models.K8sDeployInput
	err = json.NewDecoder(req.Body).Decode(&k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read k8s data"))
	}

	err = validator.Validate(k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid kubernetes data"))
	}

	// balance verification
	balance, err := a.db.GetBalanceByUserID(user.ID.String())
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		return nil, NotFound(errors.New("user package is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = deployer.ValidateK8sQuota(k8sDeployInput, balance)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New(err.Error()))
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		return nil, BadRequest(errors.New("ssh key is required"))
	}

	// unique names
	available, err := a.db.AvailableK8sName(k8sDeployInput.MasterName)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !available {
		return nil, BadRequest(errors.New("kubernetes master name is not available, please choose a different name"))
	}

	err = a.deployer.Redis.PushK8sRequest(streams.K8sDeployRequest{User: user, Input: k8sDeployInput, AdminSSHKey: a.config.AdminSSHKey, ExpirationToleranceInDays: a.config.ExpirationToleranceInDays})
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

	return ResponseMsg{
		Message: "Kubernetes cluster is found",
		Data:    cluster,
	}, Ok()
}

// K8sGetAllHandler gets all clusters for a user
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
