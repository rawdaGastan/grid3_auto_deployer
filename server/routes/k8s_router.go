// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

// K8sDeployInput deploy k8s cluster input
type K8sDeployInput struct {
	MasterName string   `json:"master_name" validate:"min=3,max=20"`
	Resources  string   `json:"resources"`
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
		writeErrResponse(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	var k8sDeployInput K8sDeployInput
	err = json.NewDecoder(req.Body).Decode(&k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusBadRequest, "Failed to read k8s data")
		return
	}
	err = validator.Validate(k8sDeployInput)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusBadRequest, "Invalid Kubernetes data")
		return
	}

	// quota verification
	quota, err := r.db.GetUserQuota(userID)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(w, http.StatusNotFound, "User quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	neededQuota, err := validateK8sQuota(k8sDeployInput, quota.Vms)
	if err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		writeErrResponse(w, http.StatusBadRequest, "SSH key is required")
		return
	}

	// deploy network and cluster
	node, networkContractID, k8sContractID, err := r.deployK8sClusterWithNetwork(k8sDeployInput, user.SSHKey)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	k8sCluster, err := r.loadK8s(k8sDeployInput, userID, node, networkContractID, k8sContractID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// update quota
	err = r.db.UpdateUserQuota(userID, quota.Vms-neededQuota)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(w, http.StatusNotFound, "User quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.db.CreateK8s(&k8sCluster)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// write response
	writeMsgResponse(w, "Kubernetes cluster is deployed successfully", map[string]int{"id": k8sCluster.ID})
}

// K8sGetHandler gets a cluster for a user
func (r *Router) K8sGetHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, "Failed to read cluster id")
		return
	}

	cluster, err := r.db.GetK8s(id)
	if err == gorm.ErrRecordNotFound || cluster.UserID != userID {
		writeErrResponse(w, http.StatusNotFound, "Kubernetes cluster not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Kubernetes cluster found", cluster)
}

// K8sGetAllHandler gets all clusters for a user
func (r *Router) K8sGetAllHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	clusters, err := r.db.GetAllK8s(userID)
	if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
		writeMsgResponse(w, "Kubernetes clusters not found", nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "Kubernetes clusters found", clusters)
}

// K8sDeleteHandler deletes a cluster for a user
func (r *Router) K8sDeleteHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, "Failed to read cluster id")
		return
	}

	cluster, err := r.db.GetK8s(id)
	if err == gorm.ErrRecordNotFound || cluster.UserID != userID {
		writeErrResponse(w, http.StatusNotFound, "Kubernetes cluster not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.cancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract))
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.db.DeleteK8s(id)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	writeMsgResponse(w, "kubernetes cluster is deleted successfully", nil)
}

// K8sDeleteAllHandler deletes all clusters for a user
func (r *Router) K8sDeleteAllHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	clusters, err := r.db.GetAllK8s(userID)
	if err == gorm.ErrRecordNotFound || len(clusters) == 0 {
		writeMsgResponse(w, "Kubernetes clusters not found", nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	for _, cluster := range clusters {
		err = r.cancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract))
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}
	}

	err = r.db.DeleteAllK8s(userID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(w, "All kubernetes clusters are deleted successfully", nil)
}
