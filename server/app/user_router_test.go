// Package app for c4s backend app
package app

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/stretchr/testify/assert"

	"github.com/codescalers/cloud4students/models"
)

// SetUp sets the needed configuration for testing
func SetUp(t testing.TB) *App {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.json")
	dbPath := filepath.Join(dir, "testing.db")

	config := fmt.Sprintf(
		`
{
	"server": {
		"host": "localhost",
		"port": ":3000",
		"redisHost": "localhost",
		"redisPort": "6379",
		"redisPass": ""		
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
        "file": %s
    },
	"app.config.Version": "v1",
	"salt": "salt"
}
	`, dbPath)

	err := os.WriteFile(configPath, []byte(config), 0644)
	assert.NoError(t, err)

	app, err := NewApp(context.Background(), configPath)
	assert.NoError(t, err)

	return app
}

func TestSignUpHandler(t *testing.T) {
	app := SetUp(t)

	// json Body of request
	body := []byte(`{
		"name": "name",
		"email": "name@gmail.com",
		"password": "1234567",
		"confirm_password": "1234567",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
	}`)

	t.Run("signup successfully", func(t *testing.T) {
		request := httptest.NewRequest("POST", app.config.Version+"/user/signup", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignUpHandler(request)
		got := response.Body.String()
		want := `{"msg":"Verification code has been sent to name@gmail.com","data":{"timeout":60}}`
		if got != want {
			t.Errorf("error : got %q, want %q", got, want)
		}
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("invalid sign up data", func(t *testing.T) {
		body = []byte(`{
		"name": "",
		"email": "name@gmail.com",
		"password": "",
		"confirm_password": "",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
		
	}`)

		request := httptest.NewRequest("POST", app.config.Version+"/user/signup", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignUpHandler(request)
		want := `{"err":"Invalid sign up data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("send empty data", func(t *testing.T) {
		request := httptest.NewRequest("POST", app.config.Version+"/user/signup", nil)
		response := httptest.NewRecorder()
		app.SignUpHandler(request)
		want := `{"err":"Failed to read sign up data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("password and confirm_password don't match", func(t *testing.T) {
		body = []byte(`{
		"name": "newName",
		"email": "newname@gmail.com",
		"password": "1234567",
		"confirm_password": "7891011",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
		
	}`)

		request := httptest.NewRequest("POST", app.config.Version+"/user/signup", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignUpHandler(request)
		want := `{"err":"Password and confirm password don't match"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("user already exists", func(t *testing.T) {
		body = []byte(`{
		"name": "aaaa",
		"email": "aaaa@gmail.com",
		"password": "1234567",
		"confirm_password": "1234567",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
	}`)
		err := app.db.CreateUser(
			&models.User{Name: "aaaa", Email: "aaaa@gmail.com", HashedPassword: []byte{}, TeamSize: 5, ProjectDesc: "desc", College: "clg", Verified: true})
		assert.NoError(t, err)
		request := httptest.NewRequest("POST", app.config.Version+"/user/signup", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignUpHandler(request)
		want := `{"err":"User already exists"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("user exists but not verified", func(t *testing.T) {
		err := app.db.CreateUser(&models.User{Name: "person", Email: "person@gmail.com", HashedPassword: []byte{}, TeamSize: 5, ProjectDesc: "desc", College: "clg", Verified: false})
		assert.NoError(t, err)
		body := []byte(`{
		"name": "person",
		"email": "person@gmail.com",
		"password": "1234567",
		"confirm_password": "1234567",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
	}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/signup", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignUpHandler(request)
		want := `{"msg":"Verification code has been sent to person@gmail.com","data":{"timeout":60}}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)

	})

}

func TestVerifySignUpCodeHandler(t *testing.T) {
	app := SetUp(t)
	body := []byte(`{
		"name":"name",
		"email":"name@gmail.com",
		"password":"strongpass",
		"confirm_password":"strongpass",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
	}`)
	request1 := httptest.NewRequest("POST", app.config.Version+"/user/signup", bytes.NewBuffer(body))
	response1 := httptest.NewRecorder()
	app.SignUpHandler(request1)
	assert.Equal(t, response1.Code, http.StatusOK)

	code, err := app.db.GetCodeByEmail("name@gmail.com")
	if err != nil {
		t.Error(err)
	}
	t.Run("verify code ", func(t *testing.T) {
		data := fmt.Sprintf(`{
			"email":"name@gmail.com",
			"code": %d
		}`, code)
		body = []byte(data)
		request2 := httptest.NewRequest("POST", app.config.Version+"/user/signup/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		app.VerifySignUpCodeHandler(request2)
		assert.Equal(t, response2.Code, http.StatusOK)

	})

	t.Run("add empty code", func(t *testing.T) {
		request2 := httptest.NewRequest("POST", app.config.Version+"/user/signup/verify_email", nil)
		response2 := httptest.NewRecorder()
		app.VerifySignUpCodeHandler(request2)
		assert.Equal(t, response2.Code, http.StatusBadRequest)
	})

	t.Run("user not found", func(t *testing.T) {
		body := []byte(`{
			"email":"user@gmail.com",
			"code": 1234
		}`)
		request2 := httptest.NewRequest("POST", app.config.Version+"/user/signup/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		app.VerifySignUpCodeHandler(request2)
		want := `{"err":"User is not found"}`
		assert.Equal(t, response2.Body.String(), want)
		assert.Equal(t, response2.Code, http.StatusNotFound)
	})

	t.Run("user already verified", func(t *testing.T) {
		err := app.db.CreateUser(&models.User{Name: "person", Email: "person@gmail.com", HashedPassword: []byte{}, TeamSize: 5, ProjectDesc: "desc", College: "clg", Verified: true})
		assert.NoError(t, err)
		body := []byte(`{
			"email":"person@gmail.com",
			"code": 1234
		}`)
		want := `{"err":"Account is already created"}`
		request2 := httptest.NewRequest("POST", app.config.Version+"/user/signup/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		app.VerifySignUpCodeHandler(request2)
		assert.Equal(t, response2.Body.String(), want)
		assert.Equal(t, response2.Code, http.StatusBadRequest)

	})

	t.Run("wrong code", func(t *testing.T) {
		err := app.db.CreateUser(&models.User{Name: "new-person", Email: "newperson@gmail.com", HashedPassword: []byte{}, TeamSize: 5, ProjectDesc: "desc", College: "clg", Verified: false, Code: 0000})
		assert.NoError(t, err)
		body := []byte(`{
			"email":"newperson@gmail.com",
			"code": 1234
		}`)
		want := `{"err":"Wrong code"}`
		request := httptest.NewRequest("POST", app.config.Version+"/user/signup/verify_email", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.VerifySignUpCodeHandler(request)
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("code expired", func(t *testing.T) {
		err := app.db.CreateUser(&models.User{Name: "newp", Email: "newp@gmail.com", HashedPassword: []byte{}, TeamSize: 5, ProjectDesc: "desc", College: "clg", Verified: false, Code: 1234, UpdatedAt: time.Now().Add(-time.Hour * 25)})
		assert.NoError(t, err)
		body := []byte(`{
			"email":"newp@gmail.com",
			"code": 1234
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/signup/verify_email", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.VerifySignUpCodeHandler(request)
		want := `{"err":"Code has expired"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})
}

func TestSignInHandler(t *testing.T) {
	app := SetUp(t)

	hashed, err := internal.HashAndSaltPassword([]byte("strongpass"))
	assert.NoError(t, err)

	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: hashed,
		Verified:       true,
	}
	err = app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("signIn successfully", func(t *testing.T) {
		body := []byte(`{
			"name":"name",
			"email":"name@gmail.com",
			"password":"strongpass"
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignInHandler(request)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("signIn with wrong password", func(t *testing.T) {
		body := []byte(`{
			"name":"name",
			"email":"name@gmail.com",
			"password":"wrongpass"
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignInHandler(request)
		got := response.Body.String()
		want := `{"err":"Password is not correct"}`
		if got != want {
			t.Errorf("error: got %q want %q", got, want)
		}
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("failed to read signIn data", func(t *testing.T) {
		body := []byte(`{
			"name":"name",
			"email":name@gmail.com,
			"password":"wrongpass"
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignInHandler(request)
		want := `{"err":"Failed to read sign in data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("user not found", func(t *testing.T) {
		body := []byte(`{
			"name":"aaaa",
			"email":"aaaa@gmail.com",
			"password":"wrongpass"
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignInHandler(request)
		want := `{"err":"User is not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})

	t.Run("user's not verified yet", func(t *testing.T) {
		err := app.db.CreateUser(&models.User{Name: "new-person", Email: "newperson@gmail.com", HashedPassword: []byte{}, TeamSize: 5, ProjectDesc: "desc", College: "clg", Verified: false, Code: 0000})
		assert.NoError(t, err)
		body := []byte(`{
			"name":"new-person",
			"email":"newperson@gmail.com",
			"password":"1234567"
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.SignInHandler(request)
		want := `{"err":"Email is not verified yet, please check the verification email in your inbox"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})
}

func TestRefreshJWTHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("refresh token not expired yet", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("POST", app.config.Version+"/user/refresh_token", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		response := httptest.NewRecorder()
		app.RefreshJWTHandler(request)
		assert.Equal(t, response.Code, http.StatusOK)
	})
	t.Run("add empty token", func(t *testing.T) {
		request := httptest.NewRequest("POST", app.config.Version+"/user/refresh_token", nil)
		response := httptest.NewRecorder()
		app.RefreshJWTHandler(request)
		want := `{"err":"Token is required"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("refresh expired token", func(t *testing.T) {
		err := app.db.CreateUser(&models.User{Name: "newp", Email: "newp@gmail.com", HashedPassword: []byte{}, TeamSize: 5, ProjectDesc: "desc", College: "clg", Verified: true, Code: 1234})
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("newp@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, 0)
		assert.NoError(t, err)

		request := httptest.NewRequest("POST", app.config.Version+"/user/refresh_token", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		response := httptest.NewRecorder()
		app.RefreshJWTHandler(request)
		assert.Equal(t, response.Code, http.StatusOK)
	})

}

func TestForgotPasswordHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("forgot password", func(t *testing.T) {
		body := []byte(`{
			"email":"name@gmail.com"
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/forgot_password", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.ForgotPasswordHandler(request)
		got := response.Body.String()
		want := `{"msg":"Verification code has been sent to name@gmail.com","data":{"timeout":60}}`
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("add wrong email", func(t *testing.T) {
		body := []byte(`{
			"email":"abcde@gmail.com"
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/forgot_password", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.ForgotPasswordHandler(request)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("failed to read data", func(t *testing.T) {
		body := []byte(`{
			"email":abcde@gmail.com
		}`)
		request := httptest.NewRequest("POST", app.config.Version+"/user/forgot_password", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.ForgotPasswordHandler(request)
		want := `{"err":"Failed to read email data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})
}

func TestVerifyForgetPasswordCodeHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	body := []byte(`{
			"email":"name@gmail.com"
		}`)
	request1 := httptest.NewRequest("POST", app.config.Version+"/user/forgot_password", bytes.NewBuffer(body))
	response1 := httptest.NewRecorder()
	app.ForgotPasswordHandler(request1)
	assert.Equal(t, response1.Code, http.StatusOK)

	t.Run("verify code", func(t *testing.T) {
		code, err := app.db.GetCodeByEmail("name@gmail.com")
		assert.NoError(t, err)

		data := fmt.Sprintf(`{
			"email":"name@gmail.com",
			"code": %d
		}`, code)
		body = []byte(data)
		request2 := httptest.NewRequest("POST", app.config.Version+"/user/forget_password/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		app.VerifyForgetPasswordCodeHandler(request2)
		assert.Equal(t, response2.Code, http.StatusOK)
	})

	t.Run("add wrong code", func(t *testing.T) {
		data := fmt.Sprintf(`{
			"email":"name@gmail.com",
			"code": %d
		}`, 00000)
		body = []byte(data)
		request2 := httptest.NewRequest("POST", app.config.Version+"/user/forget_password/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		app.VerifyForgetPasswordCodeHandler(request2)
		assert.Equal(t, response2.Code, http.StatusBadRequest)
	})

	t.Run("failed to read code", func(t *testing.T) {
		body := []byte(`{
			"email":"name@gmail.com",
			"code": "1234"
		}`)

		request := httptest.NewRequest("POST", app.config.Version+"/user/forget_password/verify_email", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.VerifyForgetPasswordCodeHandler(request)
		want := `{"err":"Failed to read password code"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("user not found", func(t *testing.T) {
		body := []byte(`{
			"email":"aaaa@gmail.com",
			"code": 1234
		}`)

		request := httptest.NewRequest("POST", app.config.Version+"/user/forget_password/verify_email", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.VerifyForgetPasswordCodeHandler(request)
		want := `{"err":"User is not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})
}

func TestChangePasswordHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("change password", func(t *testing.T) {
		body := []byte(`{
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"newpass"
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.ChangePasswordHandler(request)
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

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.ChangePasswordHandler(request)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("failed to read password data", func(t *testing.T) {
		body := []byte(`{
		"email":name@gmail.com,
		"password":"newpass",
		"confirm_password":"oldpass"
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.ChangePasswordHandler(request)
		want := `{"err":"Failed to read password data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("invalid password", func(t *testing.T) {
		body := []byte(`{
		"email":"name@gmail.com",
		"password":"",
		"confirm_password":""
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.ChangePasswordHandler(request)
		want := `{"err":"Invalid password data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("user not found", func(t *testing.T) {
		body := []byte(`{
		"password":"newpass",
		"confirm_password":"newpass"
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		app.ChangePasswordHandler(request)
		want := `{"err":"User is not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})
}

func TestUpdateUserHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("update data of user", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"name" : "newname",
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"newpass"
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateUserHandler(newRequest)
		got := response.Body.String()
		want := fmt.Sprintf(`{"msg":"User is updated successfully","data":{"user_id":"%s"}}`, user.ID.String())
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("add empty data", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(nil))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateUserHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("password and confirm password don't match", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"name" : "newname",
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"oldpas"
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateUserHandler(newRequest)
		want := `{"err":"Password and confirm password don't match"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("invalid password", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"name" : "newname",
		"email":"name@gmail.com",
		"password":"z",
		"confirm_password":"z"
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateUserHandler(newRequest)
		want := `{"err":"Invalid password"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("invalid ssh key", func(t *testing.T) {
		u.SSHKey = "z"
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"name" : "newname",
		"email":"name@gmail.com",
		"ssh_key":"k"	
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateUserHandler(newRequest)
		want := `{"err":"Invalid sshKey"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("nothing to update", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{}`)
		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateUserHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("wrong user ID", func(t *testing.T) {
		token, err := internal.CreateJWT("", u.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"name" : "newname",
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"newpass"
		}`)
		request := httptest.NewRequest("PUT", app.config.Version+"/user", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), "")
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.UpdateUserHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusInternalServerError)
	})

}

func TestGetUserHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("get user", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", app.config.Version+"/user", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.GetUserHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("user not found", func(t *testing.T) {

		token, err := internal.CreateJWT("2", u.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", app.config.Version+"/user", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), "2")
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.GetUserHandler(newRequest)
		want := `{"err":"User is not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})
}

func TestApplyForVoucherHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("failed to read voucher data", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
			"vms":10
			"public_ips":0
			"reason:"strongReason"

		}`)

		request := httptest.NewRequest("POST", app.config.Version+"/user/apply_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ApplyForVoucherHandler(newRequest)
		want := `{"err":"Failed to read voucher data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("apply for voucher", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
			"vms":10,
			"public_ips":1,
			"reason":"strongReason"

		}`)

		request := httptest.NewRequest("POST", app.config.Version+"/user/apply_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ApplyForVoucherHandler(newRequest)
		want := `{"msg":"Voucher request is being reviewed, you'll receive a confirmation mail soon","data":""}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("user already applied before", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		v := models.Voucher{
			UserID:   user.ID.String(),
			Voucher:  "voucher",
			VMs:      10,
			Approved: false,
			Rejected: false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		body := []byte(`{
			"vms":10,
			"public_ips":1,
			"reason":"strongReason"

		}`)

		request := httptest.NewRequest("POST", app.config.Version+"/user/apply_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ApplyForVoucherHandler(newRequest)
		want := `{"err":"You have already a voucher request, please wait for the confirmation mail"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

}

func TestActivateVoucherHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("activate voucher ", func(t *testing.T) {
		v := models.Voucher{
			Voucher:  "voucher",
			VMs:      10,
			Approved: true,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		err = app.db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    0,
			},
		)
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : "voucher"
		}`)
		request := httptest.NewRequest("PUT", app.config.Version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ActivateVoucherHandler(newRequest)
		got := response.Body.String()
		want := `{"msg":"Voucher is applied successfully","data":""}`
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("apply wrong voucher ", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		err = app.db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    10,
			},
		)
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : "voucher"
		}`)
		request := httptest.NewRequest("PUT", app.config.Version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ActivateVoucherHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("failed to read voucher data", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : voucher
		}`)
		request := httptest.NewRequest("PUT", app.config.Version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ActivateVoucherHandler(newRequest)
		want := `{"err":"Failed to read voucher data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("user quota not found", func(t *testing.T) {
		u := models.User{
			Name:           "ffff",
			Email:          "ffff@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
		}
		err := app.db.CreateUser(&u)
		assert.NoError(t, err)

		v := models.Voucher{
			Voucher:  "testing",
			VMs:      10,
			Approved: true,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		user, err := app.db.GetUserByEmail("ffff@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : "testing"
		}`)
		request := httptest.NewRequest("PUT", app.config.Version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ActivateVoucherHandler(newRequest)
		want := `{"err":"User quota not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})

	t.Run("voucher not found", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		err = app.db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    10,
			},
		)
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : "abcd"
		}`)
		request := httptest.NewRequest("PUT", app.config.Version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ActivateVoucherHandler(newRequest)
		want := `{"err":"User voucher not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})

	t.Run("voucher is rejected", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		v := models.Voucher{
			Voucher:  "newvoucher",
			VMs:      10,
			Rejected: true,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		err = app.db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    0,
			},
		)
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : "newvoucher"
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ActivateVoucherHandler(newRequest)
		want := `{"err":"Voucher is rejected"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("voucher is not approved yet", func(t *testing.T) {

		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		v := models.Voucher{
			Voucher:  "123456",
			VMs:      10,
			Approved: false,
			Rejected: false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		err = app.db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    10,
			},
		)
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"voucher" : "123456"
		}`)

		request := httptest.NewRequest("PUT", app.config.Version+"/user/activate_voucher", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.ActivateVoucherHandler(newRequest)
		want := `{"err":"Voucher is not approved yet"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

}
