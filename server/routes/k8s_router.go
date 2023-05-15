// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
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

	err = r.Deployer.Redis.PushK8sRequest(streams.K8sDeployRequest{User: user, Input: k8sDeployInput})
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

	err = r.Deployer.CancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract))
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
		err = r.Deployer.CancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract))
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
