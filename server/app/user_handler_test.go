// Package app for c4s backend app
package app

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"

	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/stretchr/testify/assert"

	"github.com/codescalers/cloud4students/models"
)

var salt = []byte("saltsaltsaltsalt")
var password = "1234567"
var hashedPassword = sha256.Sum256(append(salt, []byte(password)...))
var user = &models.User{
	Name:           "name",
	Email:          "name@gmail.com",
	HashedPassword: append(salt, hashedPassword[:]...),
	TeamSize:       5,
	ProjectDesc:    "desc",
	College:        "clg",
	SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
}

func TestSignUpHandler(t *testing.T) {
	app := SetUp(t)

	// json Body of request
	signUpBody := []byte(`{
		"name": "name",
		"email": "name@gmail.com",
		"password": "1234567",
		"confirm_password": "1234567",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
	}`)

	t.Run("Sign up: success", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(signUpBody),
			handlerFunc: app.SignUpHandler,
			api:         fmt.Sprintf("/%s/user/signup", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		got := response.Body.String()
		want := `{"msg":"Verification code has been sent to name@gmail.com","data":{"timeout":60}}` + "\n"
		assert.Equal(t, got, want)
		assert.Equal(t, response.Code, http.StatusCreated)
	})

	t.Run("Sign up: user exists but not verified", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(signUpBody),
			handlerFunc: app.SignUpHandler,
			api:         fmt.Sprintf("/%s/user/signup", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"msg":"Verification code has been sent to name@gmail.com","data":{"timeout":60}}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusCreated)
	})

	t.Run("Sign up: empty data", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        nil,
			handlerFunc: app.SignUpHandler,
			api:         fmt.Sprintf("/%s/user/signup", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"failed to read sign up data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Sign up: invalid data", func(t *testing.T) {
		body := []byte(`{"name": "na"}`)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.SignUpHandler,
			api:         fmt.Sprintf("/%s/user/signup", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"invalid sign up data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Sign up: password and confirm_password don't match", func(t *testing.T) {
		body := []byte(`{
		"name": "name",
		"email": "name@gmail.com",
		"password": "12345679",
		"confirm_password": "1234567",
		"team_size":5,
		"project_desc":"desc",
		"college":"clg"
	}`)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.SignUpHandler,
			api:         fmt.Sprintf("/%s/user/signup", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"password and confirm password don't match"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Sign up: user already exists", func(t *testing.T) {
		user, err := app.db.GetUserByEmail(user.Email)
		assert.NoError(t, err)

		err = app.db.UpdateVerification(user.ID.String(), true)
		assert.NoError(t, err)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(signUpBody),
			handlerFunc: app.SignUpHandler,
			api:         fmt.Sprintf("/%s/user/signup", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"user already exists"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

}

func TestVerifySignUpCodeHandler(t *testing.T) {
	app := SetUp(t)

	user.Code = 1234
	user.Verified = false
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	verifyBody := []byte(fmt.Sprintf(`{"email": "%s", "code": %d}`, user.Email, user.Code))

	t.Run("Verify sign up: success", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(verifyBody),
			handlerFunc: app.VerifySignUpCodeHandler,
			api:         fmt.Sprintf("/%s/user/signup/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Verify sign up: add empty code", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        nil,
			handlerFunc: app.VerifySignUpCodeHandler,
			api:         fmt.Sprintf("/%s/user/signup/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Verify sign up: user not found", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{"email": "%s", "code": %d}`, "", user.Code))
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.VerifySignUpCodeHandler,
			api:         fmt.Sprintf("/%s/user/signup/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"user is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Verify sign up: user already verified", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(verifyBody),
			handlerFunc: app.VerifySignUpCodeHandler,
			api:         fmt.Sprintf("/%s/user/signup/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"account is already created"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Verify sign up: wrong code", func(t *testing.T) {
		err := app.db.UpdateVerification(user.ID.String(), false)
		assert.NoError(t, err)

		body := []byte(fmt.Sprintf(`{"email": "%s", "code": %d}`, user.Email, 0))

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.VerifySignUpCodeHandler,
			api:         fmt.Sprintf("/%s/user/signup/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"wrong code"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Verify sign up: code expired", func(t *testing.T) {
		app.config.MailSender.Timeout = 0
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(verifyBody),
			handlerFunc: app.VerifySignUpCodeHandler,
			api:         fmt.Sprintf("/%s/user/signup/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"code has expired"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestSignInHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	signInBody := []byte(`{
			"name":"name",
			"email":"name@gmail.com",
			"password":"1234567"
			}`)

	t.Run("Sign in: success", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(signInBody),
			handlerFunc: app.SignInHandler,
			api:         fmt.Sprintf("/%s/user/signin", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Sign in: wrong password", func(t *testing.T) {
		body := []byte(`{
			"name":"name",
			"email":"name@gmail.com",
			"password":"wrongpass"
		}`)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.SignInHandler,
			api:         fmt.Sprintf("/%s/user/signin", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"email or password is not correct"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Sign in: invalid data", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(nil),
			handlerFunc: app.SignInHandler,
			api:         fmt.Sprintf("/%s/user/signin", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"failed to read sign in data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Sign in: user not found", func(t *testing.T) {
		body := []byte(`{
			"name":"aaaa",
			"email":"aaaa@gmail.com",
			"password":"wrongpass"
		}`)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.SignInHandler,
			api:         fmt.Sprintf("/%s/user/signin", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"user is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})

	t.Run("Sign in: user is not verified", func(t *testing.T) {
		err := app.db.UpdateVerification(user.ID.String(), false)
		assert.NoError(t, err)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(signInBody),
			handlerFunc: app.SignInHandler,
			api:         fmt.Sprintf("/%s/user/signin", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"email is not verified yet, please check the verification email in your inbox"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestRefreshJWTHandler(t *testing.T) {
	app := SetUp(t)

	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("refresh token: success", func(t *testing.T) {
		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, 0)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.RefreshJWTHandler,
				api:         fmt.Sprintf("/%s/user/refresh_token", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
		}

		response := authorizedNoMiddlewareHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("refresh token: not expired yet", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.RefreshJWTHandler,
				api:         fmt.Sprintf("/%s/user/refresh_token", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
		}

		response := authorizedNoMiddlewareHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("refresh token: add empty token", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.RefreshJWTHandler,
				api:         fmt.Sprintf("/%s/user/refresh_token", app.config.Version),
			},
			userID: user.ID.String(),
			token:  "",
			config: app.config,
		}

		response := authorizedNoMiddlewareHandler(req)
		want := `{"err":"token is required"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestForgotPasswordHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	forgetPassBody := []byte(`{
		"email":"name@gmail.com"
	}`)

	t.Run("forgot password: success", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(forgetPassBody),
			handlerFunc: app.ForgotPasswordHandler,
			api:         fmt.Sprintf("/%s/user/forgot_password", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("forgot password: add wrong email", func(t *testing.T) {
		body := []byte(`{
			"email":"abcde@gmail.com"
		}`)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.ForgotPasswordHandler,
			api:         fmt.Sprintf("/%s/user/forgot_password", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"user is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("forgot password: failed to read data", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        nil,
			handlerFunc: app.ForgotPasswordHandler,
			api:         fmt.Sprintf("/%s/user/forgot_password", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"failed to read email data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestVerifyForgetPasswordCodeHandler(t *testing.T) {
	app := SetUp(t)

	user.Code = 1234
	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	verifyBody := []byte(fmt.Sprintf(`{"email": "%s", "code": %d}`, user.Email, user.Code))

	t.Run("verify forget password: success", func(t *testing.T) {
		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(verifyBody),
			handlerFunc: app.VerifyForgetPasswordCodeHandler,
			api:         fmt.Sprintf("/%s/user/forget_password/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("verify forget password: wrong code", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{"email": "%s", "code": %d}`, user.Email, 0))

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.VerifyForgetPasswordCodeHandler,
			api:         fmt.Sprintf("/%s/user/forget_password/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"wrong code"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("verify forget password: invalid data", func(t *testing.T) {
		body := []byte(`{
			"email":"name@gmail.com",
			"code": "1234"
		}`)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.VerifyForgetPasswordCodeHandler,
			api:         fmt.Sprintf("/%s/user/forget_password/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"failed to read password code"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("verify forget password: user not found", func(t *testing.T) {
		body := []byte(`{
			"email":"aaaa@gmail.com",
			"code": 1234
		}`)

		req := unAuthHandlerConfig{
			body:        bytes.NewBuffer(body),
			handlerFunc: app.VerifyForgetPasswordCodeHandler,
			api:         fmt.Sprintf("/%s/user/forget_password/verify_email", app.config.Version),
		}

		response := unAuthorizedHandler(req)
		want := `{"err":"user is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})
}

func TestChangePasswordHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	changePassBody := []byte(`{
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"newpass"
		}`)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("change password: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(changePassBody),
				handlerFunc: app.ChangePasswordHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"msg":"Password is updated successfully"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("change password: password and confirm password don't match", func(t *testing.T) {
		body := []byte(`{
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"oldpass"
		}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.ChangePasswordHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"password does not match confirm password"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("change password: invalid password data", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ChangePasswordHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read password data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("change password: invalid password", func(t *testing.T) {
		body := []byte(`{
		"email":"name@gmail.com",
		"password": "",
		"confirm_password": ""
		}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.ChangePasswordHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"invalid password data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("change password: user not found", func(t *testing.T) {
		body := []byte(`{
		"password":"1234567",
		"confirm_password":"1234567"
		}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.ChangePasswordHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"user is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})
}

func TestUpdateUserHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	updateBody := []byte(`{
		"name" : "newname",
		"email":"name@gmail.com",
		"password":"newpass",
		"confirm_password":"newpass"
	}`)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("Update user: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(updateBody),
				handlerFunc: app.UpdateUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Update user: nothing to update", func(t *testing.T) {
		body := []byte(`{}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Update user: invalid data", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.UpdateUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read user data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Update user: password and confirm password don't match", func(t *testing.T) {
		body := []byte(`{
		"password":"newpass",
		"confirm_password":"oldpas"
		}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"password and confirm password don't match"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Update user: invalid password", func(t *testing.T) {
		body := []byte(`{"password":"z", "confirm_password":"z"}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"invalid password"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Update user: invalid ssh key", func(t *testing.T) {
		body := []byte(`{
		"ssh_key":"k"	
		}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.UpdateUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"invalid sshKey"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Update user: wrong user ID", func(t *testing.T) {
		token, err := internal.CreateJWT("", user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer([]byte(`{}`)),
				handlerFunc: app.UpdateUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"user is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})
}

func TestGetUserHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("get user: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("user not found", func(t *testing.T) {
		token, err := internal.CreateJWT("", user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.GetUserHandler,
				api:         fmt.Sprintf("/%s/user", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"user is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})
}

func TestApplyForVoucherHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	voucherBody := []byte(`{
		"vms":10,
		"public_ips":1,
		"reason":"strongReason"
	}`)

	t.Run("Apply voucher: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(voucherBody),
				handlerFunc: app.ApplyForVoucherHandler,
				api:         fmt.Sprintf("/%s/user/apply_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Apply voucher: failed to read voucher data", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ApplyForVoucherHandler,
				api:         fmt.Sprintf("/%s/user/apply_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read voucher data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Apply voucher: user already applied before", func(t *testing.T) {
		v := models.Voucher{
			UserID:   user.ID.String(),
			Voucher:  "voucher",
			VMs:      10,
			Approved: false,
			Rejected: false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(voucherBody),
				handlerFunc: app.ApplyForVoucherHandler,
				api:         fmt.Sprintf("/%s/user/apply_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"you have already a voucher request, please wait for the confirmation mail"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestActivateVoucherHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	v := models.Voucher{
		Voucher:  "voucher",
		VMs:      2,
		Approved: true,
	}

	err = app.db.CreateVoucher(&v)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	voucherBody := []byte(fmt.Sprintf(`{"voucher" : "%s"}`, v.Voucher))

	t.Run("Activate voucher: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(voucherBody),
				handlerFunc: app.ActivateVoucherHandler,
				api:         fmt.Sprintf("/%s/user/activate_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Activate voucher: wrong voucher", func(t *testing.T) {
		body := []byte(`{"voucher" : "v"}`)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.ActivateVoucherHandler,
				api:         fmt.Sprintf("/%s/user/activate_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"user voucher is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Activate voucher: invalid voucher data", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.ActivateVoucherHandler,
				api:         fmt.Sprintf("/%s/user/activate_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read voucher data"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Activate voucher: voucher not found", func(t *testing.T) {
		body := []byte(`{"voucher" : "abcd"}`)
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.ActivateVoucherHandler,
				api:         fmt.Sprintf("/%s/user/activate_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"user voucher is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Activate voucher: voucher is rejected", func(t *testing.T) {
		v := models.Voucher{
			Voucher:  "rejected_voucher",
			VMs:      10,
			Rejected: true,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		body := []byte(fmt.Sprintf(`{"voucher" : "%s"}`, v.Voucher))
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.ActivateVoucherHandler,
				api:         fmt.Sprintf("/%s/user/activate_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"voucher is rejected"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Activate voucher: voucher is not approved yet", func(t *testing.T) {
		v := models.Voucher{
			Voucher:  "pending_voucher",
			VMs:      10,
			Approved: false,
			Rejected: false,
		}
		err = app.db.CreateVoucher(&v)
		assert.NoError(t, err)

		body := []byte(fmt.Sprintf(`{"voucher" : "%s"}`, v.Voucher))
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        bytes.NewBuffer(body),
				handlerFunc: app.ActivateVoucherHandler,
				api:         fmt.Sprintf("/%s/user/activate_voucher", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"voucher is not approved yet"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}
