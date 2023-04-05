// Package models for database models
package models

import "time"

// Maintenance struct for maintenance.
type Maintenance struct {
	Active    bool      `json:"active"`
	UpdatedAt time.Time `json:"updated_at"`
}
