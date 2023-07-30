// Package deployer for handling deployments
package deployer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"gorm.io/gorm"
)

func (d *Deployer) deployVM(ctx context.Context, vmInput models.DeployVMInput, sshKey string, adminSSHKey string) (*workloads.VM, uint64, uint64, uint64, error) {
	// filter nodes
	filter, err := filterNode(vmInput.Resources, vmInput.Public)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	nodeIDs, err := deployer.FilterNodes(ctx, d.tfPluginClient, filter)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	nodeID := uint32(nodeIDs[0].NodeID)

	// create network workload
	network := buildNetwork(nodeID, fmt.Sprintf("%svmNet", vmInput.Name))

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
		Entrypoint: vmEntryPoint,
		EnvVars: map[string]string{
			"SSH_KEY": sshKey + "\n" + adminSSHKey,
		},
		NetworkName: network.Name,
	}

	dl := workloads.NewDeployment(vmInput.Name, nodeID, "", nil, network.Name, []workloads.Disk{disk}, nil, []workloads.VM{vm}, nil)
	dl.SolutionType = vmInput.Name

	// add network and deployment to be deployed
	err = d.Redis.PushVM(streams.VMDeployment{Net: &network, DL: &dl})
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// wait for deployments
	for {
		if <-d.vmDeployed {
			break
		}
	}

	// checks that network and vm are deployed successfully
	loadedNet, err := d.tfPluginClient.State.LoadNetworkFromGrid(dl.NetworkName)
	if err != nil {
		return nil, 0, 0, 0, errors.Wrapf(err, "failed to load network '%s' on node %v", dl.NetworkName, dl.NodeID)
	}

	loadedDl, err := d.tfPluginClient.State.LoadDeploymentFromGrid(nodeID, dl.Name)
	if err != nil {
		return nil, 0, 0, 0, errors.Wrapf(err, "failed to load vm '%s' on node %v", dl.Name, dl.NodeID)
	}

	return &loadedDl.Vms[0], loadedDl.ContractID, loadedNet.NodeDeploymentID[nodeID], uint64(disk.SizeGB), nil
}

// ValidateVMQuota validates the quota a vm deployment need
func ValidateVMQuota(vm models.DeployVMInput, availableResourcesQuota, availablePublicIPsQuota int) (int, error) {
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

func (d *Deployer) deployVMRequest(ctx context.Context, user models.User, input models.DeployVMInput, adminSSHKey string, expirationToleranceInDays int) (int, error) {
	pkg, err := d.db.GetPkgByID(input.PkgID)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	neededQuota, err := ValidateVMQuota(input, pkg.Vms, pkg.PublicIPs)
	if err != nil {
		return http.StatusBadRequest, err
	}

	vm, contractID, networkContractID, diskSize, err := d.deployVM(ctx, input, user.SSHKey, adminSSHKey)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	userVM := models.VM{
		UserID:            user.ID.String(),
		Name:              vm.Name,
		YggIP:             vm.YggIP,
		Resources:         input.Resources,
		Public:            input.Public,
		PublicIP:          vm.ComputedIP,
		SRU:               diskSize,
		CRU:               uint64(vm.CPU),
		MRU:               uint64(vm.Memory),
		ContractID:        contractID,
		NetworkContractID: networkContractID,
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().AddDate(0, pkg.PeriodInMonth, expirationToleranceInDays),
	}

	err = d.db.CreateVM(&userVM)
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	publicIPsQuota := pkg.PublicIPs
	if input.Public {
		publicIPsQuota -= publicQuota
	}
	// update package of user
	err = d.db.UpdateUserPackage(user.ID.String(), pkg.Vms-neededQuota, publicIPsQuota)
	if err == gorm.ErrRecordNotFound {
		return http.StatusNotFound, errors.New("User quota is not found")
	}
	if err != nil {
		log.Error().Err(err).Send()
		return http.StatusInternalServerError, errors.New(internalServerErrorMsg)
	}

	middlewares.Deployments.WithLabelValues(user.ID.String(), input.Resources, "vm").Inc()
	return 0, nil
}
