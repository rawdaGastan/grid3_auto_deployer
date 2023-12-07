// Package models for database models
package models

import "github.com/google/uuid"

// Quota struct holds available vms for each user
type Quota struct {
	ID        uuid.UUID `gorm:"primary_key; unique; type:uuid; column:id"`
	UserID    string    `json:"user_id"`
	PublicIPs int       `json:"public_ips"`
}
