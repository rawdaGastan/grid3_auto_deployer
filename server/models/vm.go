// Package models for database models
package models

import (
	"github.com/threefoldtech/grid3-go/workloads"
)

// VM struct for vms data
type VM struct {
	ID     int          `json:"id" gorm:"primaryKey"`
	UserID string       `json:"user_id"`
	VM     workloads.VM `json:"vm"`
}
