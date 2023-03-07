// Package models for database models
package models

import (
	"fmt"
	"time"

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
	err := d.db.AutoMigrate(&User{}, &VM{})
	if err != nil {
		return err
	}
	// err = d.db.AutoMigrate(&Token{})
	// if err != nil {
	// 	return err
	// }
	// err = d.db.AutoMigrate(&Quota{})
	// if err != nil {
	// 	return err
	// }
	// err = d.db.AutoMigrate(&Kubernetes{})
	// if err != nil {
	// 	return err
	// }
	return nil

}

// CreateUser creates new user
func (d *DB) CreateUser(u *User) error {
	result := d.db.Create(&u)
	return result.Error
}

// GetUserByEmail returns user by its email
func (d *DB) GetUserByEmail(email string) (*User, error) {
	var res User
	query := d.db.First(&res, "email = ?", email)
	if query.Error != nil {
		return &res, query.Error
	}

	return &res, nil
}

// GetUserByID returns user by its id
func (d *DB) GetUserByID(id string) (*User, error) {
	var res User
	query := d.db.First(&res, "id = ?", id)
	if query.Error != nil {
		return &res, query.Error
	}

	return &res, nil

}

// UpdatePassword updates password of user
func (d *DB) UpdatePassword(email string, password string) error {
	var res User
	result := d.db.Model(&res).Where("email = ?", email).Update("password", password)

	return result.Error
}

// UpdateUserByID updates information of user
func (d *DB) UpdateUserByID(id uint64, name string, password string, voucher string, updatedAt time.Time, code int) (string, error) {
	var res *User
	if name != "" {
		result := d.db.Model(&res).Where("id = ?", id).Update("name", name)
		if result.Error != nil {
			return "", result.Error
		}
	}
	if password != "" {
		result := d.db.Model(&res).Where("id = ?", id).Update("password", password)
		if result.Error != nil {
			return "", result.Error
		}
	}
	if voucher != "" {
		result := d.db.Model(&res).Where("id = ?", id).Update("voucher", voucher)
		if result.Error != nil {
			return "", result.Error
		}
	}
	if updatedAt.IsZero() {
		result := d.db.Model(&res).Where("id = ?", id).Update("updatedAt", updatedAt)
		if result.Error != nil {
			return "", result.Error
		}
	}
	if code != 0 {
		result := d.db.Model(&res).Where("id = ?", id).Update("code", code)
		if result.Error != nil {
			return "", result.Error
		}
	}
	return string(id), nil
}

// UpdateVerification updates if user is verified or not
func (d *DB) UpdateVerification(id uint64, verified bool) error {
	var res *User
	result := d.db.Model(&res).Where("id=?", id).Update("verified", verified)
	return result.Error
}

// AddVoucher applies voucher for user
func (d *DB) AddVoucher(id string, voucher string) error {
	var res *User
	result := d.db.Model(&res).Where("id = ?", id).Update("voucher", voucher)
	return result.Error
}

// CreateVM creates new vm
func (d *DB) CreateVM(vm *VM) error {
	result := d.db.Create(&vm)
	return result.Error

}

// GetVmByID return vm by its id
func (d *DB) GetVmByID(id string) (*VM, error) {
	var vm VM
	query := d.db.First(&vm, "id = ?", id)
	if query.Error != nil {
		return &vm, query.Error
	}

	return &vm, nil
}

// GetAllVms returns all vms of user
func (d *DB) GetAllVms(userID string) ([]VM, error) {
	var vms []VM
	result := d.db.Where("userID = ?", userID).Find(&vms)
	if result.Error != nil {
		return vms, result.Error
	}
	return vms, nil
}

func (d *DB) GetAllUsers() ([]User, error) { //TODO: for testing only
	// var u User
	var users []User
	// d.db.Delete(&users, []int{1, 2, 3, 4, 5})
	result := d.db.Find(&users)
	len := result.RowsAffected
	fmt.Printf("len: %v\n", len)
	if result.Error != nil {
		return users, result.Error
	}
	return users, nil
}
