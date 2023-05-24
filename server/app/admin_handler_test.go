// Package app for c4s backend app
package app

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/stretchr/testify/assert"
)

func TestGetAllUsersHandler(t *testing.T) {
	app := SetUp(t)

	admin := models.User{
		Name:     "admin",
		Email:    "admin@gmail.com",
		Verified: true,
		Admin:    true,
	}
	err := app.db.CreateUser(&admin)
	assert.NoError(t, err)

	user, err := app.db.GetUserByEmail(admin.Email)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("Get all users: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetAllUsersHandler,
				api:         fmt.Sprintf("/%s/user/all", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := adminHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Get all users: not admin", func(t *testing.T) {
		u := models.User{
			Name:     "name",
			Email:    "name@gmail.com",
			Verified: true,
		}
		err := app.db.CreateUser(&u)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail(u.Email)
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetAllUsersHandler,
				api:         fmt.Sprintf("/%s/user/all", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := adminHandler(req)
		want := `{"err":"user 'name' doesn't have an admin access"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("Get maintainence: success", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        nil,
			handlerFunc: app.GetMaintenanceHandler,
			api:         fmt.Sprintf("/%s/maintenance", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Update maintenance: success", func(t *testing.T) {
		body := []byte(`{
		"on": true
		}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateMaintenanceHandler,
				api:         fmt.Sprintf("/%s/maintenance", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := adminHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Update maintenance: send empty body", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.UpdateMaintenanceHandler,
				api:         fmt.Sprintf("/%s/maintenance", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := adminHandler(req)
		want := `{"err":"failed to read maintenance update data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Get balance: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetBalanceHandler,
				api:         fmt.Sprintf("/%s/balance", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := adminHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}
