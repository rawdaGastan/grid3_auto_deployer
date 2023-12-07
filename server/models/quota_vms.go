// Package models for database models
package models

import "time"

// QuotaVM struct holds available vms and their expiration date for each user
type QuotaVM struct {
	QoutaID        string    `json:"qouta_id"`
	Vms            int       `json:"vms"`
	ExpirationDate time.Time `json:"expiration_date"`
}
