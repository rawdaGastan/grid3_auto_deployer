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
	// DeployNetConsumerGroupName consumer group name
	DeployNetConsumerGroupName = "nets-group"

	// ReqVMConsumerGroupName consumer group name
	ReqVMConsumerGroupName = "vms-req-group"
	// ReqK8sConsumerGroupName consumer group name
	ReqK8sConsumerGroupName = "k8s-req-group"

	// DeployVMStreamName stream name
	DeployVMStreamName = "vms"
	// DeployK8sStreamName stream name
	DeployK8sStreamName = "k8s"
	// DeployNetStreamName stream name
	DeployNetStreamName = "nets"

	// ReqVMStreamName stream name
	ReqVMStreamName = "vms-req"
	// ReqK8sStreamName stream name
	ReqK8sStreamName = "k8s-req"
)

// ErrResponse is a response for deployment requests
type ErrResponse struct {
	Code *int
	Err  error
}

// VMDeployRequest type for redis vm deployment request
type VMDeployRequest struct {
	User        models.User
	Input       models.DeployVMInput
	AdminSSHKey string
}

// K8sDeployRequest type for redis k8s deployment request
type K8sDeployRequest struct {
	User        models.User
	Input       models.K8sDeployInput
	AdminSSHKey string
}

// VMDeployment type for redis vm deployment
type VMDeployment struct {
	DL *workloads.Deployment
}

// K8sDeployment type for redis k8s deployment
type K8sDeployment struct {
	DL *workloads.K8sCluster
}

// NetDeployment type for redis network deployment
type NetDeployment struct {
	DL *workloads.ZNet
}
