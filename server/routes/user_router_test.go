// Package routes for API endpoints
package routes

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/stretchr/testify/assert"

	"github.com/codescalers/cloud4students/models"
	"github.com/threefoldtech/grid3-go/deployer"
)

// SetUp sets the needed configuration for testing
func SetUp(t testing.TB) (r *Router, db models.DB, configurations internal.Configuration, version string) {
	config :=
		`
{
	"server": {
		"host": "localhost",
		"port": ":3000"
	},
	"mailSender": {
        "email": "email",
        "sendgrid_key": "my sendgrid_key",
        "timeout": 60 
    },
    "account": {
        "mnemonics": "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
		"network": "dev"
    },
	"token": {
        "secret": "secret",
        "timeout": 10
    },
	"database": {
        "file": "testing.db"
    },
	"version": "v1"
}
	`
	dir := t.TempDir()
	configPath := dir + "/config.json"
	dbPath := dir + "testing.db"

	err := os.WriteFile(configPath, []byte(config), 0644)
	assert.NoError(t, err)

	data, err := internal.ReadConfFile(configPath)
	assert.NoError(t, err)

	configuration, err := internal.ParseConf(data)
	assert.NoError(t, err)

	db = models.NewDB()
	err = db.Connect(dbPath)
	assert.NoError(t, err)

	err = db.Migrate()
	assert.NoError(t, err)

	tfPluginClient, err := deployer.NewTFPluginClient(configuration.Account.Mnemonics, "sr25519", configuration.Account.Network, "", "", "", 0, true, false)
	assert.NoError(t, err)

	version = "/" + configuration.Version
	router, err := NewRouter(configuration, db, tfPluginClient)
	assert.NoError(t, err)

	return &router, db, configuration, version
}

func TestSignUpHandler(t *testing.T) {
	router, _, _, version := SetUp(t)
	// json Body of request
	body := []byte(`{
		"name": "name",
		"email": "name@gmail.com",
		"password": "123456",
		"confirm_password": "123456",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
	}`)
	t.Run("signup successfully", func(t *testing.T) {
		request := httptest.NewRequest("POST", version+"/user/signup", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.SignUpHandler(response, request)
		got := response.Body.String()
		want := `{"msg":"Verification code has been sent to name@gmail.com","data":{"timeout":60}}`
		if got != want {
			t.Errorf("error : got %q, want %q", got, want)
		}
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("send empty data", func(t *testing.T) {
		request := httptest.NewRequest("POST", version+"/user/signup", nil)
		response := httptest.NewRecorder()
		router.SignUpHandler(response, request)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

}

func TestVerifySignUpCodeHandler(t *testing.T) {
	router, db, _, version := SetUp(t)
	body := []byte(`{
		"name":"name",
		"email":"name@gmail.com",
		"password":"strongpass",
		"confirm_password":"strongpass",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
	}`)
	request1 := httptest.NewRequest("POST", version+"/user/signup", bytes.NewBuffer(body))
	response1 := httptest.NewRecorder()
	router.SignUpHandler(response1, request1)
	assert.Equal(t, response1.Code, http.StatusOK)

	code, err := db.GetCodeByEmail("name@gmail.com")
	if err != nil {
		t.Error(err)
	}
	t.Run("verify code ", func(t *testing.T) {
		data := fmt.Sprintf(`{
			"email":"name@gmail.com",
			"code": %d
		}`, code)
		body = []byte(data)
		request2 := httptest.NewRequest("POST", version+"/user/signup/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		router.VerifySignUpCodeHandler(response2, request2)
		assert.Equal(t, response2.Code, http.StatusOK)

	})

	t.Run("add empty code", func(t *testing.T) {
		request2 := httptest.NewRequest("POST", version+"/user/signup/verify_email", nil)
		response2 := httptest.NewRecorder()
		router.VerifySignUpCodeHandler(response2, request2)
		assert.Equal(t, response2.Code, http.StatusBadRequest)
	})
}

func TestSignInHandler(t *testing.T) {
	router, db, _, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	if err != nil {
		t.Error(err)
	}

	t.Run("signIn successfully", func(t *testing.T) {
		body := []byte(`{
			"name":"name",
			"email":"name@gmail.com",
			"password":"strongpass"
		}`)
		request := httptest.NewRequest("POST", version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.SignInHandler(response, request)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("signIn with wrong password", func(t *testing.T) {
		body := []byte(`{
			"name":"name",
			"email":"name@gmail.com",
			"password":"wrongpass"
		}`)
		request := httptest.NewRequest("POST", version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.SignInHandler(response, request)
		got := response.Body.String()
		want := `{"err":"Password is not correct"}`
		if got != want {
			t.Errorf("error: got %q want %q", got, want)
		}
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})
}

func TestRefreshJWTHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("refresh jwt token", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("POST", version+"/user/refresh_token", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		response := httptest.NewRecorder()
		router.RefreshJWTHandler(response, request)
		assert.Equal(t, response.Code, http.StatusOK)
	})
	t.Run("add empty token", func(t *testing.T) {
		request := httptest.NewRequest("POST", version+"/user/refresh_token", nil)
		response := httptest.NewRecorder()
		router.RefreshJWTHandler(response, request)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

}

func TestForgotPasswordHandler(t *testing.T) {
	router, db, _, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("forgot password", func(t *testing.T) {
		body := []byte(`{
			"email":"name@gmail.com"
		}`)
		request := httptest.NewRequest("POST", version+"/user/forgot_password", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.ForgotPasswordHandler(response, request)
		got := response.Body.String()
		want := `{"msg":"Verification code has been sent to name@gmail.com","data":{"timeout":60}}`
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("add wrong email", func(t *testing.T) {
		body := []byte(`{
			"email":"abcde@gmail.com"
		}`)
		request := httptest.NewRequest("POST", version+"/user/forgot_password", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.ForgotPasswordHandler(response, request)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})
}

func TestVerifyForgetPasswordCodeHandler(t *testing.T) {
	router, db, _, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	body := []byte(`{
			"email":"name@gmail.com"
		}`)
	request1 := httptest.NewRequest("POST", version+"/user/forgot_password", bytes.NewBuffer(body))
	response1 := httptest.NewRecorder()
	router.ForgotPasswordHandler(response1, request1)
	assert.Equal(t, response1.Code, http.StatusOK)

	t.Run("verify code", func(t *testing.T) {
		code, err := db.GetCodeByEmail("name@gmail.com")
		assert.NoError(t, err)

		data := fmt.Sprintf(`{
			"email":"name@gmail.com",
			"code": %d
		}`, code)
		body = []byte(data)
		request2 := httptest.NewRequest("POST", version+"/user/forget_password/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		router.VerifyForgetPasswordCodeHandler(response2, request2)
		assert.Equal(t, response2.Code, http.StatusOK)
	})

	t.Run("add wrong code", func(t *testing.T) {
		data := fmt.Sprintf(`{
			"email":"name@gmail.com",
			"code": %d
		}`, 00000)
		body = []byte(data)
		request2 := httptest.NewRequest("POST", version+"/user/forget_password/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		router.VerifyForgetPasswordCodeHandler(response2, request2)
		assert.Equal(t, response2.Code, http.StatusBadRequest)
	})
}

func TestChangePasswordHandler(t *testing.T) {
	router, db, _, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("change password", func(t *testing.T) {
		body := []byte(`{
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"newpass"
		}`)

		request := httptest.NewRequest("PUT", version+"/user", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.ChangePasswordHandler(response, request)
		got := response.Body.String()
		want := `{"msg":"Password is updated successfully","data":""}`
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("password and confirm password don't match", func(t *testing.T) {
		body := []byte(`{
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"oldpass"
		}`)

		request := httptest.NewRequest("PUT", version+"/user", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.ChangePasswordHandler(response, request)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestUpdateUserHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("update data of user", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"name" : "newname",
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"newpass"
		}`)

		request := httptest.NewRequest("PUT", version+"/user", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.UpdateUserHandler(response, newRequest)
		got := response.Body.String()
		want := fmt.Sprintf(`{"msg":"User is updated successfully","data":{"user_id":"%s"}}`, user.ID.String())
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("add empty data", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("PUT", version+"/user", bytes.NewBuffer(nil))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.UpdateUserHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("update part of data", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"name" : "newname"
		}`)
		request := httptest.NewRequest("PUT", version+"/user", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.UpdateUserHandler(response, newRequest)
		got := response.Body.String()
		want := fmt.Sprintf(`{"msg":"User is updated successfully","data":{"user_id":"%s"}}`, user.ID.String())
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusOK)

	})

}

func TestGetUserHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("get user", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", version+"/user", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.GetUserHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}

func TestActivateVoucherHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("activate voucher ", func(t *testing.T) {
		v := models.Voucher{
			Voucher:  "voucher",
			VMs:      10,
			Approved: true,
		}
		err = db.CreateVoucher(&v)
		assert.NoError(t, err)

		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		err = db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    0,
			},
		)
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : "voucher"
		}`)
		request := httptest.NewRequest("PUT", version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.ActivateVoucherHandler(response, newRequest)
		got := response.Body.String()
		want := `{"msg":"Voucher is applied successfully","data":""}`
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("apply wrong voucher ", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		err = db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    10,
			},
		)
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : "voucher"
		}`)
		request := httptest.NewRequest("PUT", version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.ActivateVoucherHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

}
