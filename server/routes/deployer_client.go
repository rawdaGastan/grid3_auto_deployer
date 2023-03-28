// Package routes for API endpoints
package routes

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/codescalers/cloud4students/models"
	"github.com/pkg/errors"
	"github.com/threefoldtech/grid3-go/deployer"
	"github.com/threefoldtech/grid3-go/workloads"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

var (
	k8sFlist = "https://hub.grid.tf/tf-official-apps/threefoldtech-k3s-latest.flist"
	vmFlist  = "https://hub.grid.tf/tf-official-apps/base:latest.flist"

	smallCPU     = uint64(1)
	smallMemory  = uint64(2)
	smallDisk    = uint64(5)
	mediumCPU    = uint64(2)
	mediumMemory = uint64(4)
	mediumDisk   = uint64(10)
	largeCPU     = uint64(4)
	largeMemory  = uint64(8)
	largeDisk    = uint64(15)

	smallQuota  = 1
	mediumQuota = 2
	largeQuota  = 3

	trueVal  = true
	statusUp = "up"

	token = "random"
)

func (r *Router) deployK8sClusterWithNetwork(k8sDeployInput K8sDeployInput, sshKey string) (uint32, uint64, uint64, error) {
	// get available nodes
	node, err := r.getK8sAvailableNode(k8sDeployInput)
	if err != nil {
		return 0, 0, 0, err
	}

	// build network
	network := buildNetwork(node, generateNetworkName())

	// build cluster
	cluster, err := buildK8sCluster(node,
		sshKey,
		network.Name,
		k8sDeployInput,
	)
	if err != nil {
		return 0, 0, 0, err
	}

	err = r.tfPluginClient.NetworkDeployer.Deploy(context.Background(), &network)
	if err != nil {
		return 0, 0, 0, errors.Wrapf(err, "failed to deploy network on nodes %v", network.Nodes)
	}

	err = r.tfPluginClient.K8sDeployer.Deploy(context.Background(), &cluster)
	if err != nil {
		return 0, 0, 0, errors.Wrapf(err, "failed to deploy kubernetes cluster on nodes %v", network.Nodes)
	}

	return node, network.NodeDeploymentID[node], cluster.NodeDeploymentID[node], nil
}

func (r *Router) loadK8s(k8sDeployInput K8sDeployInput, userID string, node uint32, networkContractID uint64, k8sContractID uint64) (models.K8sCluster, error) {
	// load cluster
	masterNode := map[uint32]string{node: k8sDeployInput.MasterName}
	workerNodes := make(map[uint32][]string)
	workersNames := []string{}
	for _, worker := range k8sDeployInput.Workers {
		workersNames = append(workersNames, worker.Name)
	}
	workerNodes[node] = workersNames
	resCluster, err := r.tfPluginClient.State.LoadK8sFromGrid(masterNode, workerNodes, k8sDeployInput.MasterName)
	if err != nil {
		return models.K8sCluster{}, err
	}

	// save to db
	cru, mru, sru, err := calcNodeResources(k8sDeployInput.Resources)
	if err != nil {
		return models.K8sCluster{}, err
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

		cru, mru, sru, err := calcNodeResources(k8sDeployInput.Resources)
		if err != nil {
			return models.K8sCluster{}, err
		}
		workerModel := models.Worker{
			Name: worker.Name,
			CRU:  cru,
			MRU:  mru,
			SRU:  sru,
		}
		workers = append(workers, workerModel)
	}
	k8sCluster := models.K8sCluster{
		UserID:          userID,
		NetworkContract: int(networkContractID),
		ClusterContract: int(k8sContractID),
		Master:          master,
		Workers:         workers,
	}

	return k8sCluster, nil
}

func (r *Router) deployVM(vmName, resources, sshKey string) (*workloads.VM, uint64, uint64, uint64, error) {
	// filter nodes
	filter, err := filterNode(resources)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	nodeIDs, err := deployer.FilterNodes(r.tfPluginClient.GridProxyClient, filter)
	fmt.Printf("nodeIDs: %v\n", nodeIDs)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	nodeID := uint32(nodeIDs[1].NodeID)

	// create network workload
	network := buildNetwork(nodeID, generateNetworkName())

	// create disk
	disk := workloads.Disk{
		Name:   "disk",
		SizeGB: int(*filter.TotalSRU),
	}

	// create vm workload
	vm := workloads.VM{
		Name:      vmName,
		Flist:     vmFlist,
		CPU:       int(*filter.TotalCRU),
		PublicIP:  false,
		Planetary: true,
		Memory:    int(*filter.TotalMRU) * 1024,
		Mounts: []workloads.Mount{
			{DiskName: disk.Name, MountPoint: "/disk"},
		},
		Entrypoint: "/sbin/zinit init",
		EnvVars: map[string]string{
			"SSH_KEY": sshKey,
		},
		NetworkName: network.Name,
	}

	// TODO: set proper contexts
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.config.Token.Timeout)*time.Minute)
	defer cancel()
	print("after ctx")

	// deploy network
	err = r.tfPluginClient.NetworkDeployer.Deploy(ctx, &network)
	fmt.Printf("deploy network err: %v\n", err)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	znet, err := r.tfPluginClient.State.LoadNetworkFromGrid(network.Name)
	fmt.Printf("znet: %v\n", znet)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	// deploy vm
	dl := workloads.NewDeployment("vm", nodeID, "", nil, network.Name, []workloads.Disk{disk}, nil, []workloads.VM{vm}, nil)
	err = r.tfPluginClient.DeploymentDeployer.Deploy(ctx, &dl)
	fmt.Printf("deploy vm err: %v\n", err)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// checks that vm deployed successfully
	loadedVM, err := r.tfPluginClient.State.LoadVMFromGrid(nodeID, vm.Name, dl.Name)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	return &loadedVM, dl.ContractID, network.NodeDeploymentID[nodeID], uint64(disk.SizeGB), nil
}

// CancelDeployment cancel deployments from grid
func (r *Router) CancelDeployment(contractID uint64, netContractID uint64) error {
	// cancel deployment
	err := r.tfPluginClient.SubstrateConn.CancelContract(r.tfPluginClient.Identity, contractID)
	if err != nil {
		return err
	}

	// cancel network
	err = r.tfPluginClient.SubstrateConn.CancelContract(r.tfPluginClient.Identity, netContractID)
	if err != nil {
		return err
	}

	return nil
}

func calcNodeResources(resources string) (uint64, uint64, uint64, error) {
	var cru uint64
	var mru uint64
	var sru uint64
	switch resources {
	case "small":
		cru += smallCPU
		mru += smallMemory
		sru += smallDisk
	case "medium":
		cru += mediumCPU
		mru += mediumMemory
		sru += mediumDisk
	case "large":
		cru += largeCPU
		mru += largeMemory
		sru += largeDisk
	default:
		return 0, 0, 0, fmt.Errorf("unknown resource type %s", resources)
	}
	return cru, mru, sru, nil
}

func (r *Router) getK8sAvailableNode(k K8sDeployInput) (uint32, error) {
	_, mru, sru, err := calcNodeResources(k.Resources)
	if err != nil {
		return 0, err
	}

	for _, worker := range k.Workers {
		_, m, s, err := calcNodeResources(worker.Resources)
		if err != nil {
			return 0, err
		}
		mru += m
		sru += s
	}

	freeMRU := uint64(mru)
	freeSRU := uint64(sru)
	filter := types.NodeFilter{
		Status:  &statusUp,
		FreeMRU: &freeMRU,
		FreeSRU: &freeSRU,
		FarmIDs: []uint64{1},
		IPv6:    &trueVal,
	}

	nodes, err := deployer.FilterNodes(r.tfPluginClient.GridProxyClient, filter)
	if err != nil {
		return 0, err
	}

	return uint32(nodes[0].NodeID), nil
}

// choose suitable nodes based on needed resources
func filterNode(resource string) (types.NodeFilter, error) {
	cru, mru, sru, err := calcNodeResources(resource)
	if err != nil {
		return types.NodeFilter{}, err
	}

	return types.NodeFilter{
		TotalCRU: &cru,
		TotalSRU: &sru,
		TotalMRU: &mru,
		Status:   &statusUp,
		IPv6:     &trueVal,
	}, nil
}

func validateK8sQuota(k K8sDeployInput, availableQuota int) (int, error) {
	neededQuota, err := calcNeededQuota(k.Resources, availableQuota)
	if err != nil {
		return 0, err
	}

	for _, worker := range k.Workers {
		workerQuota, err := calcNeededQuota(worker.Resources, availableQuota)
		if err != nil {
			return 0, err
		}
		neededQuota += workerQuota
	}

	if availableQuota < neededQuota {
		return 0, fmt.Errorf("no available quota %v for kubernetes deployment", availableQuota)
	}

	return neededQuota, nil
}

func validateVMQuota(resources string, availableQuota int) (int, error) {
	neededQuota, err := calcNeededQuota(resources, availableQuota)
	if err != nil {
		return 0, err
	}

	if availableQuota < neededQuota {
		return 0, fmt.Errorf("no available quota %v for deployment for resources %s", availableQuota, resources)
	}

	return neededQuota, nil
}

func calcNeededQuota(resources string, availableQuota int) (int, error) {
	var neededQuota int
	switch resources {
	case "small":
		neededQuota += smallQuota
	case "medium":
		neededQuota += mediumQuota
	case "large":
		neededQuota += largeQuota
	default:
		return 0, fmt.Errorf("unknown resource type %s", resources)
	}

	return neededQuota, nil
}

func buildNetwork(node uint32, name string) workloads.ZNet {
	return workloads.ZNet{
		Name:  name,
		Nodes: []uint32{node},
		IPRange: gridtypes.NewIPNet(net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}),
		AddWGAccess: false,
	}
}

func buildK8sCluster(node uint32, sshKey, network string, k K8sDeployInput) (workloads.K8sCluster, error) {
	master := workloads.K8sNode{
		Name:      k.MasterName,
		Flist:     k8sFlist,
		Planetary: true,
		Node:      node,
	}
	cru, mru, sru, err := calcNodeResources(k.Resources)
	if err != nil {
		return workloads.K8sCluster{}, err
	}
	master.CPU = int(cru)
	master.Memory = int(mru * 1024)
	master.DiskSize = int(sru)

	workers := []workloads.K8sNode{}
	for _, worker := range k.Workers {
		w := workloads.K8sNode{
			Name:  worker.Name,
			Flist: k8sFlist,
			Node:  node,
		}
		cru, mru, sru, err := calcNodeResources(k.Resources)
		if err != nil {
			return workloads.K8sCluster{}, err
		}
		w.CPU = int(cru)
		w.Memory = int(mru * 1024)
		w.DiskSize = int(sru)
		workers = append(workers, w)
	}
	k8sCluster := workloads.K8sCluster{
		Master:       &master,
		Workers:      workers,
		NetworkName:  network,
		Token:        token,
		SSHKey:       sshKey,
		SolutionType: k.MasterName,
	}

	return k8sCluster, nil
}

// generate random names for network
func generateNetworkName() string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	name := make([]byte, 4)
	for i := range name {
		name[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(name)
}
