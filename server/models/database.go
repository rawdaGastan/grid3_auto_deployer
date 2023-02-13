package models

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Voucher  string `json:"voucher"`
}

type Quota struct {
	UserID string `json:"userID"`
	Vms    int    `json:"vms"`
	K8s    int    `json:"k8s"`
}

func ConnectDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("./database.db"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to Connect to database")
	}

	return db, err
}
