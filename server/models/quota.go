// Package models for database models
package models

// Quota struct holds available vms for each user
type Quota struct {
	UserID    string `json:"user_id"`
	Vms       int    `json:"vms"`
	PublicIPs int    `json:"public_ips"`
}
