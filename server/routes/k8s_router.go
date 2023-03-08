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
	smallK8sQouta   = 1
	mediumK8sQouta  = 2
	largeK8sQouta   = 3
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

func (r *Router) K8sDeployHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	claims, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	var k8sDeployInput K8sDeployInput
	err = json.NewDecoder(req.Body).Decode(&k8sDeployInput)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// qouta verification
	quota, err := r.db.GetUserQuota(claims.UserID)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	neededQuota, err := calcNeededQuota(k8sDeployInput)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	if neededQuota > quota.K8s {
		r.WriteErrResponse(w, fmt.Errorf("qouta not enough need %d available %d", neededQuota, quota.K8s))
		return
	}

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
	// update quota
	err = r.db.UpdateUserQuota(claims.UserID, quota.Vms, quota.K8s-neededQuota)
	if err != nil {
		r.WriteErrResponse(w, err)
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
		r.WriteErrResponse(w, err)
		return
	}

	// save to db

	cru, mru, sru, err := calcK8sNodeResources(k8sDeployInput.Resources)
	if err != nil {
		r.WriteErrResponse(w, err)
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
			r.WriteErrResponse(w, err)
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
		r.WriteErrResponse(w, err)
		return
	}

	// write response
	r.WriteMsgResponse(w, "Kubernetes cluster deployed successfully", map[string]int{"id": kCluster.ID})
}

func (r *Router) K8sGetHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	claims, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	cluster, err := r.db.GetK8s(id)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	if cluster.UserID != claims.UserID {
		r.WriteErrResponse(w, errors.New("invalid user"))
		return
	}
	r.WriteMsgResponse(w, "Kubernets cluster found", cluster)
}
func (r *Router) K8sGetAllHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	claims, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	clusters, err := r.db.GetAllK8s(claims.UserID)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "Kubernets clusters found", clusters)
}

func (r *Router) K8sDeleteHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	claims, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	cluster, err := r.db.GetK8s(id)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	if cluster.UserID != claims.UserID {
		r.WriteErrResponse(w, errors.New("invalid user"))
		return
	}

	client, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, true)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	err = client.SubstrateConn.CancelContract(client.Identity, uint64(cluster.ClusterContract))
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	err = client.SubstrateConn.CancelContract(client.Identity, uint64(cluster.NetworkContract))
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	err = r.db.DeleteK8s(id)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "Deleted succesfully", nil)

}

func buildK8sCluster(node uint32, sshkey, network string, k K8sDeployInput) (workloads.K8sCluster, error) {
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

func calcK8sNodeResources(resources string) (int, int, int, error) {
	var cru int
	var mru int
	var sru int
	switch resources {
	case "small":
		cru += k8sSmallCpu
		mru += k8sSmallMemory
		sru += k8sSmallDisk
	case "medium":
		cru += k8sMediumCpu
		mru += k8sMediumMemory
		sru += k8sMediumDisk
	case "large":
		cru += k8sLargeCpu
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
		k8sNeeded += smallK8sQouta
	case "medium":
		k8sNeeded += mediumK8sQouta
	case "large":
		k8sNeeded += largeK8sQouta
	default:
		return 0, fmt.Errorf("unknown master resource type %s", k.Resources)
	}
	for _, worker := range k.Workers {
		switch worker.Resources {
		case "small":
			k8sNeeded += smallK8sQouta
		case "medium":
			k8sNeeded += mediumK8sQouta
		case "large":
			k8sNeeded += largeK8sQouta
		default:
			return 0, fmt.Errorf("unknown w resource type %s", k.Resources)
		}
	}
	return k8sNeeded, nil
}
