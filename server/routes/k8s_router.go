// Package routes for API endpoints
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

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
	k8sSmallCPU     = 1
	k8sSmallMemory  = 2
	k8sSmallDisk    = 5
	k8sMediumCPU    = 2
	k8sMediumMemory = 4
	k8sMediumDisk   = 10
	k8sLargeCPU     = 4
	k8sLargeMemory  = 8
	k8sLargeDisk    = 15
	smallK8sQuota   = 1
	mediumK8sQuota  = 2
	largeK8sQuota   = 3
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
	userID := req.Context().Value("UserID").(string)

	var k8sDeployInput K8sDeployInput
	err := json.NewDecoder(req.Body).Decode(&k8sDeployInput)
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

	neededQuota, err := calcNeededQuota(k8sDeployInput)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	if neededQuota > quota.K8s {
		writeErrResponse(w, fmt.Errorf("quota is not enough need %d available %d", neededQuota, quota.K8s))
		return
	}

	// get tf plugin client
	client, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, true)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// get available nodes
	node, err := getK8sAvailableNodes(&client, k8sDeployInput)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// build network
	network := buildNetwork(node, k8sDeployInput.MasterName)
	// build cluster
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	cluster, err := buildK8sCluster(node,
		user.SSHKey,
		network.Name,
		k8sDeployInput,
	)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	// deploy network and cluster
	err = deployK8sClusterWithNetwork(&client, &cluster, &network)
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

	// load cluster
	masterNode := map[uint32]string{node: k8sDeployInput.MasterName}
	workerNodes := make(map[uint32][]string)
	workersNames := []string{}
	for _, worker := range k8sDeployInput.Workers {
		workersNames = append(workersNames, worker.Name)
	}
	workerNodes[node] = workersNames
	resCluster, err := client.State.LoadK8sFromGrid(masterNode, workerNodes, k8sDeployInput.MasterName)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// save to db
	cru, mru, sru, err := calcK8sNodeResources(k8sDeployInput.Resources)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	master := models.Master{
		CRU:  cru,
		MRU:  mru,
		SRU:  sru,
		Name: k8sDeployInput.MasterName,
		IP:   resCluster.Master.YggIP,
	}
	workers := []models.Worker{}
	for _, worker := range k8sDeployInput.Workers {

		cru, mru, sru, err := calcK8sNodeResources(k8sDeployInput.Resources)
		if err != nil {
			writeErrResponse(w, err)
			return
		}
		workerModel := models.Worker{
			Name: worker.Name,
			CRU:  cru,
			MRU:  mru,
			SRU:  sru,
		}
		workers = append(workers, workerModel)
	}
	kCluster := models.K8sCluster{
		UserID:          user.ID.String(),
		NetworkContract: int(network.NodeDeploymentID[node]),
		ClusterContract: int(cluster.NodeDeploymentID[node]),
		Master:          master,
		Workers:         workers,
	}

	err = r.db.CreateK8s(&kCluster)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// write response
	writeMsgResponse(w, "Kubernetes cluster deployed successfully", map[string]int{"id": kCluster.ID})
}

// K8sGetHandler gets a cluster for a user
func (r *Router) K8sGetHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("UserID").(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	cluster, err := r.db.GetK8s(id)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	if cluster.UserID != userID {
		writeErrResponse(w, errors.New("invalid user"))
		return
	}
	writeMsgResponse(w, "Kubernetes cluster found", cluster)
}

// K8sGetAllHandler gets all clusters for a user
func (r *Router) K8sGetAllHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("UserID").(string)

	clusters, err := r.db.GetAllK8s(userID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	writeMsgResponse(w, "Kubernetes clusters found", clusters)
}

// K8sDeleteHandler deletes a cluster for a user
func (r *Router) K8sDeleteHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("UserID").(string)
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	cluster, err := r.db.GetK8s(id)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	if cluster.UserID != userID {
		writeErrResponse(w, errors.New("invalid user"))
		return
	}

	client, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, true)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = client.SubstrateConn.CancelContract(client.Identity, uint64(cluster.ClusterContract))
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = client.SubstrateConn.CancelContract(client.Identity, uint64(cluster.NetworkContract))
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = r.db.DeleteK8s(id)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	writeMsgResponse(w, "Cluster is deleted successfully", nil)
}

// K8sDeleteAllHandler deletes all clusters for a user
func (r *Router) K8sDeleteAllHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("UserID").(string)

	client, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, true)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	clusters, err := r.db.GetAllK8s(userID)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	for _, cluster := range clusters {
		err = client.SubstrateConn.CancelContract(client.Identity, uint64(cluster.ClusterContract))
		if err != nil {
			writeErrResponse(w, err)
			return
		}
		err = client.SubstrateConn.CancelContract(client.Identity, uint64(cluster.NetworkContract))
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

	writeMsgResponse(w, "Deleted successfully", nil)
}

func buildK8sCluster(node uint32, sshKey, network string, k K8sDeployInput) (workloads.K8sCluster, error) {
	master := workloads.K8sNode{
		Name:      k.MasterName,
		Flist:     k8sFlist,
		Planetary: true,
		Node:      node,
	}
	cru, mru, sru, err := calcK8sNodeResources(k.Resources)
	if err != nil {
		return workloads.K8sCluster{}, err
	}
	master.CPU = cru
	master.Memory = mru * 1024
	master.DiskSize = sru

	workers := []workloads.K8sNode{}
	for _, worker := range k.Workers {
		w := workloads.K8sNode{
			Name:  worker.Name,
			Flist: k8sFlist,
			Node:  node,
		}
		cru, mru, sru, err := calcK8sNodeResources(k.Resources)
		if err != nil {
			return workloads.K8sCluster{}, err
		}
		w.CPU = cru
		w.Memory = mru * 1024
		w.DiskSize = sru
		workers = append(workers, w)
	}
	k8sCluster := workloads.K8sCluster{
		Master:      &master,
		Workers:     workers,
		NetworkName: network,
		// TODO: random token
		Token:        "nottoken",
		SSHKey:       sshKey,
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

func calcK8sNodeResources(resources string) (int, int, int, error) {
	var cru int
	var mru int
	var sru int
	switch resources {
	case "small":
		cru += k8sSmallCPU
		mru += k8sSmallMemory
		sru += k8sSmallDisk
	case "medium":
		cru += k8sMediumCPU
		mru += k8sMediumMemory
		sru += k8sMediumDisk
	case "large":
		cru += k8sLargeCPU
		mru += k8sLargeMemory
		sru += k8sLargeDisk
	default:
		return 0, 0, 0, fmt.Errorf("unknown master resource type %s", resources)
	}
	return cru, mru, sru, nil
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
	freeSRU := uint64(0)
	ipv6 := true
	_, mru, sru, err := calcK8sNodeResources(k.Resources)
	if err != nil {
		return 0, err
	}
	for _, worker := range k.Workers {
		_, m, s, err := calcK8sNodeResources(worker.Resources)
		if err != nil {
			return 0, err
		}
		mru += m
		sru += s
	}
	freeMRU = uint64(mru)
	freeSRU = uint64(sru)
	filter := types.NodeFilter{
		Status:  &status,
		FreeMRU: &freeMRU,
		FreeSRU: &freeSRU,
		FarmIDs: []uint64{1},
		IPv6:    &ipv6,
	}
	nodes, _, err := tfPluginClient.GridProxyClient.Nodes(filter, types.Limit{})
	if err != nil {
		return 0, err
	}
	if len(nodes) == 0 {
		return 0, fmt.Errorf(
			"no node with free resources available using node filter: farmIDs: %v, mru: %d, sru: %d",
			filter.FarmIDs,
			*filter.FreeMRU,
			*filter.FreeSRU,
		)
	}
	return uint32(nodes[0].NodeID), nil
}

func calcNeededQuota(k K8sDeployInput) (int, error) {
	var k8sNeeded int
	switch k.Resources {
	case "small":
		k8sNeeded += smallK8sQuota
	case "medium":
		k8sNeeded += mediumK8sQuota
	case "large":
		k8sNeeded += largeK8sQuota
	default:
		return 0, fmt.Errorf("unknown master resource type %s", k.Resources)
	}
	for _, worker := range k.Workers {
		switch worker.Resources {
		case "small":
			k8sNeeded += smallK8sQuota
		case "medium":
			k8sNeeded += mediumK8sQuota
		case "large":
			k8sNeeded += largeK8sQuota
		default:
			return 0, fmt.Errorf("unknown w resource type %s", k.Resources)
		}
	}
	return k8sNeeded, nil
}
