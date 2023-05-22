package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	t.Run("check password is correct", func(t *testing.T) {
		password := "strongPassword1234"
		hashed, err := HashAndSaltPassword([]byte(password))
		assert.NoError(t, err)

		valid := VerifyPassword(hashed, password)
		assert.True(t, valid)
	})

	t.Run("check password is not correct", func(t *testing.T) {
		password1 := "password1234"
		password2 := "password2345"
		hashed, err := HashAndSaltPassword([]byte(password1))
		assert.NoError(t, err)

		valid := VerifyPassword(hashed, password2)
		if valid {
			t.Errorf("password not correct")
		}

	})

}
