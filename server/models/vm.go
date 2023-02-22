package models

import (
	"github.com/google/uuid"
	"github.com/threefoldtech/grid3-go/workloads"
)

type VM struct {
	UserID uuid.UUID    `json:"userID"`
	VM     workloads.VM `json:"vm"`
}
