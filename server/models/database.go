package models

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

type DB struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewDB() DB {
	c := cache.New(5*time.Minute, 5*time.Minute) //TODO:
	return DB{cache: c}
}

func (d *DB) Connect(file string) error {
	gormDB, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return err
	}
	d.db = gormDB
	return nil
}

func (d *DB) Migrate() error {
	err := d.db.AutoMigrate(&User{})
	if err != nil {
		return err
	}
	// err = d.db.AutoMigrate(&Token{})
	// if err != nil {
	// 	return err
	// }
	// err = d.db.AutoMigrate(&Quota{})
	// if err != nil {
	// 	return err
	// }
	// err = d.db.AutoMigrate(&VM{})
	// if err != nil {
	// 	return err
	// }
	// err = d.db.AutoMigrate(&Kubernetes{})
	// if err != nil {
	// 	return err
	// }
	return nil

}

func (d *DB) SetCache(key string, data interface{}) {
	d.cache.Set(key, data, 5*time.Minute)
}

func (d *DB) GetCache(key string) (User, error) {
	data, found := d.cache.Get(key)
	if !found {
		return User{}, errors.New("time out, data not found in cache")
	}
	value, ok := data.(User)
	if !ok {
		return User{}, errors.New("failed to get data from cache")
	}
	return value, nil
}

func (d *DB) CreateUser(u *User) (*User, error) {
	fmt.Printf("u: %v\n", u)
	result := d.db.Create(&u)
	fmt.Printf("result.Error: %v\n", result.Error)
	return u, result.Error
}

func (d *DB) GetUserByEmail(email string) (*User, error) {
	var res User
	query := d.db.First(&res, "email = ?", email)
	if query.Error != nil {
		return &res, query.Error
	}

	return &res, nil
}

func (d *DB) GetUserById(id string) (*User, error) {
	var res User
	query := d.db.First(&res, "id = ?", id)
	if query.Error != nil {
		return &res, query.Error
	}

	return &res, nil

}

func (d *DB) GetAllUsers() ([]User, error) {
	var users []User
	result := d.db.Find(&users)
	len := result.RowsAffected
	fmt.Printf("len: %v\n", len)
	if result.Error != nil {
		return users, result.Error
	}
	return users, nil
}

func (d *DB) UpdatePassword(email string, password string) error {
	var res User
	d.db.Model(&res).Where("email = ?", email).Update("password", password)

	return nil
}

func (d *DB) UpdateUserById(id string, name string, password string, voucher string) (*User, error) {
	var res *User
	if name != "" {
		d.db.Model(&res).Where("id = ?", id).Update("name", name)
	}
	if password != "" {
		d.db.Model(&res).Where("id = ?", id).Update("password", password)
	}
	if voucher != "" {
		d.db.Model(&res).Where("id = ?", id).Update("voucher", voucher)
	}
	return res, nil
}

func (d *DB) AddVoucher(id string, voucher string) *User {
	var res *User
	d.db.Model(&res).Where("id = ?", id).Update("voucher", voucher)
	return res
}
