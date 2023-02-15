package models

import (
	"github.com/threefoldtech/grid3-go/workloads"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Voucher  string `json:"voucher"`
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