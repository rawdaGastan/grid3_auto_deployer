// Package deployer for handling deployments
package deployer

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

func buildK8sCluster(node uint32, sshKey, network string, k models.K8sCluster) (workloads.K8sCluster, error) {
	myceliumIPSeed, err := workloads.RandomMyceliumIPSeed()
	if err != nil {
		return workloads.K8sCluster{}, err
	}

	master := workloads.K8sNode{
		VM: &workloads.VM{
			Name:           k.Master.Name,
			Flist:          k8sFlist,
			Planetary:      true,
			MyceliumIPSeed: myceliumIPSeed,
			NodeID:         node,
			NetworkName:    network,
		},
	}

	cru, mru, sru, ips, err := CalcNodeResources(k.Master.Resources, k.Master.Public)
	if err != nil {
		return workloads.K8sCluster{}, err
	}

	master.CPU = uint8(cru)
	master.MemoryMB = mru * 1024
	master.DiskSizeGB = sru
	if ips == 1 {
		master.PublicIP = true
	}

	workers := []workloads.K8sNode{}
	for _, worker := range k.Workers {
		myceliumIPSeed, err := workloads.RandomMyceliumIPSeed()
		if err != nil {
			return workloads.K8sCluster{}, err
		}

		w := workloads.K8sNode{
			VM: &workloads.VM{
				Name:           worker.Name,
				Flist:          k8sFlist,
				NodeID:         node,
				NetworkName:    network,
				Planetary:      true,
				MyceliumIPSeed: myceliumIPSeed,
			},
		}

		cru, mru, sru, _, err := CalcNodeResources(worker.Resources, false)
		if err != nil {
			return workloads.K8sCluster{}, err
		}

		w.CPU = uint8(cru)
		w.MemoryMB = mru * 1024
		w.DiskSizeGB = sru
		workers = append(workers, w)
	}

	k8sCluster := workloads.K8sCluster{
		Master:       &master,
		Workers:      workers,
		NetworkName:  network,
		Token:        token,
		SSHKey:       sshKey,
		SolutionType: k.Master.Name,
	}

	return k8sCluster, nil
}

func (d *Deployer) deployK8sClusterWithNetwork(ctx context.Context, k8sDeployInput models.K8sCluster, sshKey string, adminSSHKey string) (uint32, uint64, uint64, error) {
	// get available nodes
	node, err := d.getK8sAvailableNode(ctx, k8sDeployInput)
	if err != nil {
		return 0, 0, 0, err
	}

	// build network
	network, err := buildNetwork(node, fmt.Sprintf("%sk8sNet", k8sDeployInput.Master.Name))
	if err != nil {
		return 0, 0, 0, err
	}

	// build cluster
	cluster, err := buildK8sCluster(node,
		sshKey+"\n"+adminSSHKey,
		network.Name,
		k8sDeployInput,
	)
	if err != nil {
		return 0, 0, 0, err
	}

	// add network and cluster to be deployed
	err = d.Redis.PushK8s(streams.K8sDeployment{Net: &network, DL: &cluster})
	if err != nil {
		return 0, 0, 0, err
	}

	// wait for deployments
	for {
		if <-d.k8sDeployed {
			break
		}
	}

	// checks that network and k8s are deployed successfully
	loadedNet, err := d.TFPluginClient.State.LoadNetworkFromGrid(ctx, cluster.NetworkName)
	if err != nil {
		return 0, 0, 0, errors.Wrapf(err, "failed to load network '%s' on nodes %v", cluster.NetworkName, network.Nodes)
	}

	loadedCluster, err := d.TFPluginClient.State.LoadK8sFromGrid(ctx, []uint32{node}, cluster.Master.Name)
	if err != nil {
		return 0, 0, 0, errors.Wrapf(err, "failed to load kubernetes cluster '%s' on nodes %v", cluster.Master.Name, network.Nodes)
	}

	return node, loadedNet.NodeDeploymentID[node], loadedCluster.NodeDeploymentID[node], nil
}

func (d *Deployer) loadK8s(
	ctx context.Context,
	k8s models.K8sCluster,
	node uint32,
	networkContractID uint64, k8sContractID uint64,
) (models.K8sCluster, error) {
	// load cluster
	resCluster, err := d.TFPluginClient.State.LoadK8sFromGrid(ctx, []uint32{node}, k8s.Master.Name)
	if err != nil {
		return models.K8sCluster{}, err
	}

	// Updates after deployment
	k8s.Master.PublicIP = resCluster.Master.ComputedIP
	k8s.Master.YggIP = resCluster.Master.PlanetaryIP
	k8s.Master.MyceliumIP = resCluster.Master.MyceliumIP

	for i := range k8s.Workers {
		k8s.Workers[i].PublicIP = resCluster.Workers[i].ComputedIP
		k8s.Workers[i].YggIP = resCluster.Workers[i].PlanetaryIP
		k8s.Workers[i].MyceliumIP = resCluster.Workers[i].MyceliumIP
	}

	k8s.NetworkContract = int(networkContractID)
	k8s.ClusterContract = int(k8sContractID)
	k8s.State = models.StateCreated

	err = d.db.UpdateK8s(k8s)
	if err != nil {
		log.Error().Err(err).Send()
		return models.K8sCluster{}, err
	}

	return k8s, nil
}

func (d *Deployer) getK8sAvailableNode(ctx context.Context, k models.K8sCluster) (uint32, error) {
	rootfs := make([]uint64, len(k.Workers)+1)

	_, mru, sru, ips, err := CalcNodeResources(k.Master.Resources, k.Master.Public)
	if err != nil {
		return 0, err
	}

	for _, worker := range k.Workers {
		_, m, s, _, err := CalcNodeResources(worker.Resources, false)
		if err != nil {
			return 0, err
		}
		mru += m
		sru += s

		// k8s rootfs is either 2 or 0.5
		rootfs = append(rootfs, *convertGBToBytes(uint64(2)))
	}

	freeSRU := convertGBToBytes(sru)
	filter := types.NodeFilter{
		Status:  []string{statusUp},
		FreeMRU: convertGBToBytes(mru),
		FreeSRU: freeSRU,
		FreeIPs: &ips,
		FarmIDs: []uint64{1},
		IPv4:    &k.Master.Public,
	}

	if len(strings.TrimSpace(k.Master.Region)) != 0 {
		filter.Region = &k.Master.Region
	}

	nodes, err := deployer.FilterNodes(ctx, d.TFPluginClient, filter, []uint64{*freeSRU}, nil, rootfs, 1)
	if err != nil {
		return 0, err
	}

	return uint32(nodes[0].NodeID), nil
}

func (d *Deployer) deployK8sRequest(ctx context.Context, user models.User, k8sDeployInput models.K8sCluster, adminSSHKey string) (int, error, error) {
	_, err := d.CanDeployK8s(user.ID.String(), k8sDeployInput)
	if errors.Is(err, ErrCannotDeploy) {
		return http.StatusBadRequest, err, err
	}
	if err != nil {
		return http.StatusInternalServerError, err, errors.New(internalServerErrorMsg)
	}

	// deploy network and cluster
	node, networkContractID, k8sContractID, err := d.deployK8sClusterWithNetwork(ctx, k8sDeployInput, user.SSHKey, adminSSHKey)
	if err != nil {
		return http.StatusInternalServerError, err, errors.New(internalServerErrorMsg)
	}

	k8sCluster, err := d.loadK8s(ctx, k8sDeployInput, node, networkContractID, k8sContractID)
	if err != nil {
		return http.StatusInternalServerError, err, errors.New(internalServerErrorMsg)
	}

	err = d.db.UpdateK8s(k8sCluster)
	if err != nil {
		return http.StatusInternalServerError, err, errors.New(internalServerErrorMsg)
	}

	// metrics
	middlewares.Deployments.WithLabelValues(user.ID.String(), k8sDeployInput.Master.Resources, "master").Inc()
	for _, worker := range k8sDeployInput.Workers {
		middlewares.Deployments.WithLabelValues(user.ID.String(), worker.Resources, "worker").Inc()
	}

	return 0, nil, nil
}

// CanDeployK8s checks if user can deploy kubernetes
func (d *Deployer) CanDeployK8s(userID string, k8s models.K8sCluster) (float64, error) {
	k8sPrice, err := calcPrice(d.prices, k8s.Master.Resources, k8s.Master.Public)
	if err != nil {
		return 0, errors.Wrap(err, "failed to calculate kubernetes master price")
	}

	for _, worker := range k8s.Workers {
		workerPrice, err := calcPrice(d.prices, worker.Resources, worker.Public)
		if err != nil {
			return 0, errors.Wrap(err, "failed to calculate kubernetes worker price")
		}

		k8sPrice += workerPrice
	}

	return k8sPrice, d.canDeploy(userID, k8sPrice)
}
