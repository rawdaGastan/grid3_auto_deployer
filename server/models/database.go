package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const TimeOut = 5 * time.Minute

type DB struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewDB() DB {
	c := cache.New(TimeOut, TimeOut)
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
	err = d.db.AutoMigrate(&Token{})
	if err != nil {
		return err
	}
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

func (d *DB) SetCache(key string, value interface{}) {
	d.cache.Set(key, value, TimeOut)
}

func (d *DB) GetCache(key string) (User, error) {
	data, found := d.cache.Get(key)
	if !found {
		return User{}, errors.New("Time out")
	}
	fmt.Printf("data: %v\n", data)

	value, ok := data.(User)
	if !ok {
		return User{}, errors.New("Failed to get data")
	}
	fmt.Printf("value: %v\n", value)
	return value, nil

}

func (d *DB) SignUp(u *User) (*User, error) {
	err := d.Connect("./database.db")
	if err != nil {
		return nil, err
	}

	err = d.Migrate()
	if err != nil {
		return nil, err
	}

	result := d.db.Create(&u)
	return u, result.Error
}

func (d *DB) SignIn(u *User) ([]byte, error) {
	var res User
	var data []byte
	err := d.Connect("./database.db")
	if err != nil {
		return nil, err
	}

	err = d.Migrate()
	if err != nil {
		return nil, err
	}

	query := d.db.First(&res, "email = ?", u.Email)
	if query.Error != nil {
		return nil, query.Error
	} else {
		data, _ = json.Marshal(res)
	}

	expiresAt := time.Now().Add(TimeOut).Unix()

	err = VerifyPassword(res.Password, u.Password)
	if err != nil {
		return nil, errors.Wrapf(err, "Password is not correct")
	}

	tk := Token{
		UserID: u.ID,
		Email:  u.Email,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenStr, err := token.SignedString([]byte("secret"))
	if err != nil {
		fmt.Println(err)
	}

	print(tokenStr)

	return data, err
}

func (d *DB) ChangePassword(email string, password string) error {
	err := d.Connect("./database.db")
	if err != nil {
		return err
	}

	err = d.Migrate()
	if err != nil {
		return err
	}
	var res User
	d.db.Model(&res).Where("email = ?", email).Update("password", password)

	return nil
}

func (d *DB) UpdateData(u *User) (*User, error) {
	err := d.Connect("./database.db")
	if err != nil {
		return nil, err
	}

	err = d.Migrate()
	if err != nil {
		return nil, err
	}
	var res *User
	d.db.Model(&res).Where("email = ?", u.Email).Update("password", u.Password)
	d.db.Model(&res).Where("name = ?", u.Email).Update("password", u.Name)
	d.db.Model(&res).Where("voucher = ?", u.Email).Update("password", u.Voucher)

	return res, nil

}

// TODO: its place should be outside this folder
func VerifyPassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
