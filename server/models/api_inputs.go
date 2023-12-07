// Package models for database models
package models

// DeployVMInput struct takes input of vm from user
type DeployVMInput struct {
	Name      string `json:"name" binding:"required" validate:"min=3,max=20"`
	Resources string `json:"resources" binding:"required"`
	Duration  int    `json:"duration" binding:"required"`
	Public    bool   `json:"public"`
}

// K8sDeployInput deploy k8s cluster input
type K8sDeployInput struct {
	MasterName string   `json:"master_name" validate:"min=3,max=20"`
	Resources  string   `json:"resources"`
	Public     bool     `json:"public"`
	Workers    []Worker `json:"workers"`
	Duration   int      `json:"duration" binding:"required"`
}

// WorkerInput deploy k8s worker input
type WorkerInput struct {
	Name      string `json:"name" validate:"min=3,max=20"`
	Resources string `json:"resources"`
}
