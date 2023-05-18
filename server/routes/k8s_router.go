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

	err = validator.Validate(k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "invalid kubernetes data")
		return
	}

	// quota verification
	quota, err := r.db.GetUserQuota(user.ID.String())
	if err == gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusNotFound, "user quota is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	_, err = deployer.ValidateK8sQuota(k8sDeployInput, quota.Vms, quota.PublicIPs)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, err.Error())
		return
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		writeErrResponse(req, w, http.StatusBadRequest, "ssh key is required")
	}

	// unique names
	available, err := r.db.AvailableK8sName(k8sDeployInput.MasterName)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if !available {
		writeErrResponse(req, w, http.StatusBadRequest, "kubernetes master name is not available, please choose a different name")
		return
	}

	err = r.deployer.Redis.PushK8sRequest(streams.K8sDeployRequest{User: user, Input: k8sDeployInput, AdminSSHKey: r.config.AdminSSHKey})
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// write response
	writeMsgResponse(req, w, "Kubernetes cluster request is being deployed, you'll receive a confirmation notification soon", "")
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

	err = r.deployer.CancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract), "k8s", cluster.Master.Name)
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
		err = r.deployer.CancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract), "k8s", cluster.Master.Name)
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
