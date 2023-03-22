// Package models for database models
package models

// Voucher struct holds data of vouchers
type Voucher struct {
	ID      int    `json:"id" gorm:"primaryKey"`
	UserID  string `json:"user_id"  binding:"required"`
	Voucher string `json:"voucher" gorm:"unique"`
	VMs     int    `json:"vms" binding:"required"`
	Reason  string `json:"reason" binding:"required"`
	// K8s     int    `json:"k8s" binding:"required"`
	// Used bool `json:"used" binding:"required"`
	Approved bool `json:"approved" binding:"required"`
}
