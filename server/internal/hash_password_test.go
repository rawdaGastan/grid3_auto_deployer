package internal

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Run("check password is correct", func(t *testing.T) {
		password := "strongPassword1234"
		hashed, err := HashPassword(password)
		if err != nil {
			t.Error(err)
		}
		valid := VerifyPassword(hashed, password)
		if !valid {
			t.Errorf("password not correct")
		}

	})

	t.Run("check password is not correct", func(t *testing.T) {
		password1 := "password1234"
		password2 := "password2345"
		hashed, err := HashPassword(password1)
		if err != nil {
			t.Error(err)
		}
		valid := VerifyPassword(hashed, password2)
		if valid {
			t.Errorf("password not correct")
		}

	})

}
