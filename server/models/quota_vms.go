// Package models for database models
package models

// QuotaVM struct holds available vms and their expiration date for each user
type QuotaVM struct {
	QuotaID  string `json:"qouta_id"`
	VMs      int    `json:"vms"`
	Duration int    `json:"duration" gorm:"unique"`
}
