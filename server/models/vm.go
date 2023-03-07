// Package models for database models
package models

// VM struct for vms data
type VM struct {
	ID     string `json:"id"`
	UserID uint64 `json:"userID"`
	Name   string `json:"name"`
	IP     string `json:"ip"`
}
