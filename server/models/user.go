package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct { 
	ID        uuid.UUID `uuid.UUID gorm:"type:uuid;primary_key;"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" gorm:"unique" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	Voucher   string    `json:"voucher"`
	CreatedAt time.Time `json:"timestamp"`
	Code      int       `json:"code"`
}
