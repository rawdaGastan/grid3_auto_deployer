// Package models for database models
package models

import (
	"time"

	"gorm.io/gorm"
)

// K8sCluster holds all cluster data
type K8sCluster struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	UserID          string    `json:"userID"`
	NetworkContract int       `json:"network_contract_id"`
	ClusterContract int       `json:"contract_id"`
	Master          Master    `json:"master" gorm:"foreignKey:ClusterID"`
	Workers         []Worker  `json:"workers" gorm:"foreignKey:ClusterID"`
	State           state     `json:"state"`
	Failure         string    `json:"failure"`
	PricePerMonth   float64   `json:"price"`
	CreatedAt       time.Time `json:"created_at"`
}

// Master struct for kubernetes master data
type Master struct {
	ClusterID  int    `json:"clusterID"`
	Name       string `json:"name" gorm:"unique" binding:"required"`
	CRU        uint64 `json:"cru"`
	MRU        uint64 `json:"mru"`
	SRU        uint64 `json:"sru"`
	YggIP      string `json:"ygg_ip"`
	MyceliumIP string `json:"mycelium_ip"`
	Public     bool   `json:"public"`
	PublicIP   string `json:"public_ip"`
	Resources  string `json:"resources"`
	Region     string `json:"region"`
}

// Worker struct for k8s workers data
type Worker struct {
	ClusterID  int    `json:"clusterID"`
	Name       string `json:"name" gorm:"unique" binding:"required"`
	CRU        uint64 `json:"cru"`
	MRU        uint64 `json:"mru"`
	SRU        uint64 `json:"sru"`
	YggIP      string `json:"ygg_ip"`
	MyceliumIP string `json:"mycelium_ip"`
	Public     bool   `json:"public"`
	PublicIP   string `json:"public_ip"`
	Resources  string `json:"resources"`
	Region     string `json:"region"`
}

// CreateK8s creates a new k8s cluster
func (d *DB) CreateK8s(k *K8sCluster) error {
	return d.db.Create(&k).Error
}

// UpdateK8s updates information of k8s cluster. empty and unchanged fields are not updated.
func (d *DB) UpdateK8s(k8s K8sCluster) error {
	return d.db.Model(&K8sCluster{}).Where("id = ?", k8s.ID).Updates(k8s).Error
}

// GetK8s gets a k8s cluster
func (d *DB) GetK8s(id int) (K8sCluster, error) {
	var k8s K8sCluster
	err := d.db.First(&k8s, id).Error
	if err != nil {
		return K8sCluster{}, err
	}

	var master Master
	if err = d.db.Model(&k8s).Association("Master").Find(&master); err != nil {
		return K8sCluster{}, err
	}

	var workers []Worker
	if err = d.db.Model(&k8s).Association("Workers").Find(&workers); err != nil {
		return K8sCluster{}, nil
	}

	k8s.Master = master
	k8s.Workers = workers

	return k8s, nil
}

// GetAllK8s gets all k8s clusters
func (d *DB) GetAllK8s(userID string) ([]K8sCluster, error) {
	var k8sClusters []K8sCluster
	err := d.db.Find(&k8sClusters, "user_id = ?", userID).Error
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

// GetAllSuccessfulK8s returns all K8s of user that have a state succeeded
func (d *DB) GetAllSuccessfulK8s(userID string) ([]K8sCluster, error) {
	var k8sClusters []K8sCluster
	err := d.db.Find(&k8sClusters, "user_id = ? and state = 'CREATED'", userID).Error
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

// DeleteK8s deletes a k8s cluster
func (d *DB) DeleteK8s(id int) error {
	var k8s K8sCluster
	err := d.db.First(&k8s, id).Error
	if err != nil {
		return err
	}
	return d.db.Select("Master", "Workers").Delete(&k8s).Error
}

// DeleteAllK8s deletes all k8s clusters
func (d *DB) DeleteAllK8s(userID string) error {
	var k8sClusters []K8sCluster
	err := d.db.Find(&k8sClusters, "user_id = ?", userID).Error
	if err != nil {
		return err
	}

	return d.db.Select("Master", "Workers").Delete(&k8sClusters).Error
}

// AvailableK8sName returns if name available
func (d *DB) AvailableK8sName(name string) (bool, error) {
	var names []string
	query := d.db.Table("masters").
		Select("name").
		Where("name = ?", name).
		Scan(&names)

	if query.Error != nil {
		return false, query.Error
	}
	return len(names) == 0, query.Error
}

// UpdateK8sState updates state of k8s cluster
func (d *DB) UpdateK8sState(id int, failure string, state state) error {
	var k8s K8sCluster
	result := d.db.Model(&k8s).Where("id = ?", id).Update("state", state)
	if state == StateFailed {
		result.Update("failure", failure)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
