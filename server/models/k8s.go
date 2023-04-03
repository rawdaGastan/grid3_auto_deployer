// Package models for database models
package models

// K8sCluster holds all cluster data
type K8sCluster struct {
	ID              int      `json:"id" gorm:"primaryKey"`
	UserID          string   `json:"userID"`
	NetworkContract int      `json:"network_contract_id"`
	ClusterContract int      `json:"contract_id"`
	Master          Master   `json:"master" gorm:"foreignKey:ClusterID"`
	Workers         []Worker `json:"workers" gorm:"foreignKey:ClusterID"`
}

// Master struct for kubernetes master data
type Master struct {
	ClusterID int    `json:"clusterID"`
	Name      string `json:"name"`
	CRU       uint64 `json:"cru"`
	MRU       uint64 `json:"mru"`
	SRU       uint64 `json:"sru"`
	YggIP     string `json:"ygg_ip"`
	Public    bool   `json:"public"`
	PublicIP  string `json:"public_ip"`
	Resources string `json:"resources"`
}

// Worker struct for k8s workers data
type Worker struct {
	ClusterID int    `json:"clusterID"`
	Name      string `json:"name"`
	CRU       uint64 `json:"cru"`
	MRU       uint64 `json:"mru"`
	SRU       uint64 `json:"sru"`
	Resources string `json:"resources"`
}
