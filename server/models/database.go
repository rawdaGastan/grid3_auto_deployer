// Package models for database models
package models

import (
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
func (d *DB) CreateUser(u *User) error {
	result := d.db.Create(&u)
	return result.Error
}

// GetUserByEmail returns user by its email
func (d *DB) GetUserByEmail(email string) (User, error) {
	var res User
	query := d.db.First(&res, "email = ?", email)
	return res, query.Error
}

// GetUserByID returns user by its id
func (d *DB) GetUserByID(id string) (User, error) {
	var res User
	query := d.db.First(&res, "id = ?", id)
	return res, query.Error
}

// ListAllUsers returns all users to admin
func (d *DB) ListAllUsers() ([]UserUsedQuota, error) {
	var res []UserUsedQuota
	query := d.db.Table("users").
		Select("*, vouchers.vms - quota.vms as used_vms, vouchers.public_ips - quota.public_ips as used_public_ips").
		Joins("left join quota on quota.user_id = users.id").
		Joins("left join vouchers on vouchers.user_id = users.id and vouchers.used = true and vouchers.approved = true").
		Where("verified = true").
		Scan(&res)
	return res, query.Error
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

// UpdateUserByID updates information of user. empty and unchanged fields are not updated.
func (d *DB) UpdateUserByID(user User) error {
	result := d.db.Model(&User{}).Where("id = ?", user.ID.String()).Updates(user)
	return result.Error
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

// GetNotUsedVoucherByUserID returns not used voucher by its user id
func (d *DB) GetNotUsedVoucherByUserID(id string) (Voucher, error) {
	var res Voucher
	query := d.db.First(&res, "user_id = ? AND used = false", id)
	return res, query.Error
}

// CreateVM creates new vm
func (d *DB) CreateVM(vm *VM) error {
	result := d.db.Create(&vm)
	return result.Error

}

// GetVMByID return vm by its id
func (d *DB) GetVMByID(id int) (VM, error) {
	var vm VM
	query := d.db.Model(VM{ID: id}).First(&vm)
	return vm, query.Error
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
func (d *DB) CreateQuota(q *Quota) error {
	result := d.db.Create(&q)
	return result.Error
}

// UpdateUserQuota updates quota
func (d *DB) UpdateUserQuota(userID string, vms int, publicIPs int) error {
	return d.db.Model(&Quota{}).Where("user_id = ?", userID).Updates(map[string]interface{}{"vms": vms, "public_ips": publicIPs}).Error
}

// GetUserQuota gets user quota available vms (vms will be used for both vms and k8s clusters)
func (d *DB) GetUserQuota(userID string) (Quota, error) {
	var res Quota
	query := d.db.First(&res, "user_id = ?", userID)
	return res, query.Error
}

// CreateVoucher creates a new voucher
func (d *DB) CreateVoucher(v *Voucher) error {
	result := d.db.Create(&v)
	return result.Error
}

// GetVoucher gets voucher
func (d *DB) GetVoucher(voucher string) (Voucher, error) {
	var res Voucher
	query := d.db.First(&res, "voucher = ?", voucher)
	return res, query.Error
}

// GetVoucherByID gets voucher by ID
func (d *DB) GetVoucherByID(id int) (Voucher, error) {
	var res Voucher
	query := d.db.First(&res, id)
	return res, query.Error
}

// ListAllVouchers returns all vouchers to admin
func (d *DB) ListAllVouchers() ([]Voucher, error) {
	var res []Voucher
	query := d.db.Find(&res)
	return res, query.Error
}

// UpdateVoucher approves voucher by voucher id
func (d *DB) UpdateVoucher(id int, approved bool) (Voucher, error) {
	var voucher Voucher
	query := d.db.First(&voucher, id).Update("approved", approved)
	return voucher, query.Error
}

// ApproveAllVouchers approves all vouchers
func (d *DB) ApproveAllVouchers() ([]Voucher, error) {
	var vouchers []Voucher
	return vouchers, d.db.Find(&vouchers).Update("approved", true).Error
}

// DeactivateVoucher if it is used
func (d *DB) DeactivateVoucher(voucher string) error {
	return d.db.Model(Voucher{}).Where("voucher = ?", voucher).Update("used", true).Error
}

// CreateK8s creates a new k8s cluster
func (d *DB) CreateK8s(k *K8sCluster) error {
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
