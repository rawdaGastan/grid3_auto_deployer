// Package models for database models
package models

import "time"

// Maintenance struct for maintenance.
type Maintenance struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Active    bool      `json:"active"`
	UpdatedAt time.Time `json:"updated_at"`
}
