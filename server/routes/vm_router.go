// Package routes for API endpoints
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/threefoldtech/grid3-go/deployer"
	"github.com/threefoldtech/grid3-go/workloads"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

//TODO: add bin folder

// DeployVmInput struct takes input of vm from user
type DeployVmInput struct {
	Name      string `json:"name" binding:"required"`
	Resources string `json:"resources" binding:"required"`
	SSHKey    string `json:"ssh_key" binding:"required"`
}

var (
	// Flist for the vm
	Flist        = "https://hub.grid.tf/tf-official-apps/base:latest.flist"
	trueVal      = true
	statusUp     = "up"
	smallCPU     = uint64(1)
	smallMemory  = uint64(2)
	smallDisk    = uint64(5)
	mediumCPU    = uint64(2)
	mediumMemory = uint64(4)
	mediumDisk   = uint64(10)
	largeCPU     = uint64(4)
	largeMemory  = uint64(8)
	largeDisk    = uint64(15)
)

// DeployVMHandler creates vm for user and deploy it
func (r *Router) DeployVMHandler(w http.ResponseWriter, req *http.Request) {
	setupCorsResponse(&w, req)
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
	var InputVM DeployVmInput
	err = json.NewDecoder(req.Body).Decode(&InputVM)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// TODO: move to function validate quota (shared)
	// check quota of user
	quota, err := r.db.GetUserQuota(claims.UserID)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	availableVms := 0
	switch InputVM.Resources {
	case "small":
		availableVms = 1
	case "medium":
		availableVms = 2
	case "large":
		availableVms = 3
	}
	if quota.Vms < availableVms {
		r.WriteErrResponse(w, fmt.Errorf("no available vms"))
		return
	}

	vm, contractID, networkContractID, diskSize, err := r.deployVM(InputVM)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	userVM := models.VM{
		UserID:            claims.UserID,
		Name:              vm.Name,
		IP:                vm.YggIP,
		Resources:         InputVM.Resources,
		SRU:               diskSize,
		CRU:               uint64(vm.CPU),
		MRU:               uint64(vm.Memory),
		ContractID:        contractID,
		NetworkContractID: networkContractID,
	}

	err = r.db.CreateVM(&userVM)
	if err != nil {
		r.WriteErrResponse(w, err)
	}

	// update quota of user
	err = r.db.UpdateUserQuota(claims.UserID, quota.Vms-availableVms, quota.K8s)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	r.WriteMsgResponse(w, "vm deployed successfully", map[string]int{"ID": userVM.ID})
}

// choose suitable nodes based on needed resources
func filterNode(resource string) types.NodeFilter {
	var filter types.NodeFilter
	switch resource {
	case "small":
		filter = types.NodeFilter{
			TotalCRU: &smallCPU,
			TotalSRU: &smallDisk,
			TotalMRU: &smallMemory,
			Status:   &statusUp,
			IPv6:     &trueVal,
		}
	case "medium":
		filter = types.NodeFilter{
			TotalCRU: &mediumCPU,
			TotalSRU: &mediumDisk,
			TotalMRU: &mediumMemory,
			Status:   &statusUp,
			IPv6:     &trueVal,
		}
	case "large":
		filter = types.NodeFilter{
			TotalCRU: &largeCPU,
			TotalSRU: &largeDisk,
			TotalMRU: &largeMemory,
			Status:   &statusUp,
			IPv6:     &trueVal,
		}

	}

	return filter

}

func (r *Router) deployVM(VM DeployVmInput) (*workloads.VM, uint64, uint64, uint64, error) {
	// create tfPluginClient
	tfPluginClient, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, false)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// filter nodes
	filter := filterNode(VM.Resources)
	nodeIDs, err := deployer.FilterNodes(tfPluginClient.GridProxyClient, filter)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	nodeID := uint32(nodeIDs[0].NodeID)

	// generate network name
	name := generateNetworkName()

	// create network workload
	network := workloads.ZNet{
		Name:        name,
		Description: "A network to deploy",
		Nodes:       []uint32{nodeID},
		IPRange: gridtypes.NewIPNet(net.IPNet{
			IP:   net.IPv4(10, 1, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}),
		AddWGAccess: true,
	}

	// create disk
	disk := workloads.Disk{
		Name:   "disk",
		SizeGB: int(*filter.TotalSRU),
	}

	// create vm workload
	vm := workloads.VM{
		Name:      VM.Name,
		Flist:     Flist,
		CPU:       int(*filter.TotalCRU),
		PublicIP:  false,
		Planetary: true,
		Memory:    int(*filter.TotalMRU) * 1024,
		Mounts: []workloads.Mount{
			{DiskName: disk.Name, MountPoint: "/disk"},
		},
		Entrypoint: "/sbin/zinit init",
		EnvVars: map[string]string{
			"SSH_KEY": VM.SSHKey,
		},
		NetworkName: network.Name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// deploy network
	err = tfPluginClient.NetworkDeployer.Deploy(ctx, &network)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// deploy vm
	dl := workloads.NewDeployment("vm", nodeID, "", nil, network.Name, []workloads.Disk{disk}, nil, []workloads.VM{vm}, nil)
	err = tfPluginClient.DeploymentDeployer.Deploy(ctx, &dl)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// checks that vm deployed successfully
	loadedVM, err := tfPluginClient.State.LoadVMFromGrid(nodeID, vm.Name, dl.Name)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	fmt.Printf("loadedVM: %v\n", loadedVM)

	return &loadedVM, dl.ContractID, network.NodeDeploymentID[nodeID], uint64(disk.SizeGB), nil
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

// GetVMHandler returns vm by its id
func (r *Router) GetVMHandler(w http.ResponseWriter, req *http.Request) {
	setupCorsResponse(&w, req)
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	_, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	id := mux.Vars(req)["id"]
	vm, err := r.db.GetVMByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "vm", vm)
}

// ListVMsHandler returns all vms of user
func (r *Router) ListVMsHandler(w http.ResponseWriter, req *http.Request) {
	setupCorsResponse(&w, req)
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

	vms, err := r.db.GetAllVms(claims.UserID)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "", vms)

}

// DeleteVM deletes vm by its id
func (r *Router) DeleteVM(w http.ResponseWriter, req *http.Request) {
	setupCorsResponse(&w, req)
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	_, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	id := mux.Vars(req)["id"]

	vm, err := r.db.GetVMByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	err = r.cancelDeployment(vm)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	err = r.db.DeleteVMByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "vm deleted successfully", "")
}

func (r *Router) cancelDeployment(vm *models.VM) error {
	tfPluginClient, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, false)
	if err != nil {
		return err
	}

	// cancel vm
	err = tfPluginClient.SubstrateConn.CancelContract(tfPluginClient.Identity, vm.ContractID)
	if err != nil {
		return err
	}

	// cancel network
	err = tfPluginClient.SubstrateConn.CancelContract(tfPluginClient.Identity, vm.NetworkContractID)
	if err != nil {
		return err
	}

	return nil

}

// DeleteAllVMs deletes all vms of user
func (r *Router) DeleteAllVMs(w http.ResponseWriter, req *http.Request) {
	setupCorsResponse(&w, req)
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

	vms, err := r.db.GetAllVms(claims.UserID)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	for _, vm := range vms {
		err = r.cancelDeployment(&vm)
		if err != nil {
			r.WriteErrResponse(w, err)
			return
		}
	}

	err = r.db.DeleteAllVms(claims.UserID)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "all vms deleted successfully", "")

}
