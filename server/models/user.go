package models

import (
	"time"
)

// User struct holds data of users
type User struct {
	ID             string    `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	Name           string    `json:"name" binding:"required"`
	Email          string    `json:"email" gorm:"unique" binding:"required"`
	HashedPassword string    `json:"hashedPassword" binding:"required"`
	Voucher        string    `json:"voucher"`
	UpdatedAt      time.Time `json:"timestamp"`
	Code           int       `json:"code"`
	Verified       bool      `json:"verified"`
}
