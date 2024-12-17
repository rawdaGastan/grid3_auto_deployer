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
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

func (d *Deployer) deployVM(ctx context.Context, vmInput models.VM, sshKey string, adminSSHKey string) (*workloads.VM, uint64, uint64, error) {
	// filter nodes
	cru, mru, sru, ips, err := CalcNodeResources(vmInput.Resources, vmInput.Public)
	if err != nil {
		return nil, 0, 0, err
	}

	freeSRU := convertGBToBytes(sru)
	filter := types.NodeFilter{
		FarmIDs:  []uint64{1},
		TotalCRU: &cru,
		FreeSRU:  freeSRU,
		FreeMRU:  convertGBToBytes(mru),
		FreeIPs:  &ips,
		Status:   []string{statusUp},
		IPv4:     &vmInput.Public,
	}

	if len(strings.TrimSpace(vmInput.Region)) != 0 {
		filter.Region = &vmInput.Region
	}

	nodeIDs, err := deployer.FilterNodes(ctx, d.tfPluginClient, filter, []uint64{*freeSRU}, nil, nil, 1)
	if err != nil {
		return nil, 0, 0, err
	}
	nodeID := uint32(nodeIDs[0].NodeID)

	// create network workload
	network, err := buildNetwork(nodeID, fmt.Sprintf("%svmNet", vmInput.Name))
	if err != nil {
		return nil, 0, 0, err
	}

	// create disk
	disk := workloads.Disk{
		Name:   "disk",
		SizeGB: sru,
	}

	myceliumIPSeed, err := workloads.RandomMyceliumIPSeed()
	if err != nil {
		return nil, 0, 0, err
	}

	// create vm workload
	vm := workloads.VM{
		Name:           vmInput.Name,
		Flist:          vmFlist,
		CPU:            uint8(*filter.TotalCRU),
		PublicIP:       vmInput.Public,
		Planetary:      true,
		MyceliumIPSeed: myceliumIPSeed,
		MemoryMB:       mru * 1024,
		Mounts: []workloads.Mount{
			{Name: disk.Name, MountPoint: "/disk"},
		},
		Entrypoint: vmEntryPoint,
		EnvVars: map[string]string{
			"SSH_KEY": sshKey + "\n" + adminSSHKey,
		},
		NetworkName: network.Name,
		NodeID:      nodeID,
	}

	dl := workloads.NewDeployment(vmInput.Name, nodeID, vmInput.Name, nil, network.Name, []workloads.Disk{disk}, nil, []workloads.VM{vm}, nil, nil, nil)

	// add network and deployment to be deployed
	err = d.Redis.PushVM(streams.VMDeployment{Net: &network, DL: &dl})
	if err != nil {
		return nil, 0, 0, err
	}

	// wait for deployments
	for {
		if <-d.vmDeployed {
			break
		}
	}

	// checks that network and vm are deployed successfully
	loadedNet, err := d.tfPluginClient.State.LoadNetworkFromGrid(ctx, dl.NetworkName)
	if err != nil {
		return nil, 0, 0, errors.Wrapf(err, "failed to load network '%s' on node %v", dl.NetworkName, dl.NodeID)
	}

	loadedDl, err := d.tfPluginClient.State.LoadDeploymentFromGrid(ctx, nodeID, dl.Name)
	if err != nil {
		return nil, 0, 0, errors.Wrapf(err, "failed to load vm '%s' on node %v", dl.Name, dl.NodeID)
	}

	return &loadedDl.Vms[0], loadedDl.ContractID, loadedNet.NodeDeploymentID[nodeID], nil
}

func (d *Deployer) deployVMRequest(ctx context.Context, user models.User, vm models.VM, adminSSHKey string) (int, error, error) {
	_, err := d.CanDeployVM(user.ID.String(), vm)
	if errors.Is(err, ErrCannotDeploy) {
		return http.StatusBadRequest, err, err
	}
	if err != nil {
		return http.StatusInternalServerError, err, errors.New(internalServerErrorMsg)
	}

	deployedVM, contractID, networkContractID, err := d.deployVM(ctx, vm, user.SSHKey, adminSSHKey)
	if err != nil {
		return http.StatusInternalServerError, err, errors.New(internalServerErrorMsg)
	}

	// Updates after deployment
	vm.YggIP = deployedVM.PlanetaryIP
	vm.MyceliumIP = deployedVM.MyceliumIP
	vm.PublicIP = deployedVM.ComputedIP
	vm.ContractID = contractID
	vm.NetworkContractID = networkContractID
	vm.State = models.StateCreated

	err = d.db.UpdateVM(vm)
	if err != nil {
		return http.StatusInternalServerError, err, errors.New(internalServerErrorMsg)
	}

	middlewares.Deployments.WithLabelValues(user.ID.String(), vm.Resources, "vm").Inc()
	return 0, nil, nil
}

// CanDeployVM checks if user can deploy a vm according to its price
func (d *Deployer) CanDeployVM(userID string, vm models.VM) (float64, error) {
	vmPrice, err := calcPrice(d.prices, vm.Resources, vm.Public)
	if err != nil {
		return 0, err
	}

	return vmPrice, d.canDeploy(userID, vmPrice)
}
