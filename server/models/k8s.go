// package models for database models
package models

//TODO: Change name to be same as repo

import (
	"github.com/threefoldtech/grid3-go/workloads"
)

// Kubernetes struct for k8s data
type Kubernetes struct {
	ID     int                  `json:"id" gorm:"primaryKey"`
	UserID string               `json:"userID"`
	K8s    workloads.K8sCluster `json:"k8s"`
}
