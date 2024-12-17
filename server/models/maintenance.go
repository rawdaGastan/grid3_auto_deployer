// Package models for database models
package models

import "time"

// Maintenance struct for maintenance.
type Maintenance struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Active    bool      `json:"active"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateMaintenance updates if maintenance is on or off
func (d *DB) UpdateMaintenance(on bool) error {
	return d.db.Model(&Maintenance{}).Where("active = ?", !on).Updates(map[string]interface{}{"active": on, "updated_at": time.Now()}).Error
}

// GetMaintenance gets if maintenance is on or off
func (d *DB) GetMaintenance() (Maintenance, error) {
	var res Maintenance
	query := d.db.First(&res)
	return res, query.Error
}
