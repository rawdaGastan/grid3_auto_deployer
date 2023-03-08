// Package routes for API endpoints
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/threefoldtech/grid3-go/deployer"
	integration "github.com/threefoldtech/grid3-go/integration_tests"
	"github.com/threefoldtech/grid3-go/workloads"
	"github.com/threefoldtech/grid_proxy_server/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

//TODO: add bin folder

// DeployVmInput struct takes input of vm from user
type DeployVmInput struct {
	Name      string `json:"name" binding:"required"`
	Resources string `json:"resources" binding:"required"`
	Image     string `json:"image" binding:"required"`
}

var (
	// Flist for the vm
	Flist        = "https://hub.grid.tf/tf-official-apps/base:latest.flist"
	trueVal      = true
	statusUp     = "up"
	smallCPU     = uint64(1)
	smallMemory  = uint64(2)
	smallDisk    = uint64(10)
	mediumCPU    = uint64(2)
	mediumMemory = uint64(4)
	mediumDisk   = uint64(15)
	largeCPU     = uint64(1)
	largeMemory  = uint64(2)
	largeDisk    = uint64(10)
)

// DeployVMHandler creates vm for user and deploy it
func (r *Router) DeployVMHandler(w http.ResponseWriter, req *http.Request) {
	//TODO: remove id of user , get it from token
	//TODO: validation that user has available vms
	setupCorsResponse(&w, req)
	id := mux.Vars(req)["id"]
	var VM DeployVmInput
	err := json.NewDecoder(req.Body).Decode(&VM)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}

	vm, err := r.deployVM(VM)
	if err != nil {
		r.WriteErrResponse(w, err)
	}

	userVM := models.VM{ //TODO: id ??
		UserID: user.ID.String(),
		Name:   vm.Name,
		IP:     vm.YggIP,
	}

	err = r.db.CreateVM(&userVM)
	if err != nil {
		r.WriteErrResponse(w, err)
	}

}

// choose suitable nodes based on needed resources
func (r *Router) filterNode(resource string) types.NodeFilter {
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

func (r *Router) deployVM(VM DeployVmInput) (*workloads.VM, error) {
	// create tfPluginClient
	tfPluginClient, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, true)
	if err != nil {
		return nil, err
	}

	// filter nodes
	filter := r.filterNode(VM.Resources)
	nodeIDs, err := deployer.FilterNodes(tfPluginClient.GridProxyClient, filter)
	if err != nil {
		return nil, err
	}
	nodeID := uint32(nodeIDs[0].NodeID)

	// generate network name
	name := r.generateNetworkName()

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
	publicKey, _, err := integration.GenerateSSHKeyPair()
	if err != nil {
		return nil, err
	}

	// create vm workload
	vm := workloads.VM{
		Name:       VM.Name,
		Flist:      Flist,
		CPU:        2,
		PublicIP:   false,
		Planetary:  true,
		Memory:     1024,
		Entrypoint: "/sbin/zinit init",
		EnvVars: map[string]string{
			"SSH_KEY": publicKey,
		},
		IP:          "10.20.2.5",
		NetworkName: network.Name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// deploy network
	err = tfPluginClient.NetworkDeployer.Deploy(ctx, &network)
	if err != nil {
		return nil, err
	}

	// deploy vm
	dl := workloads.NewDeployment("vm", nodeID, "", nil, network.Name, nil, nil, []workloads.VM{vm}, nil)
	err = tfPluginClient.DeploymentDeployer.Deploy(ctx, &dl)
	if err != nil {
		return nil, err
	}

	// checks that vm deployed successfully
	loadedVM, err := tfPluginClient.State.LoadVMFromGrid(nodeID, vm.Name, dl.Name)
	if err != nil {
		return nil, err
	}

	fmt.Printf("loadedVM: %v\n", loadedVM)

	return &vm, nil
}

// generate random names for network
func (r *Router) generateNetworkName() string {
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
	// TODO: no id needed
	id := mux.Vars(req)["id"]
	vms, err := r.db.GetAllVms(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "", vms)

}

// DeleteVM deletes vm by its id
func (r *Router) DeleteVM(w http.ResponseWriter, req *http.Request) {
	setupCorsResponse(&w, req)
	id := mux.Vars(req)["id"]
	err := r.db.DeleteVMByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "vm deleted successfully", "")
}

// DeleteAllVMs deletes all vms of user
func (r *Router) DeleteAllVMs(w http.ResponseWriter, req *http.Request) {
	setupCorsResponse(&w, req)
	// TODO: no id needed
	id := mux.Vars(req)["id"]
	err := r.db.DeleteAllVms(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "all vms deleted successfully", "")

}
