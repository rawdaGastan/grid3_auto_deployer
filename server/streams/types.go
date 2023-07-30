// Package streams for redis streams
package streams

import (
	"github.com/codescalers/cloud4students/models"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

const (
	// DeployVMConsumerGroupName consumer group name
	DeployVMConsumerGroupName = "vms-group"
	// DeployK8sConsumerGroupName consumer group name
	DeployK8sConsumerGroupName = "k8s-group"

	// ReqVMConsumerGroupName consumer group name
	ReqVMConsumerGroupName = "vms-req-group"
	// ReqK8sConsumerGroupName consumer group name
	ReqK8sConsumerGroupName = "k8s-req-group"

	// DeployVMStreamName stream name
	DeployVMStreamName = "vms"
	// DeployK8sStreamName stream name
	DeployK8sStreamName = "k8s"

	// ReqVMStreamName stream name
	ReqVMStreamName = "vms-req"
	// ReqK8sStreamName stream name
	ReqK8sStreamName = "k8s-req"
)

// VMDeployRequest type for redis vm deployment request
type VMDeployRequest struct {
	User                      models.User
	Input                     models.DeployVMInput
	AdminSSHKey               string
	ExpirationToleranceInDays int
}

// K8sDeployRequest type for redis k8s deployment request
type K8sDeployRequest struct {
	User                      models.User
	Input                     models.K8sDeployInput
	AdminSSHKey               string
	ExpirationToleranceInDays int
}

// VMDeployment type for redis vm deployment
type VMDeployment struct {
	Net *workloads.ZNet
	DL  *workloads.Deployment
}

// K8sDeployment type for redis k8s deployment
type K8sDeployment struct {
	Net *workloads.ZNet
	DL  *workloads.K8sCluster
}
