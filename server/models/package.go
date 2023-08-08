// Package models for database models
package models

import (
	"time"
)

// VMType is the name of the VM type
type VMType string

// vm types
const (
	Small  VMType = "small"
	Medium VMType = "medium"
	Large  VMType = "large"
)

// Package struct for user packages
type Package struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	UserID        string    `json:"user_id"`
	Vms           int       `json:"vms"`
	PublicIPs     int       `json:"public_ips"`
	PeriodInMonth int       `json:"period"`
	Cost          uint64    `json:"cost"`
	RealCost      uint64    `json:"real_cost"`
	CreatedAt     time.Time `json:"Created_at"`
	VMType        VMType    `json:"vm_type"`
}

// CreatePackage creates new package
func (d *DB) CreatePackage(p *Package) error {
	return d.db.Create(&p).Error
}

// GetPackage return pkg by its id
func (d *DB) GetPackage(id int) (Package, error) {
	var pkg Package
	query := d.db.First(&pkg, id)
	return pkg, query.Error
}

// GetPkgByUserID return pkg by its user ID
func (d *DB) GetPkgByUserID(userID string) (Package, error) {
	var pkg Package
	query := d.db.First(&pkg, "user_id = ?", userID)
	return pkg, query.Error
}

// UpdatePackage updates package
func (d *DB) UpdatePackage(pkg Package) error {
	result := d.db.Model(&User{}).Where("id = ?", pkg.ID).Updates(pkg)
	return result.Error
}

// ListPackages returns all packages of user
func (d *DB) ListPackages(userID string) ([]Package, error) {
	var packages []Package
	result := d.db.Where("user_id = ?", userID).Find(&packages)
	if result.Error != nil {
		return []Package{}, result.Error
	}
	return packages, result.Error
}

// GetExpiredPackages returns expired vms
func (d *DB) GetExpiredPackages(expirationToleranceInDays int) ([]Package, error) {
	var res []Package
	query := d.db.Table("packages").
		Select("*").
		Joins("left join vms on vms.user_id = packages.user_id").
		Joins("left join clusters on clusters.user_id = packages.user_id").
		Where("expires_at < ?", time.Now().AddDate(0, 0, -expirationToleranceInDays)).
		Group("packages.user_id").
		Scan(&res)
	return res, query.Error
}
