// Package models for database models
package models

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm/clause"

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
	err := d.db.AutoMigrate(&User{}, &Quota{}, &VM{}, &K8sCluster{}, &Master{}, &Worker{}, &Voucher{})
	if err != nil {
		return err
	}

	return nil
}

// CreateUser creates new user
func (d *DB) CreateUser(u User) error {
	result := d.db.Create(&u)
	return result.Error
}

// GetUserByEmail returns user by its email
func (d *DB) GetUserByEmail(email string) (User, error) {
	var res User
	query := d.db.First(&res, "email = ?", email)
	if query.Error != nil {
		return User{}, query.Error
	}

	return res, nil
}

// GetUserByID returns user by its id
func (d *DB) GetUserByID(id string) (User, error) {
	var res User
	query := d.db.First(&res, "id = ?", id)
	if query.Error != nil {
		return User{}, query.Error
	}

	return res, nil

}

// GetCodeByEmail returns verification code for unit testing
func (d *DB) GetCodeByEmail(email string) (int, error) {
	var res User
	query := d.db.First(&res, "email = ?", email)
	if query.Error != nil {
		return 0, query.Error
	}
	return res.Code, nil
}

// UpdatePassword updates password of user
func (d *DB) UpdatePassword(email string, password string) error {
	var res User
	result := d.db.Model(&res).Where("email = ?", email).Update("hashed_password", password)

	return result.Error
}

// UpdateUserByID updates information of user
func (d *DB) UpdateUserByID(id string, name string, password string, sshKey string, updatedAt time.Time, code int) (string, error) {
	var res User
	if name != "" {
		result := d.db.Model(&res).Where("id = ?", id).Update("name", name)
		if result.Error != nil {
			return "", result.Error
		}
	}
	if password != "" {
		result := d.db.Model(&res).Where("id = ?", id).Update("hashed_password", password)
		if result.Error != nil {
			return "", result.Error
		}
	}
	if sshKey != "" {
		result := d.db.Model(&res).Where("id = ?", id).Update("ssh_key", sshKey)
		if result.Error != nil {
			return "", result.Error
		}
	}
	if !updatedAt.IsZero() {
		result := d.db.Model(&res).Where("id = ?", id).Update("updated_at", updatedAt)
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
func (d *DB) UpdateVerification(id string, verified bool) error {
	var res User
	result := d.db.Model(&res).Where("id=?", id).Update("verified", verified)
	return result.Error
}

// AddUserVoucher applies voucher for user
func (d *DB) AddUserVoucher(id string, voucher string) error {
	var res User
	result := d.db.Model(&res).Where("id = ?", id).Update("voucher", voucher)

	if result.Error != nil {
		return result.Error
	}

	return d.DeactivateVoucher(voucher)
}

// CreateVM creates new vm
func (d *DB) CreateVM(vm VM) error {
	result := d.db.Create(&vm)
	return result.Error

}

// GetVMByID return vm by its id
func (d *DB) GetVMByID(id int) (VM, error) {
	var vm VM
	query := d.db.Model(VM{ID: id}).First(&vm)
	if query.Error != nil {
		return vm, query.Error
	}

	return vm, nil
}

// GetAllVms returns all vms of user
func (d *DB) GetAllVms(userID string) ([]VM, error) {
	var vms []VM
	result := d.db.Where("user_id = ?", userID).Find(&vms)
	if result.Error != nil {
		return []VM{}, result.Error
	}
	return vms, result.Error
}

// DeleteVMByID deletes vm by its id
func (d *DB) DeleteVMByID(id int) error {
	var vm VM
	result := d.db.Delete(&vm, id)
	return result.Error
}

// DeleteAllVms deletes all vms of user
func (d *DB) DeleteAllVms(userID string) error {
	var vms []VM
	result := d.db.Clauses(clause.Returning{}).Where("user_id = ?", userID).Delete(&vms)
	return result.Error
}

// CreateQuota creates a new quota
func (d *DB) CreateQuota(q Quota) error {
	result := d.db.Create(&q)
	return result.Error
}

// UpdateUserQuota updates quota
func (d *DB) UpdateUserQuota(userID string, vms, k8s int) error {
	quota := Quota{userID, vms, k8s}
	return d.db.Debug().Model(Quota{}).Where("user_id = ?", userID).Updates(quota).Error
}

// GetUserQuota gets user quota available (vms and k8s)
func (d *DB) GetUserQuota(userID string) (Quota, error) {
	var res Quota
	query := d.db.First(&res, "user_id = ?", userID)
	if query.Error != nil {
		return res, query.Error
	}

	return res, query.Error
}

// CreateVoucher creates a new voucher
func (d *DB) CreateVoucher(v Voucher) error {
	result := d.db.Create(&v)
	return result.Error
}

// GetVoucher gets voucher
func (d *DB) GetVoucher(voucher string) (Voucher, error) {
	var res Voucher
	query := d.db.First(&res, "voucher = ?", voucher)
	if query.Error != nil {
		return res, query.Error
	}

	return res, query.Error
}

// DeactivateVoucher if it is used
func (d *DB) DeactivateVoucher(voucher string) error {
	return d.db.Debug().Model(Voucher{}).Where("voucher = ?", voucher).Update("used", true).Error
}

// CreateK8s creates a new k8s cluster
func (d *DB) CreateK8s(k K8sCluster) error {
	result := d.db.Create(&k)
	return result.Error
}

// CreateWorker creates a new k8s worker
func (d *DB) CreateWorker(k Worker) error {
	result := d.db.Create(&k)
	return result.Error
}

// GetK8s gets a k8s cluster
func (d *DB) GetK8s(id int) (K8sCluster, error) {
	var k8s K8sCluster
	err := d.db.First(&k8s, id).Error
	if err != nil {
		return K8sCluster{}, err
	}
	var master Master
	err = d.db.Model(&k8s).Association("Master").Find(&master)
	if err != nil {
		return K8sCluster{}, err
	}
	var workers []Worker
	err = d.db.Model(&k8s).Association("Workers").Find(&workers)
	if err != nil {
		return K8sCluster{}, err
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
