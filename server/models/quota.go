// Package models for database models
package models

// Quota struct holds available vms && k8s for each user
type Quota struct {
	UserID string `json:"userID"`
	Vms    int    `json:"vms"`
	K8s    int    `json:"k8s"`
}

//TODO: vouchers table (id, voucher(random), vms, k8s,limit,num of used) --> admin
//TODO: Add admin 