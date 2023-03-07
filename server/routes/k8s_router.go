// Package routes for API endpoints
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/threefoldtech/grid3-go/deployer"
	"github.com/threefoldtech/grid3-go/workloads"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

var (
	k8sFlist        = "https://hub.grid.tf/tf-official-apps/threefoldtech-k3s-latest.flist"
	k8sSmallCpu     = 1
	k8sSmallMemory  = 2
	k8sSmallDisk    = 10
	k8sMediumCpu    = 2
	k8sMediumMemory = 4
	k8sMediumDisk   = 15
	k8sLargeCpu     = 4
	k8sLargeMemory  = 8
	k8sLargeDisk    = 20
)

type K8sDeployInput struct {
	MasterName string   `json:"master_name"`
	Resources  string   `json:"resources"`
	Workers    []Worker `json:"workers"`
}
type Worker struct {
	Name      string `json:"name"`
	Resources string `json:"resources"`
}

type K8sGetResponse struct {
	Master  models.Master   `json:"master"`
	Workers []models.Worker `json:"workers"`
}

func (r *Router) K8sDeployHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: should be a function
	// user authorization
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.config.Token.Secret), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			r.WriteErrResponse(w, err)
			return
		}
		r.WriteErrResponse(w, err)
		return
	}
	if !tkn.Valid {
		r.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		return
	}

	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		r.WriteErrResponse(w, fmt.Errorf("token is expired"))
		return
	}
	var k8sDeployInput K8sDeployInput
	err = json.NewDecoder(req.Body).Decode(&k8sDeployInput)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// TODO: qouta verification

	// get tf plugin client
	client, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, true)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// get available nodes
	node, err := getK8sAvailableNodes(&client, k8sDeployInput)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// build network
	network := buildNetwork(node, k8sDeployInput.MasterName)
	// build cluster
	user, err := r.db.GetUserByID(claims.UserID)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	cluster, err := buildK8sCluster(node,
		user.SSHKey,
		network.Name,
		k8sDeployInput,
	)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	// deploy network and cluster
	err = deployK8sClusterWithNetwork(&client, &cluster, &network)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// load cluster
	masterNode := map[uint32]string{node: k8sDeployInput.MasterName}
	workerNodes := make(map[uint32][]string)
	workers := []string{}
	for _, worker := range k8sDeployInput.Workers {
		workers = append(workers, worker.Name)
	}
	workerNodes[node] = workers
	resCluster, err := client.State.LoadK8sFromGrid(masterNode, workerNodes, k8sDeployInput.MasterName)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// save to db
	k := models.Master{
		UserID:    user.ID.String(),
		Resources: k8sDeployInput.Resources,
		Name:      k8sDeployInput.MasterName,
		IP:        resCluster.Master.YggIP,
	}
	err = r.db.CreateK8s(&k)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	for _, worker := range k8sDeployInput.Workers {
		workerModel := models.Worker{
			ClusterID: k.ID,
			Name:      worker.Name,
			Resources: worker.Resources,
		}
		err := r.db.CreateWorker(&workerModel)
		if err != nil {
			r.WriteErrResponse(w, err)
			return
		}
	}

	// write response
	r.WriteMsgResponse(w, "Kubernetes cluster deployed successfully", map[string]int{"id": k.ID})
}

func (r *Router) K8sGetHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: should be a function
	// user authorization
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.config.Token.Secret), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			r.WriteErrResponse(w, err)
			return
		}
		r.WriteErrResponse(w, err)
		return
	}
	if !tkn.Valid {
		r.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		return
	}

	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		r.WriteErrResponse(w, fmt.Errorf("token is expired"))
		return
	}
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	master, workers, err := r.db.GetK8s(id)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "Kubernets cluster found", K8sGetResponse{
		Master:  master,
		Workers: workers,
	})
}

func buildK8sCluster(node uint32, sshkey, network string, k K8sDeployInput) (workloads.K8sCluster, error) {
	master := workloads.K8sNode{
		Name:      k.MasterName,
		Flist:     k8sFlist,
		Planetary: true,
		Node:      node,
	}
	switch k.Resources {
	case "small":
		master.CPU = k8sSmallCpu
		master.Memory = k8sSmallMemory * 1024
		master.DiskSize = k8sSmallDisk
	case "medium":
		master.CPU = k8sMediumCpu
		master.Memory = k8sMediumMemory * 1024
		master.DiskSize = k8sMediumDisk
	case "large":
		master.CPU = k8sLargeCpu
		master.Memory = k8sLargeMemory * 1024
		master.DiskSize = k8sLargeDisk
	default:
		return workloads.K8sCluster{}, fmt.Errorf("unknown master resource type %s", k.Resources)
	}
	workers := []workloads.K8sNode{}
	for _, worker := range k.Workers {
		w := workloads.K8sNode{
			Name:  worker.Name,
			Flist: k8sFlist,
			Node:  node,
		}
		switch worker.Resources {
		case "small":
			w.CPU = k8sSmallCpu
			w.Memory = k8sSmallMemory * 1024
			w.DiskSize = k8sSmallDisk
		case "medium":
			w.CPU = k8sMediumCpu
			w.Memory = k8sMediumMemory * 1024
			w.DiskSize = k8sMediumDisk
		case "large":
			w.CPU = k8sLargeCpu
			w.Memory = k8sLargeMemory * 1024
			w.DiskSize = k8sLargeDisk
		default:
			return workloads.K8sCluster{}, fmt.Errorf("unknown w resource type %s", k.Resources)
		}
	}
	k8sCluster := workloads.K8sCluster{
		Master:      &master,
		Workers:     workers,
		NetworkName: network,
		// TODO: random token
		Token: "nottoken",
		// TODO: sshkey
		SSHKey:       sshkey,
		SolutionType: k.MasterName,
	}

	return k8sCluster, nil
}

func buildNetwork(node uint32, name string) workloads.ZNet {
	return workloads.ZNet{
		Name:  name,
		Nodes: []uint32{node},
		IPRange: gridtypes.NewIPNet(net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}),
	}
}

func deployK8sClusterWithNetwork(tfPluginClient *deployer.TFPluginClient, cluster *workloads.K8sCluster, network *workloads.ZNet) error {
	err := tfPluginClient.NetworkDeployer.Deploy(context.Background(), network)
	if err != nil {
		return errors.Wrapf(err, "failed to deploy network on nodes %v", network.Nodes)
	}
	err = tfPluginClient.K8sDeployer.Deploy(context.Background(), cluster)
	if err != nil {
		return errors.Wrapf(err, "failed to deploy kubernetes cluster on nodes %v", network.Nodes)
	}
	return nil
}

func getK8sAvailableNodes(tfPluginClient *deployer.TFPluginClient, k K8sDeployInput) (uint32, error) {
	status := "up"
	freeMRU := uint64(0)
	freeHRU := uint64(0)
	ipv6 := true
	switch k.Resources {
	case "small":
		freeMRU += uint64(k8sSmallMemory)
		freeHRU += uint64(k8sSmallDisk)
	case "medium":
		freeMRU += uint64(k8sMediumMemory)
		freeHRU += uint64(k8sMediumDisk)
	case "large":
		freeMRU += uint64(k8sLargeMemory)
		freeHRU += uint64(k8sLargeDisk)
	default:
		return 0, fmt.Errorf("unknown master resource type %s", k.Resources)
	}
	for _, worker := range k.Workers {
		switch worker.Resources {
		case "small":
			freeMRU += uint64(k8sSmallMemory)
			freeHRU += uint64(k8sSmallDisk)
		case "medium":
			freeMRU += uint64(k8sMediumMemory)
			freeHRU += uint64(k8sMediumDisk)
		case "large":
			freeMRU += uint64(k8sLargeMemory)
			freeHRU += uint64(k8sLargeDisk)
		default:
			return 0, fmt.Errorf("unknown w resource type %s", k.Resources)
		}
	}
	filter := types.NodeFilter{
		Status:  &status,
		FreeMRU: &freeHRU,
		FreeHRU: &freeHRU,
		FarmIDs: []uint64{1},
		IPv6:    &ipv6,
	}
	nodes, _, err := tfPluginClient.GridProxyClient.Nodes(filter, types.Limit{})
	if err != nil {
		return 0, err
	}
	if len(nodes) == 0 {
		return 0, fmt.Errorf(
			"no node with free resources available using node filter: farmIDs: %v, mru: %d, hru: %d",
			filter.FarmIDs,
			*filter.FreeMRU,
			*filter.FreeHRU,
		)
	}
	return uint32(nodes[0].NodeID), nil
}
