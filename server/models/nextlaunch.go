package models

import "time"

// NextLaunch struct for next launch revealing
type NextLaunch struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Launched  bool      `json:"launched"`
	UpdatedAt time.Time `json:"updated_at"`
}
