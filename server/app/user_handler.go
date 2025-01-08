// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/validators"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

// SignUpInput struct for data needed when user creates account
type SignUpInput struct {
	FirstName       string `json:"first_name" binding:"required" validate:"min=3,max=20"`
	LastName        string `json:"last_name" binding:"required" validate:"min=3,max=20"`
	Email           string `json:"email" binding:"required" validate:"mail"`
	Password        string `json:"password" binding:"required" validate:"password"`
	ConfirmPassword string `json:"confirm_password" binding:"required" validate:"password"`
	SSHKey          string `json:"ssh_key"`
}

// VerifyCodeInput struct takes verification code from user
type VerifyCodeInput struct {
	Email string `json:"email" binding:"required" validate:"mail"`
	Code  int    `json:"code" binding:"required" validate:"nonzero"`
}

// SignInInput struct for data needed when user sign in
type SignInInput struct {
	Email    string `json:"email" binding:"required" validate:"mail"`
	Password string `json:"password" binding:"required" validate:"password"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email" binding:"required" validate:"mail"`
	Password        string `json:"password" binding:"required" validate:"password"`
	ConfirmPassword string `json:"confirm_password" binding:"required" validate:"password"`
}

// UpdateUserInput struct for user to updates his data
type UpdateUserInput struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	SSHKey          string `json:"ssh_key"`
}

// EmailInput struct for user when forgetting password
type EmailInput struct {
	Email string `json:"email" binding:"required" validate:"mail"`
}

// ApplyForVoucherInput struct for user to apply for voucher
type ApplyForVoucherInput struct {
	Balance uint64 `json:"balance" binding:"required" validate:"min=0"`
	Reason  string `json:"reason" binding:"required" validate:"nonzero"`
}

// AddVoucherInput struct for voucher applied by user
type AddVoucherInput struct {
	Voucher string `json:"voucher" binding:"required" validate:"nonzero"`
}

type CodeTimeout struct {
	Timeout int `json:"timeout" binding:"required" validate:"nonzero"`
}

type AccessTokenResponse struct {
	Token   string `json:"access_token"`
	Timeout int    `json:"timeout"`
}

type RefreshTokenResponse struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Timeout int    `json:"timeout"`
}

type clientSecretResponse struct {
	ClientSecret string `json:"client_secret"`
}

// SignUpHandler creates account for user
// Example endpoint: Register a new user
// @Summary Register a new user
// @Description Register a new user
// @Tags User
// @Accept  json
// @Produce  json
// @Param registration body SignUpInput true "User registration input"
// @Success 201 {object} CodeTimeout
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/signup [post]
func (a *App) SignUpHandler(req *http.Request) (interface{}, Response) {
	var signUp SignUpInput
	err := json.NewDecoder(req.Body).Decode(&signUp)

	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read sign up data"))
	}

	err = validator.Validate(signUp)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid sign up data"))
	}

	// password and confirm password should match
	if signUp.Password != signUp.ConfirmPassword {
		return nil, BadRequest(errors.New("password and confirm password don't match"))
	}

	user, getErr := a.db.GetUserByEmail(signUp.Email)
	// check if user already exists and verified
	if getErr != gorm.ErrRecordNotFound {
		if user.Verified {
			return nil, BadRequest(errors.New("user already exists"))
		}
	}

	if len(strings.TrimSpace(signUp.SSHKey)) != 0 {
		if err := validators.ValidateSSH(signUp.SSHKey); err != nil {
			log.Error().Err(err).Send()
			return nil, BadRequest(errors.New("invalid sshKey"))
		}
	}

	// send verification code if user is not verified or not exist
	code := internal.GenerateRandomCode()
	subject, body := internal.SignUpMailContent(code, a.config.MailSender.Timeout, fmt.Sprintf("%s %s", signUp.FirstName, signUp.LastName), a.config.Server.Host)
	err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, signUp.Email, subject, body, "")
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	hashedPassword, err := internal.HashAndSaltPassword([]byte(signUp.Password))
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	u := models.User{
		FirstName:      signUp.FirstName,
		LastName:       signUp.LastName,
		Email:          signUp.Email,
		HashedPassword: hashedPassword,
		Code:           code,
		SSHKey:         signUp.SSHKey,
		Admin:          internal.Contains(a.config.Admins, signUp.Email),
	}

	// update code if user is not verified but exists
	if getErr != gorm.ErrRecordNotFound {
		if !user.Verified {
			u.ID = user.ID
			u.UpdatedAt = time.Now()
			err = a.db.UpdateUserByID(u)
			if err != nil {
				log.Error().Err(err).Send()
				return nil, InternalServerError(errors.New(internalServerErrorMsg))
			}
		}
	}

	// check if user doesn't exist
	if getErr != nil {
		err = a.db.CreateUser(&u)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	return ResponseMsg{
		Message: "Verification code has been sent to " + signUp.Email,
		Data:    CodeTimeout{Timeout: a.config.MailSender.Timeout},
	}, Created()
}

// VerifySignUpCodeHandler gets verification code to create user
// Example endpoint: Verify new user's registration
// @Summary Verify new user's registration
// @Description Verify new user's registration
// @Tags User
// @Accept  json
// @Produce  json
// @Param code body VerifyCodeInput true "Verification code input"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/signup/verify_email [post]
func (a *App) VerifySignUpCodeHandler(req *http.Request) (interface{}, Response) {
	var data VerifyCodeInput
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read sign up code data"))
	}

	user, err := a.db.GetUserByEmail(data.Email)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if user.Verified {
		return nil, BadRequest(errors.New("account is already created"))
	}

	if user.Code != data.Code {
		return nil, BadRequest(errors.New("wrong code"))
	}

	if user.UpdatedAt.Add(time.Duration(a.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		return nil, BadRequest(errors.New("code has expired"))
	}
	err = a.db.UpdateUserVerification(user.ID.String(), true)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}
	middlewares.UserCreations.WithLabelValues(user.ID.String(), user.Email).Inc()

	subject, body := internal.WelcomeMailContent(user.Name(), a.config.Server.Host)
	err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, user.Email, subject, body, "")
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Account is created successfully.",
	}, Created()
}

// SignInHandler allows user to sign in to the system
// Example endpoint: Sign in user
// @Summary Sign in user
// @Description Sign in user
// @Tags User
// @Accept  json
// @Produce  json
// @Param login body SignInInput true "User login input"
// @Success 201 {object} AccessTokenResponse
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/signin [post]
func (a *App) SignInHandler(req *http.Request) (interface{}, Response) {
	var input SignInInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read sign in data"))
	}

	user, err := a.db.GetUserByEmail(input.Email)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !user.Verified {
		return nil, BadRequest(errors.New("email is not verified yet, please check the verification email in your inbox"))
	}

	match := internal.VerifyPassword(user.HashedPassword, input.Password)
	if !match {
		return nil, BadRequest(errors.New("email or password is not correct"))
	}

	token, err := internal.CreateJWT(user.ID.String(), user.Email, a.config.Token.Secret, a.config.Token.Timeout)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "You are signed in successfully",
		Data:    AccessTokenResponse{Token: token, Timeout: a.config.Token.Timeout},
	}, Created()
}

// RefreshJWTHandler refreshes the user's token
// Example endpoint: Generate a refresh token
// @Summary Generate a refresh token
// @Description Generate a refresh token
// @Tags User
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 201 {object} RefreshTokenResponse
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/refresh_token [post]
func (a *App) RefreshJWTHandler(req *http.Request) (interface{}, Response) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		return nil, BadRequest(errors.New("token is required"))
	}

	if strings.TrimSpace(splitToken[1]) == "" {
		return nil, BadRequest(errors.New("token is required"))
	}
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.config.Token.Secret), nil
	})

	// if user doesn't exist
	if _, err := a.db.GetUserByID(claims.UserID); err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}

	// if token didn't expire
	if err == nil && time.Until(claims.ExpiresAt.Time) < time.Duration(a.config.Token.Timeout)*time.Minute && tkn.Valid {
		return ResponseMsg{
			Message: "Access Token is valid",
			Data:    map[string]string{"access_token": reqToken, "refresh_token": reqToken},
		}, Created()
	}

	expirationTime := time.Now().Add(time.Duration(a.config.Token.Timeout) * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newToken, err := token.SignedString([]byte(a.config.Token.Secret))
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Token is refreshed successfully",
		Data:    RefreshTokenResponse{Access: reqToken, Refresh: newToken, Timeout: a.config.Token.Timeout},
	}, Created()
}

// ForgotPasswordHandler sends user verification code
// Example endpoint: Send code to forget password email for verification
// @Summary Send code to forget password email for verification
// @Description Send code to forget password email for verification
// @Tags User
// @Accept  json
// @Produce  json
// @Param forgetPassword body EmailInput true "User forget password input"
// @Success 201 {object} CodeTimeout
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/forgot_password [post]
func (a *App) ForgotPasswordHandler(req *http.Request) (interface{}, Response) {
	var email EmailInput
	err := json.NewDecoder(req.Body).Decode(&email)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read email data"))
	}

	user, err := a.db.GetUserByEmail(email.Email)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !user.Verified {
		return nil, BadRequest(errors.New("email is not verified yet, please check the verification email in your inbox"))
	}

	// send verification code
	code := internal.GenerateRandomCode()
	subject, body := internal.ResetPasswordMailContent(code, a.config.MailSender.Timeout, user.Name(), a.config.Server.Host)
	err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, email.Email, subject, body, "")

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.db.UpdateUserByID(
		models.User{
			ID:        user.ID,
			UpdatedAt: time.Now(),
			Code:      code,
		},
	)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Verification code has been sent to " + email.Email,
		Data:    CodeTimeout{Timeout: a.config.MailSender.Timeout},
	}, Ok()
}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
// Example endpoint: Verify user's email to reset password
// @Summary Verify user's email to reset password
// @Description Verify user's email to reset password
// @Tags User
// @Accept  json
// @Produce  json
// @Param forgetPassword body VerifyCodeInput true "User Verify forget password input"
// @Success 201 {object} AccessTokenResponse
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/forget_password/verify_email [post]
func (a *App) VerifyForgetPasswordCodeHandler(req *http.Request) (interface{}, Response) {
	var data VerifyCodeInput
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read password code"))
	}

	user, err := a.db.GetUserByEmail(data.Email)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !user.Verified {
		return nil, BadRequest(errors.New("email is not verified yet, please check the verification email in your inbox"))
	}

	if user.Code != data.Code {
		return nil, BadRequest(errors.New("wrong code"))
	}

	if user.UpdatedAt.Add(time.Duration(a.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		return nil, BadRequest(errors.New("code has expired"))
	}

	// token
	token, err := internal.CreateJWT(user.ID.String(), user.Email, a.config.Token.Secret, a.config.Token.Timeout)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Code is verified",
		Data:    AccessTokenResponse{Token: token, Timeout: a.config.Token.Timeout},
	}, Ok()
}

// ChangePasswordHandler changes password of user
// Example endpoint: Change user password
// @Summary Change user password
// @Description Change user password
// @Tags User
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param password body ChangePasswordInput true "New password"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/change_password [put]
func (a *App) ChangePasswordHandler(req *http.Request) (interface{}, Response) {
	var data ChangePasswordInput
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read password data"))
	}

	err = validator.Validate(data)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid password data"))
	}

	if data.ConfirmPassword != data.Password {
		return nil, BadRequest(errors.New("password does not match confirm password"))
	}

	hashedPassword, err := internal.HashAndSaltPassword([]byte(data.Password))
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.db.UpdateUserPassword(data.Email, hashedPassword)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Password is updated successfully",
		Data:    nil,
	}, Ok()
}

// UpdateUserHandler updates user's data
// Example endpoint: Change user data
// @Summary Change user data
// @Description Change user data
// @Tags User
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param updates body UpdateUserInput true "User updates"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user [put]
func (a *App) UpdateUserHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	var input UpdateUserInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read user data"))
	}
	updates := 0

	var hashedPassword []byte
	if len(strings.TrimSpace(input.Password)) != 0 {
		updates++
		// password and confirm password should match
		if input.Password != input.ConfirmPassword {
			return nil, BadRequest(errors.New("password and confirm password don't match"))
		}

		err = validators.ValidatePass(input.Password)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, BadRequest(errors.New("invalid password"))
		}

		// hash password
		hashedPassword, err = internal.HashAndSaltPassword([]byte(input.Password))
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	if len(strings.TrimSpace(input.SSHKey)) != 0 {
		updates++
		if err := validators.ValidateSSH(input.SSHKey); err != nil {
			log.Error().Err(err).Send()
			return nil, BadRequest(errors.New("invalid sshKey"))
		}
	}

	if len(strings.TrimSpace(input.FirstName)) != 0 {
		updates++
	}

	if len(strings.TrimSpace(input.LastName)) != 0 {
		updates++
	}

	if updates == 0 {
		return ResponseMsg{
			Message: "Nothing to update",
			Data:    nil,
		}, Ok()
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}
	err = a.db.UpdateUserByID(
		models.User{
			ID:             userUUID,
			FirstName:      input.FirstName,
			LastName:       input.LastName,
			HashedPassword: hashedPassword,
			SSHKey:         input.SSHKey,
			UpdatedAt:      time.Now(),
		},
	)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "User is updated successfully",
	}, Ok()
}

// GetUserHandler returns user by its id
// Example endpoint: Get user
// @Summary Get user
// @Description Get user
// @Tags User
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user [get]
func (a *App) GetUserHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "User exists",
		Data:    map[string]interface{}{"user": user},
	}, Ok()
}

// ApplyForVoucherHandler makes user apply for voucher that would be accepted by admin
// Example endpoint: Apply for a new voucher
// @Summary Apply for a new voucher
// @Description Apply for a new voucher
// @Tags User
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param voucher body ApplyForVoucherInput true "New voucher details"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/apply_voucher [post]
func (a *App) ApplyForVoucherHandler(req *http.Request) (interface{}, Response) {
	var input ApplyForVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		return nil, BadRequest(errors.New("failed to read voucher data"))
	}

	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	userVoucher, err := a.db.GetNotUsedVoucherByUserID(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("voucher is not found"))
	}
	if userVoucher.Voucher != "" && !userVoucher.Approved && !userVoucher.Rejected {
		return nil, BadRequest(errors.New("you have already a voucher request, please wait for the confirmation mail"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid voucher data"))
	}

	// generate voucher for user but can't use it until admin approves it
	v := internal.GenerateRandomVoucher(5)
	voucher := models.Voucher{
		Voucher: v,
		UserID:  userID,
		Balance: input.Balance,
		Reason:  input.Reason,
	}

	err = a.db.CreateVoucher(&voucher)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}
	middlewares.VoucherApplied.WithLabelValues(userID, voucher.Voucher, fmt.Sprint(voucher.Balance)).Inc()

	return ResponseMsg{
		Message: "Voucher request is being reviewed, you'll receive a confirmation mail soon",
		Data:    nil,
	}, Created()
}

// ActivateVoucherHandler makes user adds voucher to his account
// Example endpoint: Activate a voucher
// @Summary Activate a voucher
// @Description Activate a voucher
// @Tags User
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param voucher body AddVoucherInput true "Voucher input"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/activate_voucher [put]
func (a *App) ActivateVoucherHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input AddVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read voucher data"))
	}

	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	voucherBalance, err := a.db.GetVoucher(input.Voucher)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user voucher is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if voucherBalance.Rejected {
		return nil, BadRequest(errors.New("voucher is rejected"))
	}

	if !voucherBalance.Approved {
		return nil, BadRequest(errors.New("voucher is not approved yet"))
	}

	if voucherBalance.Used {
		return nil, BadRequest(errors.New("voucher is already used"))
	}

	err = a.db.DeactivateVoucher(userID, input.Voucher)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	user.VoucherBalance += float64(voucherBalance.Balance)

	// try to settle old invoices
	invoices, err := a.db.ListUnpaidInvoices(user.ID.String())
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, invoice := range invoices {
		response := a.payInvoice(&user, "", voucherAndBalance, invoice.Total, invoice.ID)
		if response.Err() != nil {
			log.Error().Err(response.Err()).Send()
		}
	}

	err = a.db.UpdateUserVoucherBalance(user.ID.String(), user.VoucherBalance)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.db.UpdateUserBalance(user.ID.String(), user.Balance)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	middlewares.VoucherActivated.WithLabelValues(userID, voucherBalance.Voucher, fmt.Sprint(voucherBalance.Balance)).Inc()

	return ResponseMsg{
		Message: "Voucher is applied successfully",
		Data:    nil,
	}, Ok()
}

// Example endpoint: Charge user balance
// @Summary Charge user balance
// @Description Charge user balance
// @Tags User
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param balance body ChargeBalance true "Balance charging details"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/charge_balance [put]
func (a *App) ChargeBalance(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input ChargeBalance
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read input data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid input data"))
	}

	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	intent, err := createPaymentIntent(user.StripeCustomerID, input.PaymentMethodID, a.config.Currency, input.Amount)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	user.Balance += float64(input.Amount)

	// try to settle old invoices
	invoices, err := a.db.ListUnpaidInvoices(user.ID.String())
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, invoice := range invoices {
		response := a.payInvoice(&user, "", voucherAndBalance, invoice.Total, invoice.ID)
		if response.Err() != nil {
			log.Error().Err(response.Err()).Send()
		}
	}

	err = a.db.UpdateUserBalance(user.ID.String(), user.Balance)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.db.UpdateUserVoucherBalance(user.ID.String(), user.VoucherBalance)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Balance is charged successfully",
		Data:    clientSecretResponse{ClientSecret: intent.ClientSecret},
	}, Ok()
}

// DeleteUserHandler deletes account for user
// Example endpoint: Deletes account for user
// @Summary Deletes account for user
// @Description Deletes account for user
// @Tags User
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user [delete]
func (a *App) DeleteUserHandler(req *http.Request) (interface{}, Response) {
	// TODO: delete customer from stripe
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// 1. Create last invoice to pay if there were active deployments
	if err := a.createInvoice(user, time.Now()); err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	invoices, err := a.db.ListUnpaidInvoices(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// 2. Try to pay invoices
	for _, invoice := range invoices {
		cards, err := a.db.GetUserCards(user.ID.String())
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		// No cards option
		if len(cards) == 0 {
			response := a.payInvoice(&user, "", voucherAndBalance, invoice.Total, invoice.ID)
			if response.Err() != nil {
				log.Error().Err(response.Err()).Send()
				return nil, InternalServerError(errors.New(internalServerErrorMsg))
			}
			continue
		}

		// Use default card
		response := a.payInvoice(&user, user.StripeDefaultPaymentID, voucherAndBalanceAndCard, invoice.Total, invoice.ID)
		if response.Err() != nil {
			log.Error().Err(response.Err()).Send()
		} else {
			continue
		}

		for _, card := range cards {
			if card.PaymentMethodID == user.StripeDefaultPaymentID {
				continue
			}

			response := a.payInvoice(&user, card.PaymentMethodID, voucherAndBalanceAndCard, invoice.Total, invoice.ID)
			if response.Err() != nil {
				log.Error().Err(response.Err()).Send()
				continue
			}
			break
		}

		return nil, BadRequest(errors.New("failed to pay your invoices, please pay them first before deleting your account"))
	}

	// 3. Delete user vms
	vms, err := a.db.GetAllVms(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, vm := range vms {
		err = a.deployer.CancelDeployment(vm.ContractID, vm.NetworkContractID, "vm", vm.Name)
		if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	err = a.db.DeleteAllVms(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// 4. Delete user k8s
	clusters, err := a.db.GetAllK8s(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, cluster := range clusters {
		err = a.deployer.CancelDeployment(uint64(cluster.ClusterContract), uint64(cluster.NetworkContract), "k8s", cluster.Master.Name)
		if err != nil && !strings.Contains(err.Error(), "ContractNotExists") {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	if len(clusters) > 0 {
		err = a.db.DeleteAllK8s(userID)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	// 5. Remove cards
	cards, err := a.db.GetUserCards(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, card := range cards {
		err = detachPaymentMethod(card.PaymentMethodID)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	err = a.db.DeleteAllCards(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// 6. TODO: should invoices be deleted?

	// 7. Remove cards
	err = a.db.DeleteUser(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "User is deleted successfully",
	}, Ok()
}
