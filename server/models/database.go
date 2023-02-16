package main

import (
	"encoding/json"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectDB(file string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&User{})
	return db, err
}

func (newUser *User) CreateUser() (*User, error) {
	db, err := ConnectDB("./database.db")
	if err != nil {
		return nil, err
	}
	result := db.Create(&newUser)
	return newUser, result.Error
}

func (newUser *User) SignIn(email string, password string) ([]byte, error) {
	var res User
	var data []byte
	db, err := ConnectDB("./database.db")
	if err != nil {
		return nil, err
	}

	query := db.First(&res, email, password)
	if query.Error != nil {
		return nil, query.Error
	} else {
		data, err = json.Marshal(res)
	}

	return data, err
}

// func (newUSer *User) UpdateInformation() (*User, error) {
// 	db, err := ConnectDB("./database.db")
// 	if err != nil {
// 		return nil, err
// 	}


// }
