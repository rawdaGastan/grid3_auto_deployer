// Package models for database models
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Quota struct holds available vms for each user
type Quota struct {
	ID        string    `gorm:"primary_key; unique; column:id"`
	UserID    string    `json:"user_id"`
	PublicIPs int       `json:"public_ips"`
	QuotaVMs  []QuotaVM `json:"vms" gorm:"foreignKey:quota_id"`
}

// BeforeCreate generates a new uuid
func (quota *Quota) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	quota.ID = id.String()
	return
}
