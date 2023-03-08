// Package routes for API endpoints
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/goombaio/namegenerator"
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
	//TODO: validation that user has available vms
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

	//TODO: should be separated or not ?
	tfPluginClient, err := deployer.NewTFPluginClient(r.config.Account.Mnemonics, "sr25519", "dev", "", "", "", true, true)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	name := generateNetworkName()

	filter := filterNode(VM.Resources)
	nodeIDs, err := deployer.FilterNodes(tfPluginClient.GridProxyClient, filter)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	nodeID := uint32(nodeIDs[0].NodeID)
	fmt.Printf("nodeID: %v\n", nodeID)

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
		r.WriteErrResponse(w, err)
	}
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

	err = tfPluginClient.NetworkDeployer.Deploy(ctx, &network)
	if err != nil {
		r.WriteErrResponse(w, err)
	}

	dl := workloads.NewDeployment("vm", nodeID, "", nil, network.Name, nil, nil, []workloads.VM{vm}, nil)
	err = tfPluginClient.DeploymentDeployer.Deploy(ctx, &dl)
	if err != nil {
		r.WriteErrResponse(w, err)
	}

	v, err := tfPluginClient.State.LoadVMFromGrid(nodeID, vm.Name, dl.Name)
	if err != nil {
		r.WriteErrResponse(w, err)
	}

	r.WriteMsgResponse(w, "vm", v)

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

// TODO: to be fixed
func generateNetworkName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	name := nameGenerator.Generate()
	return name
}

// GetVMHandler returns vm by its id
func (r *Router) GetVMHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	vm, err := r.db.GetVMByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "vm", vm)
}

// ListVMsHandler returns all vms of user
func (r *Router) ListVMsHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	vms, err := r.db.GetAllVms(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "", vms)

}

// DeleteVM deletes vm by its id
func (r *Router) DeleteVM(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	err := r.db.DeleteVMByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "vm deleted successfully", "")
}

// DeleteAllVMs deletes all vms of user
func (r *Router) DeleteAllVMs(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	err := r.db.DeleteAllVms(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "all vms deleted successfully", "")

}
