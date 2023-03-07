// Package models for database models
package models

import (
	"github.com/threefoldtech/grid3-go/workloads"
)

// Kubernetes struct for k8s data
type Kubernetes struct {
	ID     int                  `json:"id" gorm:"primaryKey"`
	UserID string               `json:"user_id"`
	K8s    workloads.K8sCluster `json:"k8s"`
}
