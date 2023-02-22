package models

//TODO: Change name to be same as repo

import (
	"github.com/google/uuid"
	"github.com/threefoldtech/grid3-go/workloads"
)

type Kubernetes struct {
	UserID uuid.UUID            `json:"userID"`
	K8s    workloads.K8sCluster `json:"k8s"`
}
