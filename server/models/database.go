// Package models for database models
package models

import (
	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

// DB struct hold db instance
type DB struct {
	db *gorm.DB
}

// NewDB creates new DB
func NewDB() DB {
	return DB{}
}

// Connect connects to database file
func (d *DB) Connect(file string) error {
	gormDB, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return err
	}
	d.db = gormDB
	return nil
}

// Migrate migrates db schema
func (d *DB) Migrate() error {
	err := d.db.AutoMigrate(
		&User{}, &State{}, &Card{}, &Invoice{}, &VM{}, &K8sCluster{}, &Master{}, &Worker{},
		&Voucher{}, &Maintenance{}, &Notification{}, &NextLaunch{},
	)
	if err != nil {
		return err
	}

	if err := d.CreateState(); err != nil {
		return err
	}

	// add maintenance
	if err := d.db.Delete(&Maintenance{}, "1 = 1").Error; err != nil {
		return err
	}
	// add next launch
	if err := d.db.Delete(&NextLaunch{}, "1 = 1").Error; err != nil {
		return err
	}
	if err := d.db.Create(&NextLaunch{Launched: true}).Error; err != nil {
		return err
	}
	return d.db.Create(&Maintenance{}).Error
}
