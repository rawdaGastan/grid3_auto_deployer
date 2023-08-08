// Package models for database models
package models

// DeployVMInput struct takes input of vm from user
type DeployVMInput struct {
	Name      string `json:"name" binding:"required" validate:"min=3,max=20"`
	Resources VMType `json:"resources" binding:"required"`
	Public    bool   `json:"public"`
	PkgID     int    `json:"pkg_id"`
}

// K8sDeployInput deploy k8s cluster input
type K8sDeployInput struct {
	MasterName string   `json:"master_name" validate:"min=3,max=20"`
	Resources  VMType   `json:"resources"`
	Public     bool     `json:"public"`
	Workers    []Worker `json:"workers"`
	PkgID      int      `json:"pkg_id"`
}

// WorkerInput deploy k8s worker input
type WorkerInput struct {
	Name      string `json:"name" validate:"min=3,max=20"`
	Resources VMType `json:"resources"`
}
