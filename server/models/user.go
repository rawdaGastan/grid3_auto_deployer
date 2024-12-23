// Package models for database models
package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User struct holds data of users
type User struct {
	ID                     uuid.UUID `gorm:"primary_key; unique; type:uuid; column:id"`
	StripeCustomerID       string    `json:"stripe_customer_id"`
	StripeDefaultPaymentID string    `json:"stripe_default_payment_id"`
	FirstName              string    `json:"first_name" binding:"required"`
	LastName               string    `json:"last_name" binding:"required"`
	Email                  string    `json:"email" gorm:"unique" binding:"required"`
	HashedPassword         []byte    `json:"hashed_password" binding:"required"`
	UpdatedAt              time.Time `json:"updated_at"`
	Code                   int       `json:"code"`
	SSHKey                 string    `json:"ssh_key"`
	Verified               bool      `json:"verified"`
	// checks if user type is admin
	Admin          bool    `json:"admin"`
	Balance        float64 `json:"balance"`
	VoucherBalance float64 `json:"voucher_balance"`
}

// BeforeCreate generates a new uuid per user
func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	user.ID = id
	return
}

func (user *User) Name() string {
	return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
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
func (d *DB) ListAllUsers() ([]User, error) {
	var res []User
	return res, d.db.Where("verified = true").Find(&res).Error
}

// ListAdmins gets all admins
func (d *DB) ListAdmins() ([]User, error) {
	var admins []User
	return admins, d.db.Where("admin = true and verified = true").Find(&admins).Error
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

// UpdateUserPassword updates password of user
func (d *DB) UpdateUserPassword(email string, password []byte) error {
	var res User
	result := d.db.Model(&res).Where("email = ?", email).Update("hashed_password", password)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// UpdateUserByID updates information of user. empty and unchanged fields are not updated.
func (d *DB) UpdateUserByID(user User) error {
	return d.db.Model(&User{}).Where("id = ?", user.ID.String()).Updates(user).Error
}

// UpdateAdminUserByID updates admin information of user.
func (d *DB) UpdateAdminUserByID(id string, admin bool) error {
	return d.db.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{"admin": admin, "updated_at": time.Now()}).Error
}

// UpdateUserPaymentMethod updates user payment method ID
func (d *DB) UpdateUserPaymentMethod(id string, paymentID string) error {
	var res User
	return d.db.Model(&res).Where("id = ?", id).Update("stripe_default_payment_id", paymentID).Error
}

// UpdateUserBalance updates user balance
func (d *DB) UpdateUserBalance(id string, balance float64) error {
	var res User
	return d.db.Model(&res).Where("id = ?", id).Update("balance", balance).Error
}

// UpdateUserVoucherBalance updates user voucher balance
func (d *DB) UpdateUserVoucherBalance(id string, balance float64) error {
	var res User
	return d.db.Model(&res).Where("id = ?", id).Update("voucher_balance", balance).Error
}

// UpdateUserVerification updates if user is verified or not
func (d *DB) UpdateUserVerification(id string, verified bool) error {
	var res User
	return d.db.Model(&res).Where("id = ?", id).Update("verified", verified).Error
}

// DeleteUser deletes user by its id
func (d *DB) DeleteUser(id string) error {
	var user User
	return d.db.Where("id = ?", id).Delete(&user).Error
}
