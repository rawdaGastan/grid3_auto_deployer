// Package models for database models
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Quota struct holds available vms for each user
type Quota struct {
	ID        uuid.UUID `gorm:"primary_key; unique; type:uuid; column:id"`
	UserID    string    `json:"user_id"`
	PublicIPs int       `json:"public_ips"`
	VMs       []QuotaVM `json:"vms"`
}

// BeforeCreate generates a new uuid
func (quota *Quota) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	quota.ID = id
	return
}
