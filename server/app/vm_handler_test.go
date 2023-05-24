// Package app for c4s backend app
package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/stretchr/testify/assert"
)

func TestGetAndDeleteVMHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("Get vm: not found", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetVMHandler,
				api:         fmt.Sprintf("/%s/vm/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"virtual machine is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Det vm: not found", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.DeleteVMHandler,
				api:         fmt.Sprintf("/%s/vm/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		want := `{"err":"virtual machine is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get vm: success", func(t *testing.T) {
		vm := models.VM{
			ID:        1,
			UserID:    user.ID.String(),
			Name:      "new-vm",
			YggIP:     "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}
		err = app.db.CreateVM(&vm)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetVMHandler,
				api:         fmt.Sprintf("/%s/vm/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Get vm: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.DeleteVMHandler,
				api:         fmt.Sprintf("/%s/vm/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Get vm: vm not belong to user", func(t *testing.T) {
		vm := models.VM{
			ID:        1,
			UserID:    "userID",
			Name:      "new-vm",
			YggIP:     "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}
		err = app.db.CreateVM(&vm)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetVMHandler,
				api:         fmt.Sprintf("/%s/vm/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		want := `{"err":"virtual machine is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Delete vm: vm not belong to user", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.DeleteVMHandler,
				api:         fmt.Sprintf("/%s/vm/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		want := `{"err":"virtual machine is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get VM: invalid vm id", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetVMHandler,
				api:         fmt.Sprintf("/%s/vm/", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read vm id"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Delete VM: invalid vm id", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.DeleteVMHandler,
				api:         fmt.Sprintf("/%s/vm/", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read vm id"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestListAndDeleteVMsHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("Get all vms: no vms", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ListVMsHandler,
				api:         fmt.Sprintf("/%s/vm", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"msg":"no virtual machines found","data":[]}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Delete all vms: no vms", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.DeleteAllVMsHandler,
				api:         fmt.Sprintf("/%s/vm", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"msg":"Virtual machines are not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Get all vms: success", func(t *testing.T) {
		vm := models.VM{
			UserID:    user.ID.String(),
			Name:      "vm",
			YggIP:     "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}

		err = app.db.CreateVM(&vm)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ListVMsHandler,
				api:         fmt.Sprintf("/%s/vm", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		var res ResponseMsg
		err = json.Unmarshal(response.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, len(res.Data.([]interface{})), 1)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Delete all vms: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.DeleteAllVMsHandler,
				api:         fmt.Sprintf("/%s/vm", app.config.Version),
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
