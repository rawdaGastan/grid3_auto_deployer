package models

import "github.com/threefoldtech/grid3-go/workloads"

type VM struct {
	UserID uint64       `json:"userID"`
	VM     workloads.VM `json:"vm"`
}
