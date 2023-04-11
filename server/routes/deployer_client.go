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
	publicQuota = 1

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
	cru, mru, sru, _, err := calcNodeResources(k8sDeployInput.Resources, k8sDeployInput.Public)
	if err != nil {
		return models.K8sCluster{}, err
	}
	master := models.Master{
		CRU:       cru,
		MRU:       mru,
		SRU:       sru,
		Public:    k8sDeployInput.Public,
		PublicIP:  resCluster.Master.ComputedIP,
		Name:      k8sDeployInput.MasterName,
		YggIP:     resCluster.Master.YggIP,
		Resources: k8sDeployInput.Resources,
	}
	workers := []models.Worker{}
	for _, worker := range k8sDeployInput.Workers {

		cru, mru, sru, _, err := calcNodeResources(worker.Resources, false)
		if err != nil {
			return models.K8sCluster{}, err
		}
		workerModel := models.Worker{
			Name:      worker.Name,
			CRU:       cru,
			MRU:       mru,
			SRU:       sru,
			Resources: worker.Resources,
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

func (r *Router) deployVM(ctx context.Context, vmInput DeployVMInput, sshKey string) (*workloads.VM, uint64, uint64, uint64, error) {
	// filter nodes
	filter, err := filterNode(vmInput.Resources, vmInput.Public)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	nodeIDs, err := deployer.FilterNodes(r.tfPluginClient.GridProxyClient, filter)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	nodeID := uint32(nodeIDs[0].NodeID)

	// create network workload
	network := buildNetwork(nodeID, generateNetworkName())

	// create disk
	disk := workloads.Disk{
		Name:   "disk",
		SizeGB: int(*filter.FreeSRU),
	}

	// create vm workload
	vm := workloads.VM{
		Name:      vmInput.Name,
		Flist:     vmFlist,
		CPU:       int(*filter.TotalCRU),
		PublicIP:  vmInput.Public,
		Planetary: true,
		Memory:    int(*filter.FreeMRU) * 1024,
		Mounts: []workloads.Mount{
			{DiskName: disk.Name, MountPoint: "/disk"},
		},
		Entrypoint: "/sbin/zinit init",
		EnvVars: map[string]string{
			"SSH_KEY": sshKey,
		},
		NetworkName: network.Name,
	}

	// deploy network
	err = r.tfPluginClient.NetworkDeployer.Deploy(ctx, &network)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// deploy vm
	dl := workloads.NewDeployment(vmInput.Name, nodeID, "", nil, network.Name, []workloads.Disk{disk}, nil, []workloads.VM{vm}, nil)
	err = r.tfPluginClient.DeploymentDeployer.Deploy(ctx, &dl)
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
func (r *Router) cancelDeployment(contractID uint64, netContractID uint64) error {
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

func calcNodeResources(resources string, public bool) (uint64, uint64, uint64, uint64, error) {
	var cru uint64
	var mru uint64
	var sru uint64
	var ips uint64
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
		return 0, 0, 0, 0, fmt.Errorf("unknown resource type %s", resources)
	}
	if public {
		ips = 1
	}
	return cru, mru, sru, ips, nil
}

func (r *Router) getK8sAvailableNode(k K8sDeployInput) (uint32, error) {
	_, mru, sru, ips, err := calcNodeResources(k.Resources, k.Public)
	if err != nil {
		return 0, err
	}

	for _, worker := range k.Workers {
		_, m, s, _, err := calcNodeResources(worker.Resources, false)
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
		FreeIPs: &ips,
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
func filterNode(resource string, public bool) (types.NodeFilter, error) {
	cru, mru, sru, ips, err := calcNodeResources(resource, public)
	if err != nil {
		return types.NodeFilter{}, err
	}

	return types.NodeFilter{
		TotalCRU: &cru,
		FreeSRU:  &sru,
		FreeMRU:  &mru,
		FreeIPs:  &ips,
		IPv4:     &trueVal,
		Status:   &statusUp,
		IPv6:     &trueVal,
	}, nil
}

func validateK8sQuota(k K8sDeployInput, availableResourcesQuota, availablePublicIPsQuota int) (int, error) {
	neededQuota, err := calcNeededQuota(k.Resources)
	if err != nil {
		return 0, err
	}

	for _, worker := range k.Workers {
		workerQuota, err := calcNeededQuota(worker.Resources)
		if err != nil {
			return 0, err
		}
		neededQuota += workerQuota
	}

	if availableResourcesQuota < neededQuota {
		return 0, fmt.Errorf("no available quota %d for kubernetes deployment, you can request a new voucher", availableResourcesQuota)
	}
	if k.Public && availablePublicIPsQuota < publicQuota {
		return 0, fmt.Errorf("no available quota %d for public ips", availablePublicIPsQuota)
	}

	return neededQuota, nil
}

func validateVMQuota(vm DeployVMInput, availableResourcesQuota, availablePublicIPsQuota int) (int, error) {
	neededQuota, err := calcNeededQuota(vm.Resources)
	if err != nil {
		return 0, err
	}

	if availableResourcesQuota < neededQuota {
		return 0, fmt.Errorf("no available quota %d for deployment for resources %s, you can request a new voucher", availableResourcesQuota, vm.Resources)
	}
	if vm.Public && availablePublicIPsQuota < publicQuota {
		return 0, fmt.Errorf("no available quota %d for public ips", availablePublicIPsQuota)
	}

	return neededQuota, nil
}

func calcNeededQuota(resources string) (int, error) {
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
	cru, mru, sru, ips, err := calcNodeResources(k.Resources, k.Public)
	if err != nil {
		return workloads.K8sCluster{}, err
	}
	master.CPU = int(cru)
	master.Memory = int(mru * 1024)
	master.DiskSize = int(sru)
	if ips == 1 {
		master.PublicIP = true
	}

	workers := []workloads.K8sNode{}
	for _, worker := range k.Workers {
		w := workloads.K8sNode{
			Name:  worker.Name,
			Flist: k8sFlist,
			Node:  node,
		}
		cru, mru, sru, _, err := calcNodeResources(k.Resources, false)
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
