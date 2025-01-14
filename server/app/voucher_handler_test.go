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

func TestGenerateVoucherHandler(t *testing.T) {
	app := SetUp(t)

	user.Admin = true
	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	voucherBody := []byte(`{
		"length": 5,
		"balance": 10
	}`)

	t.Run("Generate voucher: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(voucherBody),
				handlerFunc: app.GenerateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher", app.config.Version),
			},

			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusCreated)
	})

	t.Run("Generate voucher: invalid data", func(t *testing.T) {
		body := []byte(`{
			"length": 2,
			"balance": 1
		}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.GenerateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher", app.config.Version),
			},

			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"invalid voucher data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Generate voucher: failed to read voucher data", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GenerateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher", app.config.Version),
			},

			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read voucher data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestListVouchersHandler(t *testing.T) {
	app := SetUp(t)

	user.Admin = true
	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("List vouchers: no vouchers found", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ListVouchersHandler,
				api:         fmt.Sprintf("/%s/voucher", app.config.Version),
			},

			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"msg":"Vouchers are not found","data":[]}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("List vouchers: success", func(t *testing.T) {
		v := models.Voucher{
			UserID:   user.ID.String(),
			Voucher:  "voucher",
			Approved: true,
		}

		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ListVouchersHandler,
				api:         fmt.Sprintf("/%s/voucher", app.config.Version),
			},

			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}

func TestUpdateVoucherHandler(t *testing.T) {
	app := SetUp(t)

	user.Admin = true
	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	v := models.Voucher{
		UserID:  user.ID.String(),
		Voucher: "voucher",
	}
	err = app.db.CreateVoucher(&v)
	assert.NoError(t, err)

	t.Run("Update voucher: success", func(t *testing.T) {
		body := []byte(`{"approved": true}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher/1", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		want := `{"msg":"Update mail has been sent to the user"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Update voucher: voucher already approved", func(t *testing.T) {
		body := []byte(`{"approved": true}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher/1", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		want := `{"err":"voucher is already approved"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, 400)
	})

	t.Run("Reject voucher: success", func(t *testing.T) {
		body := []byte(`{"approved": false}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher/1", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		want := `{"msg":"Update mail has been sent to the user"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Update voucher: failed to read voucher id", func(t *testing.T) {
		body := []byte(`{"approved": false}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher/", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read voucher id"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Update voucher: failed to read voucher data", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.UpdateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher/1", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read voucher update data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Update voucher: voucher not found", func(t *testing.T) {
		body := []byte(`{"approved": false}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher/2", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  2,
		}

		response := authorizedHandler(req)
		want := `{"err":"voucher is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Update voucher: voucher is already rejected", func(t *testing.T) {
		body := []byte(`{"approved": false}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateVoucherHandler,
				api:         fmt.Sprintf("/%s/voucher/1", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
			varID:  1,
		}

		response := authorizedHandler(req)
		want := `{"err":"voucher is already rejected"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, 400)
	})
}

func TestApproveAllVouchers(t *testing.T) {
	app := SetUp(t)

	user.Admin = true
	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("approve all: no vouchers found", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ApproveAllVouchersHandler,
				api:         fmt.Sprintf("/%s/voucher", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("approve all: success", func(t *testing.T) {
		v := models.Voucher{
			UserID:  user.ID.String(),
			Voucher: "voucher",
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ApproveAllVouchersHandler,
				api:         fmt.Sprintf("/%s/voucher", app.config.Version),
			},
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"msg":"All vouchers are approved and confirmation mails has been sent to the users"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}
