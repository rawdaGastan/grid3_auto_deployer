package internal

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Run("check password is correct", func(t *testing.T) {
		password := "strongPassword1234"
		hashed, err := HashAndSaltPassword(password, "salt")
		if err != nil {
			t.Error(err)
		}
		valid := VerifyPassword(hashed, password, "salt")
		if !valid {
			t.Errorf("password not correct")
		}

	})

	t.Run("check password is not correct", func(t *testing.T) {
		password1 := "password1234"
		password2 := "password2345"
		hashed, err := HashAndSaltPassword(password1, "salt")
		if err != nil {
			t.Error(err)
		}
		valid := VerifyPassword(hashed, password2, "salt")
		if valid {
			t.Errorf("password not correct")
		}

	})

}
