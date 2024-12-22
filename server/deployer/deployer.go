// Package deployer for handling deployments
package deployer

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/codescalers/cloud4students/validators"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

const internalServerErrorMsg = "Something went wrong"

var (
	ErrCannotDeploy = errors.New("cannot proceed with deployment, either add a valid card or apply for a new voucher")

	vmEntryPoint = "/init.sh"

	k8sFlist = "https://hub.grid.tf/tf-official-apps/threefoldtech-k3s-latest.flist"
	vmFlist  = "https://hub.grid.tf/tf-official-vms/ubuntu-22.04.flist"

	smallCPU     = uint64(1)
	smallMemory  = uint64(2)
	smallDisk    = uint64(25)
	mediumCPU    = uint64(2)
	mediumMemory = uint64(4)
	mediumDisk   = uint64(50)
	largeCPU     = uint64(4)
	largeMemory  = uint64(8)
	largeDisk    = uint64(100)

	statusUp = "up"

	token = "random"
)

// Deployer struct holds deployments configuration
type Deployer struct {
	db             models.DB
	Redis          streams.RedisClient
	TFPluginClient deployer.TFPluginClient
	prices         internal.Prices

	vmDeployed  chan bool
	k8sDeployed chan bool
}

// NewDeployer create new deployer
func NewDeployer(
	db models.DB, redis streams.RedisClient, tfPluginClient deployer.TFPluginClient, prices internal.Prices,
) (Deployer, error) {
	// validations
	err := validator.SetValidationFunc("ssh", validators.ValidateSSHKey)
	if err != nil {
		return Deployer{}, err
	}
	err = validator.SetValidationFunc("password", validators.ValidatePassword)
	if err != nil {
		return Deployer{}, err
	}
	err = validator.SetValidationFunc("mail", validators.ValidateMail)
	if err != nil {
		return Deployer{}, err
	}

	return Deployer{
		db,
		redis,
		tfPluginClient,
		prices,
		make(chan bool),
		make(chan bool),
	}, nil
}

// PeriodicRequests for executing deployment api requests
func (d *Deployer) PeriodicRequests(ctx context.Context, sec int) {
	ticker := time.NewTicker(time.Second * time.Duration(sec))
	for range ticker.C {
		d.ConsumeVMRequest(ctx, false)
		d.ConsumeK8sRequest(ctx, false)
	}
}

// PeriodicDeploy for executing deployments
func (d *Deployer) PeriodicDeploy(ctx context.Context, sec int) {
	ticker := time.NewTicker(time.Second * time.Duration(sec))

	for range ticker.C {
		vmNets, vms, err := d.consumeVMs()
		if err != nil {
			log.Error().Err(err).Msg("failed to consume vms")
		}

		k8sNets, clusters, err := d.consumeK8s()
		if err != nil {
			log.Error().Err(err).Msg("failed to consume clusters")
		}

		if len(vms) > 0 {
			err := d.TFPluginClient.NetworkDeployer.BatchDeploy(ctx, vmNets)
			if err != nil {
				log.Error().Err(err).Msg("failed to batch deploy network")
			}

			err = d.TFPluginClient.DeploymentDeployer.BatchDeploy(ctx, vms)
			if err != nil {
				log.Error().Err(err).Msg("failed to batch deploy vm")
			}

			for i := 0; i < len(vms); i++ {
				d.vmDeployed <- true
			}
		}

		if len(clusters) > 0 {
			err := d.TFPluginClient.NetworkDeployer.BatchDeploy(ctx, k8sNets)
			if err != nil {
				log.Error().Err(err).Msg("failed to batch deploy network")
			}

			err = d.TFPluginClient.K8sDeployer.BatchDeploy(ctx, clusters)
			if err != nil {
				log.Error().Err(err).Msg("failed to batch deploy clusters")
			}

			for i := 0; i < len(clusters); i++ {
				d.k8sDeployed <- true
			}
		}
	}
}

// CancelDeployment cancel deployments from grid
func (d *Deployer) CancelDeployment(contractID uint64, netContractID uint64, dlType string, dlName string) error {
	// cancel deployment
	err := d.TFPluginClient.SubstrateConn.CancelContract(d.TFPluginClient.Identity, contractID)
	if err != nil {
		return err
	}

	// cancel network
	err = d.TFPluginClient.SubstrateConn.CancelContract(d.TFPluginClient.Identity, netContractID)
	if err != nil {
		return err
	}

	// update state
	for node, contracts := range d.TFPluginClient.State.CurrentNodeDeployments {
		contracts = workloads.Delete(contracts, contractID)
		contracts = workloads.Delete(contracts, netContractID)
		d.TFPluginClient.State.CurrentNodeDeployments[node] = contracts

		d.TFPluginClient.State.Networks.DeleteNetwork(fmt.Sprintf("%s%sNet", dlType, dlName))
	}

	return nil
}

func buildNetwork(node uint32, name string) (workloads.ZNet, error) {
	myceliumKey, err := workloads.RandomMyceliumKey()
	if err != nil {
		return workloads.ZNet{}, err
	}

	return workloads.ZNet{
		Name:  name,
		Nodes: []uint32{node},
		IPRange: workloads.NewIPRange(net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}),
		AddWGAccess:  false,
		MyceliumKeys: map[uint32][]byte{node: myceliumKey},
	}, nil
}

func CalcNodeResources(resources string, public bool) (uint64, uint64, uint64, uint64, error) {
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

func calcPrice(prices internal.Prices, resources string, public bool) (float64, error) {
	var price float64
	switch resources {
	case "small":
		price += prices.SmallVM
	case "medium":
		price += prices.MediumVM
	case "large":
		price += prices.LargeVM
	default:
		return 0, fmt.Errorf("unknown resource type %s", resources)
	}

	if public {
		price += prices.PublicIP
	}
	return price, nil
}

func convertGBToBytes(gb uint64) *uint64 {
	bytes := gb * 1024 * 1024 * 1024
	return &bytes
}

// canDeploy checks if user has a valid card so can deploy or has enough voucher money
func (d *Deployer) canDeploy(userID string, costPerMonth float64) error {
	// check if user has a valid card
	_, err := d.db.GetUserCards(userID)
	if err == gorm.ErrRecordNotFound {
		// If no? check if user has enough voucher balance respecting his active deployments (debt)
		user, err := d.db.GetUserByID(userID)
		if err != nil {
			return err
		}

		// calculate new debt during the current month (for active deployments)
		newDebt, err := d.calculateUserDebtInMonth(userID)
		if err != nil {
			return err
		}

		userDebt, err := d.db.CalcUserDebt(userID)
		if err != nil {
			return err
		}

		debt := userDebt + newDebt
		// if user has enough money for new cost and his debt then can deploy
		if user.VoucherBalance > debt+costPerMonth {
			return nil
		}

		return ErrCannotDeploy
	}

	return err
}

// calculateUserDebtInMonth calculates how much money does user have used
// from the start of current month
func (d *Deployer) calculateUserDebtInMonth(userID string) (float64, error) {
	var debt float64
	usagePercentageInMonth := UsagePercentageInMonth(time.Now())

	vms, err := d.db.GetAllVms(userID)
	if err != nil {
		return 0, err
	}

	for _, vm := range vms {
		debt += float64(vm.PricePerMonth) * usagePercentageInMonth
	}

	clusters, err := d.db.GetAllK8s(userID)
	if err != nil {
		return 0, err
	}

	for _, c := range clusters {
		debt += float64(c.PricePerMonth) * usagePercentageInMonth
	}

	return debt, nil
}

// UsagePercentageInMonth calculates percentage of hours till specific time during the month
// according to total hours of the same month
func UsagePercentageInMonth(end time.Time) float64 {
	start := time.Date(end.Year(), end.Month(), 0, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(end.Year(), end.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	return end.Sub(start).Hours() / endMonth.Sub(start).Hours()
}
