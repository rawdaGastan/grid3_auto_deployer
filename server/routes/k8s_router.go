// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

// K8sDeployInput deploy k8s cluster input
type K8sDeployInput struct {
	MasterName string   `json:"master_name" validate:"min=3,max=20"`
	Resources  string   `json:"resources"`
	Public     bool     `json:"public"`
	Workers    []Worker `json:"workers"`
}

// Worker deploy k8s worker input
type Worker struct {
	Name      string `json:"name" validate:"min=3,max=20"`
	Resources string `json:"resources"`
}

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

	var k8sDeployInput K8sDeployInput
	err = json.NewDecoder(req.Body).Decode(&k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read k8s data")
		return
	}
	err = validator.Validate(k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Invalid Kubernetes data")
		return
	}

	// quota verification
	quota, err := r.db.GetUserQuota(userID)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	neededQuota, err := validateK8sQuota(k8sDeployInput, quota.Vms, quota.PublicIPs)
	if err != nil {
		writeErrResponse(req, w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		writeErrResponse(req, w, http.StatusBadRequest, "SSH key is required")
		return
	}

	// unique names
	available, err := r.db.AvailableK8sName(k8sDeployInput.MasterName)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if !available {
		writeErrResponse(req, w, http.StatusBadRequest, "Kubernetes master name is not available, please choose a different name")
		return
	}

	// deploy network and cluster
	node, networkContractID, k8sContractID, err := r.deployK8sClusterWithNetwork(k8sDeployInput, user.SSHKey)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	k8sCluster, err := r.loadK8s(k8sDeployInput, userID, node, networkContractID, k8sContractID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	publicIPsQuota := quota.PublicIPs
	if k8sDeployInput.Public {
		publicIPsQuota -= publicQuota
	}
	// update quota
	err = r.db.UpdateUserQuota(userID, quota.Vms-neededQuota, publicIPsQuota)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.db.CreateK8s(&k8sCluster)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// metrics
	middlewares.Deployments.WithLabelValues(userID, k8sDeployInput.Resources, "master").Inc()
	for _, worker := range k8sDeployInput.Workers {
		middlewares.Deployments.WithLabelValues(userID, worker.Resources, "worker").Inc()
	}

	// write response
	writeMsgResponse(req, w, "Kubernetes cluster is deployed successfully", map[string]int{"id": k8sCluster.ID})
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
	if err != nil {
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
		if err != nil {
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
