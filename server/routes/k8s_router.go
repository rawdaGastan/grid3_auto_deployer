// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rawdaGastan/cloud4students/middlewares"
)

// K8sDeployInput deploy k8s cluster input
type K8sDeployInput struct {
	MasterName string   `json:"master_name"`
	Resources  string   `json:"resources"`
	Workers    []Worker `json:"workers"`
}

// Worker deploy k8s worker input
type Worker struct {
	Name      string `json:"name"`
	Resources string `json:"resources"`
}

// K8sDeployHandler deploy k8s handler
func (r *Router) K8sDeployHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	var k8sDeployInput K8sDeployInput
	err = json.NewDecoder(req.Body).Decode(&k8sDeployInput)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// quota verification
	quota, err := r.db.GetUserQuota(userID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	neededQuota, err := validateK8sQuota(k8sDeployInput, quota.K8s)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	if len(strings.TrimSpace(user.SSHKey)) == 0 {
		writeErrResponse(w, fmt.Errorf("ssh key is required"))
		return
	}

	// deploy network and cluster
	node, networkContractID, k8sContractID, err := r.deployK8sClusterWithNetwork(k8sDeployInput, user.SSHKey)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	k8sCluster, err := r.loadK8s(k8sDeployInput, userID, node, networkContractID, k8sContractID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// update quota
	err = r.db.UpdateUserQuota(userID, quota.Vms, quota.K8s-neededQuota)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = r.db.CreateK8s(&k8sCluster)
	if err != nil {
		writeErrResponse(w, err)
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
		writeErrResponse(w, err)
		return
	}

	cluster, err := r.db.GetK8s(id)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if cluster.UserID != userID {
		writeNotFoundResponse(w, errors.New("invalid user"))
		return
	}
	writeMsgResponse(w, "Kubernetes cluster is found", cluster)
}

// K8sGetAllHandler gets all clusters for a user
func (r *Router) K8sGetAllHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	clusters, err := r.db.GetAllK8s(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	writeMsgResponse(w, "Kubernetes clusters are found", clusters)
}

// K8sDeleteHandler deletes a cluster for a user
func (r *Router) K8sDeleteHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	cluster, err := r.db.GetK8s(id)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if cluster.UserID != userID {
		writeNotFoundResponse(w, errors.New("invalid user"))
		return
	}

	err = r.cancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract))
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = r.db.DeleteK8s(id)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	writeMsgResponse(w, "kubernetes cluster is deleted successfully", nil)
}

// K8sDeleteAllHandler deletes all clusters for a user
func (r *Router) K8sDeleteAllHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	clusters, err := r.db.GetAllK8s(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	for _, cluster := range clusters {
		err = r.cancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract))
		if err != nil {
			writeErrResponse(w, err)
			return
		}
	}

	err = r.db.DeleteAllK8s(userID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	writeMsgResponse(w, "All kubernetes clusters are deleted successfully", nil)
}
