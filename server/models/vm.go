// Package models for database models
package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// VM struct for vms data
type VM struct {
	ID                int       `json:"id" gorm:"primaryKey"`
	UserID            string    `json:"user_id"`
	Name              string    `json:"name" gorm:"unique" binding:"required"`
	YggIP             string    `json:"ygg_ip"`
	MyceliumIP        string    `json:"mycelium_ip"`
	Public            bool      `json:"public"`
	PublicIP          string    `json:"public_ip"`
	Resources         string    `json:"resources"`
	Region            string    `json:"region"`
	SRU               uint64    `json:"sru"`
	CRU               uint64    `json:"cru"`
	MRU               uint64    `json:"mru"`
	ContractID        uint64    `json:"contractID"`
	NetworkContractID uint64    `json:"networkContractID"`
	State             state     `json:"state"`
	Failure           string    `json:"failure"`
	PricePerMonth     float64   `json:"price"`
	CreatedAt         time.Time `json:"created_at"`
}

// CreateVM creates new vm
func (d *DB) CreateVM(vm *VM) error {
	return d.db.Create(&vm).Error
}

// GetVMByID return vm by its id
func (d *DB) GetVMByID(id int) (VM, error) {
	var vm VM
	return vm, d.db.First(&vm, id).Error
}

// GetAllVms returns all vms of user
func (d *DB) GetAllVms(userID string) ([]VM, error) {
	var vms []VM
	return vms, d.db.Where("user_id = ?", userID).Find(&vms).Error
}

// GetAllSuccessfulVms returns all vms of user that have a state succeeded
func (d *DB) GetAllSuccessfulVms(userID string) ([]VM, error) {
	var vms []VM
	return vms, d.db.Where("user_id = ? and state = 'CREATED'", userID).Find(&vms).Error
}

// UpdateVM updates information of vm. empty and unchanged fields are not updated.
func (d *DB) UpdateVM(vm VM) error {
	return d.db.Model(&VM{}).Where("id = ?", vm.ID).Updates(vm).Error
}

// UpdateVMState updates state of vm
func (d *DB) UpdateVMState(id int, failure string, state state) error {
	var vm VM
	result := d.db.Model(&vm).Where("id = ?", id).Update("state", state)
	if state == StateFailed {
		result.Update("failure", failure)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// AvailableVMName returns if name available
func (d *DB) AvailableVMName(name string) (bool, error) {
	var names []string
	query := d.db.Table("vms").
		Select("name").
		Where("name = ?", name).
		Scan(&names)

	if query.Error != nil {
		return false, query.Error
	}
	return len(names) == 0, query.Error
}

// DeleteVMByID deletes vm by its id
func (d *DB) DeleteVMByID(id int) error {
	var vm VM
	return d.db.Delete(&vm, id).Error
}

// DeleteAllVms deletes all vms of user
func (d *DB) DeleteAllVms(userID string) error {
	var vms []VM
	return d.db.Clauses(clause.Returning{}).Where("user_id = ?", userID).Delete(&vms).Error
}
