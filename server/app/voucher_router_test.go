// Package app for c4s backend app
package app

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGenerateVoucherHandler(t *testing.T) {
	app := SetUp(t)
	admin := models.User{
		Name:           "admin",
		Email:          "admin@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		Admin:          true,
	}
	err := app.db.CreateUser(&admin)
	assert.NoError(t, err)

	t.Run("generate voucher ", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"length": 5,
		"vms": 10,
		"public_ips": 1
		}`)

		request := httptest.NewRequest("POST", app.config.Version+"/voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.GenerateVoucherHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("generate voucher with invalid body", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"length": 1,
		"vms": 10,
		"public_ips": 1
		}`)

		request := httptest.NewRequest("POST", app.config.Version+"/voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.GenerateVoucherHandler(newRequest)
		want := `{"err":"Invalid voucher data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("failed to read voucher data", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"length": "1",
		"vms": 10,
		"public_ips": 1
		}`)

		request := httptest.NewRequest("POST", app.config.Version+"/voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.GenerateVoucherHandler(newRequest)
		want := `{"err":"Failed to read voucher data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})
}

func TestListVouchersHandler(t *testing.T) {
	app := SetUp(t)
	admin := models.User{
		Name:           "admin",
		Email:          "admin@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		Admin:          true,
	}
	err := app.db.CreateUser(&admin)
	assert.NoError(t, err)

	t.Run("no vouchers found", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", app.config.Version+"/voucher", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ListVouchersHandler(newRequest)
		want := `{"msg":"Vouchers are not found","data":[]}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("list all vouchers ", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		u := models.User{
			Name:           "name",
			Email:          "name@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
			SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		}
		err = app.db.CreateUser(&u)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		v := models.Voucher{
			UserID:    user.ID.String(),
			Voucher:   "voucher",
			VMs:       10,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  true,
			Rejected:  false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", app.config.Version+"/voucher", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ListVouchersHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

}

func TestUpdateVoucherHandler(t *testing.T) {
	app := SetUp(t)
	admin := models.User{
		Name:           "admin",
		Email:          "admin@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		Admin:          true,
	}
	err := app.db.CreateUser(&admin)
	assert.NoError(t, err)

	t.Run("user not found", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		v := models.Voucher{
			ID:        1,
			UserID:    "1234",
			Voucher:   "voucher1",
			VMs:       10,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  false,
			Rejected:  false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		body := []byte(`{
		"approved": true
		}`)

		req := httptest.NewRequest("PUT", app.config.Version+"/voucher/1", bytes.NewBuffer(body))
		request := mux.SetURLVars(req, map[string]string{
			"id": "1",
		})
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateVoucherHandler(newRequest)
		want := `{"err":"User is not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})

	t.Run("approve user voucher ", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		u := models.User{
			Name:           "name",
			Email:          "name@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
			SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		}
		err = app.db.CreateUser(&u)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		v := models.Voucher{
			ID:        2,
			UserID:    user.ID.String(),
			Voucher:   "voucher2",
			VMs:       10,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  false,
			Rejected:  false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		body := []byte(`{
		"approved": true
		}`)

		req := httptest.NewRequest("PUT", app.config.Version+"/voucher/1", bytes.NewBuffer(body))
		request := mux.SetURLVars(req, map[string]string{
			"id": "2",
		})
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateVoucherHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("reject user voucher ", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		u := models.User{
			Name:           "bbbb",
			Email:          "bbbb@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
			SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		}
		err = app.db.CreateUser(&u)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("bbbb@gmail.com")
		assert.NoError(t, err)

		v := models.Voucher{
			ID:        3,
			UserID:    user.ID.String(),
			Voucher:   "voucher3",
			VMs:       10,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  false,
			Rejected:  false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		body := []byte(`{
		"approved": false
		}`)

		req := httptest.NewRequest("PUT", app.config.Version+"/voucher/2", bytes.NewBuffer(body))
		request := mux.SetURLVars(req, map[string]string{
			"id": "3",
		})
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateVoucherHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("voucher already approved", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		u := models.User{
			Name:           "user",
			Email:          "user@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
			SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		}
		err = app.db.CreateUser(&u)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("user@gmail.com")
		assert.NoError(t, err)

		v := models.Voucher{
			ID:        4,
			UserID:    user.ID.String(),
			Voucher:   "voucher4",
			VMs:       10,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  true,
			Rejected:  false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		body := []byte(`{
		"approved": true
		}`)

		req := httptest.NewRequest("PUT", app.config.Version+"/voucher/4", bytes.NewBuffer(body))
		request := mux.SetURLVars(req, map[string]string{
			"id": "4",
		})
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateVoucherHandler(newRequest)
		want := `{"err":"Voucher is already approved"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, 400)

	})

	t.Run("failed to read voucher id", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"approved": true
		}`)

		req := httptest.NewRequest("PUT", app.config.Version+"/voucher/", bytes.NewBuffer(body))
		request := mux.SetURLVars(req, map[string]string{
			"id": "",
		})
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateVoucherHandler(newRequest)
		want := `{"err":"Failed to read voucher id"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("voucher not found", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"approved": true
		}`)

		req := httptest.NewRequest("PUT", app.config.Version+"/voucher/10", bytes.NewBuffer(body))
		request := mux.SetURLVars(req, map[string]string{
			"id": "10",
		})
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateVoucherHandler(newRequest)
		want := `{"err":"Voucher is not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("voucher is already rejected", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		u := models.User{
			Name:           "user",
			Email:          "aaaa@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
			SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		}
		err = app.db.CreateUser(&u)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("aaaa@gmail.com")
		assert.NoError(t, err)

		v := models.Voucher{
			ID:        5,
			UserID:    user.ID.String(),
			Voucher:   "voucher5",
			VMs:       10,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  false,
			Rejected:  true,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		body := []byte(`{
		"rejected": true
		}`)

		req := httptest.NewRequest("PUT", app.config.Version+"/voucher/5", bytes.NewBuffer(body))
		request := mux.SetURLVars(req, map[string]string{
			"id": "5",
		})
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateVoucherHandler(newRequest)
		want := `{"err":"Voucher is already rejected"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, 400)

	})

	t.Run("failed to read data", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"rejected": 
		}`)

		req := httptest.NewRequest("PUT", app.config.Version+"/voucher/1", bytes.NewBuffer(body))
		request := mux.SetURLVars(req, map[string]string{
			"id": "1",
		})
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateVoucherHandler(newRequest)
		want := `{"err":"Failed to read voucher update data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, 400)

	})

}

func TestApproveAllVouchers(t *testing.T) {
	app := SetUp(t)
	admin := models.User{
		Name:           "admin",
		Email:          "admin@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		Admin:          true,
	}
	err := app.db.CreateUser(&admin)
	assert.NoError(t, err)

	// t.Run("wrong access to endpoint", func(t *testing.T) {
	// 	user := models.User{
	// 		Name:           "abcd",
	// 		Email:          "abcd@gmail.com",
	// 		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
	// 		Verified:       true,
	// 		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	// 	}
	// 	err := app.db.CreateUser(&user)
	// 	assert.NoError(t, err)

	// 	u, err := app.db.GetUserByEmail("abcd@gmail.com")
	// 	assert.NoError(t, err)

	// 	token, err := internal.CreateJWT(u.ID.String(), u.Email, app.config.Token.Secret, app.config.Token.Timeout)
	// 	assert.NoError(t, err)

	// 	request := httptest.NewRequest("PUT", app.config.Version+"/voucher", nil)
	// 	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	// 	ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), u.ID.String())
	// 	newRequest := request.WithContext(ctx)
	// 	response := httptest.NewRecorder()
	// 	app.ApproveAllVouchers(newRequest)
	// 	assert.Equal(t, response.Code, http.StatusOK)

	// })

	t.Run("no vouchers found", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)
		request := httptest.NewRequest("PUT", app.config.Version+"/voucher", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ApproveAllVouchersHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("admin approve all vouchers ", func(t *testing.T) {
		userAdmin, err := app.db.GetUserByEmail("admin@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(userAdmin.ID.String(), userAdmin.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		user1 := models.User{
			Name:           "abcd",
			Email:          "abcd@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
			SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		}
		err = app.db.CreateUser(&user1)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("abcd@gmail.com")
		assert.NoError(t, err)

		v1 := models.Voucher{
			ID:        1,
			UserID:    user.ID.String(),
			Voucher:   "voucher1",
			VMs:       10,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  false,
			Rejected:  false,
		}
		err = app.db.CreateVoucher(&v1)
		assert.NoError(t, err)

		user2 := models.User{
			Name:           "aaaa",
			Email:          "aaaa@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
			SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		}
		err = app.db.CreateUser(&user2)
		assert.NoError(t, err)

		user, err = app.db.GetUserByEmail("aaaa@gmail.com")
		assert.NoError(t, err)

		v2 := models.Voucher{
			ID:        2,
			UserID:    user.ID.String(),
			Voucher:   "voucher2",
			VMs:       10,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  false,
			Rejected:  false,
		}
		err = app.db.CreateVoucher(&v2)
		assert.NoError(t, err)

		request := httptest.NewRequest("PUT", app.config.Version+"/voucher", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), userAdmin.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ApproveAllVouchersHandler(newRequest)
		want := `{"msg":"All vouchers are approved and confirmation mails has been sent to the users","data":""}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}
