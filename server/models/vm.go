// Package models for database models
package models

import "time"

// VM struct for vms data
type VM struct {
	ID                int       `json:"id" gorm:"primaryKey"`
	UserID            string    `json:"user_id"`
	Name              string    `json:"name" gorm:"unique" binding:"required"`
	YggIP             string    `json:"ygg_ip"`
	Public            bool      `json:"public"`
	PublicIP          string    `json:"public_ip"`
	Resources         string    `json:"resources"`
	SRU               uint64    `json:"sru"`
	CRU               uint64    `json:"cru"`
	MRU               uint64    `json:"mru"`
	ContractID        uint64    `json:"contractID"`
	NetworkContractID uint64    `json:"networkContractID"`
	ExpirationDate    time.Time `json:"expirationDate" binding:"required"`
}

// DeploymentsCount has the vms and ips reserved in the grid
type DeploymentsCount struct {
	VMs int64 `json:"vms"`
	IPs int64 `json:"ips"`
}
