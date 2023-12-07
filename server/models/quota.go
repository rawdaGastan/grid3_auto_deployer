// Package models for database models
package models

import "time"

// Quota struct holds available vms for each user
type Quota struct {
	UserID    string            `json:"user_id"`
	Vms       map[time.Time]int `json:"vms"`
	PublicIPs int               `json:"public_ips"`
}
