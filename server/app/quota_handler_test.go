// Package app for c4s backend app
package app

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/stretchr/testify/assert"
)

func TestQuotaRouter(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("get quota: not found", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetQuotaHandler,
				api:         fmt.Sprintf("/%s/quota", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"user quota is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("get quota: success", func(t *testing.T) {
		err = app.db.CreateQuota(
			&models.Quota{
				UserID:    user.ID.String(),
				Vms:       map[time.Time]int{time.Now().Add(time.Hour): 10},
				PublicIPs: 1,
			},
		)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetQuotaHandler,
				api:         fmt.Sprintf("/%s/quota", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}
