// Package models for database models
package models

import (
	"time"
)

// Package struct for user packages
type Package struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	UserID         string    `json:"user_id"`
	Vms            int       `json:"vms"`
	PublicIPs      int       `json:"public_ips"`
	VmsCount       int       `json:"vms_count"`
	PublicIPsCount int       `json:"public_ips_count"`
	PeriodInMonth  int       `json:"period"`
	Cost           float64   `json:"cost"`
	CreatedAt      time.Time `json:"Created_at"`
}

// CreatePackage creates new package
func (d *DB) CreatePackage(p *Package) error {
	return d.db.Create(&p).Error
}

// GetPkgByID return pkg by its id
func (d *DB) GetPkgByID(id int) (Package, error) {
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

// UpdateUserPackage updates quota
func (d *DB) UpdateUserPackage(userID string, vms int, publicIPs int) error {
	return d.db.Model(&Package{}).Where("user_id = ?", userID).Updates(map[string]interface{}{"vms": vms, "public_ips": publicIPs}).Error
}
