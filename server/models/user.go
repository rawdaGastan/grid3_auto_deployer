// Package models for database models
package models

import (
	"time"
)

// User struct holds data of users
type User struct {
	ID             uint64    `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Name           string    `json:"name" binding:"required"`
	Email          string    `json:"email" gorm:"unique" binding:"required"`
	HashedPassword string    `json:"hashedPassword" binding:"required"`
	Voucher        string    `json:"voucher"`
	UpdatedAt      time.Time `json:"timestamp"`
	Code           int       `json:"code"`
	Verified       bool      `json:"verified"`
}

//TODO: add ssh key added when sign up && update
