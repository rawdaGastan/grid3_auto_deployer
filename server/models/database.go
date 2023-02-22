package models

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

const expirationTimeout = 5 * time.Minute

type DB struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewDB() DB {
	c := cache.New(expirationTimeout, expirationTimeout)
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
	d.cache.Set(key, data, expirationTimeout)
}

func (d *DB) GetCache(key string) (User, error) {
	data, found := d.cache.Get(key)
	if !found {
		return User{}, errors.New("Time out")
	}
	value, ok := data.(User)
	if !ok {
		return User{}, errors.New("Failed to get data")
	}
	return value, nil
}

func (d *DB) CreateUser(u *User) (*User, error) {
	result := d.db.Create(&u)
	return u, result.Error
}

func (d *DB) GetUserByEmail(email string, secret string) (User, error) {
	var res User
	query := d.db.First(&res, "email = ?", email)
	if query.Error != nil {
		fmt.Printf("query.Error: %v\n", query.Error)
		return User{}, query.Error
	}

	return res, nil
}

func (d *DB) UpdatePassword(email string, password string) error {
	var res User
	d.db.Model(&res).Where("email = ?", email).Update("password", password)

	return nil
}

func (d *DB) UpdateData(u *User) (*User, error) {
	var res *User
	d.db.Model(&res).Where("email = ?", u.Email).Update("password", u.Password)
	d.db.Model(&res).Where("name = ?", u.Email).Update("password", u.Name)
	d.db.Model(&res).Where("voucher = ?", u.Email).Update("password", u.Voucher)

	return res, nil
}
