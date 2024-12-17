// Package models for database models
package models

import (
	"time"

	"gorm.io/gorm/clause"
)

// Voucher struct holds data of vouchers
type Voucher struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"  binding:"required"`
	Voucher   string    `json:"voucher" gorm:"unique"`
	Balance   uint64    `json:"balance" binding:"required"`
	Reason    string    `json:"reason" binding:"required"`
	Used      bool      `json:"used" binding:"required"`
	Approved  bool      `json:"approved" binding:"required"`
	Rejected  bool      `json:"rejected" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateVoucher creates a new voucher
func (d *DB) CreateVoucher(v *Voucher) error {
	return d.db.Create(&v).Error
}

// GetVoucher gets voucher
func (d *DB) GetVoucher(voucher string) (Voucher, error) {
	var res Voucher
	return res, d.db.First(&res, "voucher = ?", voucher).Error
}

// GetVoucherByID gets voucher by ID
func (d *DB) GetVoucherByID(id int) (Voucher, error) {
	var res Voucher
	return res, d.db.First(&res, id).Error
}

// ListAllVouchers returns all vouchers to admin
func (d *DB) ListAllVouchers() ([]Voucher, error) {
	var res []Voucher
	return res, d.db.Find(&res).Error
}

// UpdateVoucher approves voucher by voucher id
func (d *DB) UpdateVoucher(id int, approved bool) (Voucher, error) {
	var voucher Voucher
	query := d.db.First(&voucher, id)
	if query.Error != nil {
		return voucher, query.Error
	}

	return voucher, d.db.Model(&voucher).Clauses(clause.Returning{}).Updates(map[string]interface{}{"approved": approved, "rejected": !approved}).Error
}

// GetAllPendingVouchers gets all pending vouchers
func (d *DB) GetAllPendingVouchers() ([]Voucher, error) {
	var vouchers []Voucher
	return vouchers, d.db.Where("approved = false and rejected = false").Find(&vouchers).Error
}

// DeactivateVoucher if it is used
func (d *DB) DeactivateVoucher(userID string, voucher string) error {
	return d.db.Model(Voucher{}).Where("voucher = ?", voucher).Updates(map[string]interface{}{"used": true, "user_id": userID}).Error
}

// GetNotUsedVoucherByUserID returns not used voucher by its user id
func (d *DB) GetNotUsedVoucherByUserID(id string) (Voucher, error) {
	var res Voucher
	return res, d.db.Last(&res, "user_id = ? AND used = false", id).Error
}
