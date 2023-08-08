// Package models for database models
package models

import "time"

// K8sCluster holds all cluster data
type K8sCluster struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	UserID          string    `json:"userID"`
	NetworkContract int       `json:"network_contract_id"`
	ClusterContract int       `json:"contract_id"`
	Master          Master    `json:"master" gorm:"foreignKey:ClusterID"`
	Workers         []Worker  `json:"workers" gorm:"foreignKey:ClusterID"`
	CreatedAt       time.Time `json:"Created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// Master struct for kubernetes master data
type Master struct {
	ClusterID int    `json:"clusterID"`
	Name      string `json:"name" gorm:"unique" binding:"required"`
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
	Resources VMType `json:"resources"`
}

// GetExpiredK8s gets expired k8s clusters
func (d *DB) GetExpiredK8s(userID string) ([]K8sCluster, error) {
	var k8sClusters []K8sCluster
	err := d.db.Find(&k8sClusters, "expires_at < ? and user_id = ?", time.Now(), userID).Error
	if err != nil {
		return nil, err
	}
	for i := range k8sClusters {
		k8sClusters[i], err = d.GetK8s(k8sClusters[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return k8sClusters, nil
}
