// Package models for database models
package models

// VM struct for vms data
type VM struct {
	ID     string `json:"id"`
	UserID string `json:"userID"`
	Name   string `json:"name"`
	IP     string `json:"ip"`
}
