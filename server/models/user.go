package models

import "time"

type User struct {
	ID        uint64    `json:"id" gorm:"unique;primaryKey`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" gorm:"unique" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	Voucher   string    `json:"voucher"`
	CreatedAt time.Time `json:"timestamp"`
	Code      int       `json:"code"`
}
