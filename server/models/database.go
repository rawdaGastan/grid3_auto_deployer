package models

import (
	"time"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

func NewDB() DB {
	return DB{}
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

func (d *DB) CreateUser(u *User) error {
	result := d.db.Create(&u)
	return result.Error
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

func (d *DB) GetAllUsers() ([]User, error) { //TODO: for testing only
	// var u User
	var users []User
	// d.db.Delete(&users, []int{1, 2, 3, 4, 5})
	result := d.db.Find(&users)
	// len := result.RowsAffected
	// fmt.Printf("len: %v\n", len)
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

func (d *DB) UpdateUserById(id string, name string, password string, voucher string, updatedAt time.Time, code int) (string, error) {
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
	if updatedAt.IsZero() {
		d.db.Model(&res).Where("id = ?", id).Update("updatedAt", updatedAt)
	}
	if code != 0 {
		d.db.Model(&res).Where("id = ?", id).Update("code", code)
	}
	return id, nil
}

func (d *DB) UpdateVerification(id string, verified bool) {
	var res *User
	d.db.Model(&res).Where("id=?", id).Update("verified", verified)
}

func (d *DB) AddVoucher(id string, voucher string) *User {
	var res *User
	d.db.Model(&res).Where("id = ?", id).Update("voucher", voucher)
	return res
}
