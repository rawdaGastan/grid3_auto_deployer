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

// K8sDeployHandler deploy k8s handler
func (r *Router) K8sDeployHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	var k8sDeployInput models.K8sDeployInput
	err = json.NewDecoder(req.Body).Decode(&k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read k8s data")
		return
	}

	id := len(r.k8sRequestResponse) + 1
	err = r.redis.PushK8sRequest(streams.K8sDeployRequest{ID: id, User: user, Input: k8sDeployInput})
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// wait for request
	key := fmt.Sprintf("%s %d", k8sDeployInput.MasterName, id)
	r.mutex.Lock()
	r.k8sRequestResponse[key] = streams.ErrResponse{}
	resCode := r.k8sRequestResponse[key].Code
	r.mutex.Unlock()

	for resCode == nil {
		r.mutex.Lock()
		resCode = r.k8sRequestResponse[key].Code
		r.mutex.Unlock()
	}

	res := r.k8sRequestResponse[key]
	if res.Err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, *res.Code, res.Err.Error())
		return
	}

	// write response
	writeMsgResponse(req, w, "Kubernetes cluster is deployed successfully", "")
}

// K8sGetHandler gets a cluster for a user
func (r *Router) K8sGetHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read cluster id")
		return
	}

	cluster, err := r.db.GetK8s(id)
	if err == gorm.ErrRecordNotFound || cluster.UserID != userID {
		writeErrResponse(req, w, http.StatusNotFound, "Kubernetes cluster not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Kubernetes cluster found", cluster)
}

// K8sGetAllHandler gets all clusters for a user
func (r *Router) K8sGetAllHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	clusters, err := r.db.GetAllK8s(userID)
	if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
		writeMsgResponse(req, w, "Kubernetes clusters not found", clusters)
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Kubernetes clusters found", clusters)
}

// K8sDeleteHandler deletes a cluster for a user
func (r *Router) K8sDeleteHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read cluster id")
		return
	}

	cluster, err := r.db.GetK8s(id)
	if err == gorm.ErrRecordNotFound || cluster.UserID != userID {
		writeErrResponse(req, w, http.StatusNotFound, "Kubernetes cluster not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.cancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract))
	if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.db.DeleteK8s(id)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// metrics
	middlewares.Deletions.WithLabelValues(userID, "k8s").Inc()
	writeMsgResponse(req, w, "kubernetes cluster is deleted successfully", nil)
}

// K8sDeleteAllHandler deletes all clusters for a user
func (r *Router) K8sDeleteAllHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	clusters, err := r.db.GetAllK8s(userID)
	if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
		writeMsgResponse(req, w, "Kubernetes clusters not found", nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	for _, cluster := range clusters {
		err = r.cancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract))
		if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}
	}

	err = r.db.DeleteAllK8s(userID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	for _, c := range clusters {
		middlewares.Deletions.WithLabelValues(c.UserID, "k8s").Inc()
	}
	writeMsgResponse(req, w, "All kubernetes clusters are deleted successfully", nil)
}

func (r *Router) deployK8sRequest(ctx context.Context, user models.User, k8sDeployInput models.K8sDeployInput) (int, error) {
	err := validator.Validate(k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusBadRequest, errors.New("invalid kubernetes data")
	}

	// quota verification
	quota, err := r.db.GetUserQuota(user.ID.String())
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		return http.StatusNotFound, errors.New("user quota is not found")
	}
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	neededQuota, err := validateK8sQuota(k8sDeployInput, quota.Vms, quota.PublicIPs)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		return http.StatusBadRequest, errors.New("ssh key is required")
	}

	// unique names
	available, err := r.db.AvailableK8sName(k8sDeployInput.MasterName)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	if !available {
		return http.StatusBadRequest, errors.New("kubernetes master name is not available, please choose a different name")
	}

	// deploy network and cluster
	node, networkContractID, k8sContractID, err := r.deployK8sClusterWithNetwork(ctx, k8sDeployInput, user.SSHKey)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	k8sCluster, err := r.loadK8s(k8sDeployInput, user.ID.String(), node, networkContractID, k8sContractID)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}
	publicIPsQuota := quota.PublicIPs
	if k8sDeployInput.Public {
		publicIPsQuota -= publicQuota
	}
	// update quota
	err = r.db.UpdateUserQuota(user.ID.String(), quota.Vms-neededQuota, publicIPsQuota)
	if err == gorm.ErrRecordNotFound {
		return http.StatusNotFound, errors.New("user quota is not found")
	}
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	err = r.db.CreateK8s(&k8sCluster)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	// metrics
	middlewares.Deployments.WithLabelValues(user.ID.String(), k8sDeployInput.Resources, "master").Inc()
	for _, worker := range k8sDeployInput.Workers {
		middlewares.Deployments.WithLabelValues(user.ID.String(), worker.Resources, "worker").Inc()
	}

	return 0, nil
}
