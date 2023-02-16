package main

import (
	"time"

	"github.com/threefoldtech/grid3-go/workloads"
)

type User struct {
	ID        uint64    `json:"id" gorm:"unique;primaryKey;type:uuid;not null"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" gorm:"unique" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	Voucher   string    `json:"voucher"`
	Verified  bool      `json:"verified" gorm:"not null"`
	CreatedAt time.Time `json:"createdat" gorm:"not null"`
	UpdatedAt time.Time `json:"updatedat" gorm:"not null"`
}

type Quota struct {
	UserID string `json:"userID"`
	Vms    int    `json:"vms"`
	K8s    int    `json:"k8s"`
}

type VM struct {
	UserID string       `json:"userID"`
	VM     workloads.VM `json:"vm"`
}

type Kubernetes struct {
	UserID string               `json:"userID"`
	K8s    workloads.K8sCluster `json:"k8s"`
}
