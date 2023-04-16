package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateJWT(t *testing.T) {
	t.Run("create jwt token", func(t *testing.T) {
		token, err := CreateJWT("1", "email@gmail.com", "secret", 60)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

}
