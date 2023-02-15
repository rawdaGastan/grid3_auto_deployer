package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type database struct {
	db *gorm.DB
}

func (d *database) ConnectDB(file string) (err error) {
	d.db, err = gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return err
	}
	err = d.db.AutoMigrate(&User{}, &Quota{})
	return err
}
