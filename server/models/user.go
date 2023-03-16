// Package models for database models
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User struct holds data of users
type User struct {
	ID             uuid.UUID `gorm:"primary_key; unique; type:uuid; column:id"`
	Name           string    `json:"name" binding:"required"`
	Email          string    `json:"email" gorm:"unique" binding:"required"`
	HashedPassword string    `json:"hashed_password" binding:"required"`
	Voucher        string    `json:"voucher"`
	UpdatedAt      time.Time `json:"updated_at"`
	Code           int       `json:"code"`
	SSHKey         string    `json:"ssh_key"`
	Verified       bool      `json:"verified"`
	// checks if user type is admin
	Admin bool `json:"admin"`
}

// BeforeCreate generates a new uuid
func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	user.ID = id
	return
}
