// Package models for database models
package models

const (
	// VMsRedirection deployment
	VMsType = "vms"
	K8sType = "k8s"
)

// Notification struct holds data of notifications
type Notification struct {
	ID     int    `json:"id" gorm:"primaryKey"`
	UserID string `json:"user_id"  binding:"required"`
	Msg    string `json:"msg" binding:"required"`
	Seen   bool   `json:"seen" binding:"required"`
	// to allow redirecting from notifications to the right pages
	Type string `json:"type" binding:"required"`
}
