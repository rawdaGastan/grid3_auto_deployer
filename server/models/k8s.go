// Package models for database models
package models

type K8sCluster struct {
	ID              int      `json:"id" gorm:"primaryKey"`
	UserID          string   `json:"userID"`
	NetworkContract int      `json:"networkContract"`
	ClusterContract int      `json:"clusterContract"`
	Master          Master   `json:"master" gorm:"foreignKey:ClusterID"`
	Workers         []Worker `json:"workers" gorm:"foreignKey:ClusterID"`
}

// Master struct for kubernetes master data
type Master struct {
	ClusterID int    `json:"clusterID"`
	Name      string `json:"name"`
	CRU       int    `json:"cru"`
	MRU       int    `json:"mru"`
	SRU       int    `json:"sru"`
	IP        string `json:"ip"`
}

// Worker struct for k8s workers data
type Worker struct {
	ClusterID int    `json:"clusterID"`
	Name      string `json:"name"`
	CRU       int    `json:"cru"`
	MRU       int    `json:"mru"`
	SRU       int    `json:"sru"`
}
