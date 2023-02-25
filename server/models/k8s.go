package models

//TODO: Change name to be same as repo

import (
	"github.com/threefoldtech/grid3-go/workloads"
)

type Kubernetes struct {
	UserID string               `json:"userID"`
	K8s    workloads.K8sCluster `json:"k8s"`
}
