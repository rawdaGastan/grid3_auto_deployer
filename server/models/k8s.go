// Package models for database models
package models

// Master struct for kubernetes master data
type Master struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	UserID    string `json:"userID"`
	Name      string `json:"Name"`
	Resources string `json:"resources"`
	IP        string `json:"ip"`
}

// Worker struct for k8s workers data
type Worker struct {
	ClusterID int    `json:"clusterID"`
	Name      string `json:"name"`
	Resources string `json:"resources"`
}
