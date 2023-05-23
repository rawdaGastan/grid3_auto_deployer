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
	Name            string `json:"name" binding:"required" validate:"min=3,max=20"`
	Email           string `json:"email" binding:"required" validate:"mail"`
	Password        string `json:"password" binding:"required" validate:"password"`
	ConfirmPassword string `json:"confirm_password" binding:"required" validate:"password"`
	TeamSize        int    `json:"team_size" binding:"required" validate:"min=1,max=20"`
	ProjectDesc     string `json:"project_desc" binding:"required" validate:"nonzero"`
	College         string `json:"college" binding:"required" validate:"nonzero"`
}

// VerifyCodeInput struct takes verification code from user
type VerifyCodeInput struct {
	Email string `json:"email" binding:"required"`
	Code  int    `json:"code" binding:"required"`
}

// SignInInput struct for data needed when user sign in
type SignInInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required" validate:"password"`
	ConfirmPassword string `json:"confirm_password" binding:"required" validate:"password"`
}

// UpdateUserInput struct for user to updates his data
type UpdateUserInput struct {
	Name            string `json:"name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	SSHKey          string `json:"ssh_key"`
}

// EmailInput struct for user when forgetting password
type EmailInput struct {
	Email string `json:"email" binding:"required"`
}

// ApplyForVoucherInput struct for user to apply for voucher
type ApplyForVoucherInput struct {
	VMs       int    `json:"vms" binding:"required" validate:"min=0"`
	PublicIPs int    `json:"public_ips" binding:"required" validate:"min=0"`
	Reason    string `json:"reason" binding:"required" validate:"nonzero"`
}

// AddVoucherInput struct for voucher applied by user
type AddVoucherInput struct {
	Voucher string `json:"voucher" binding:"required"`
}

// SignUpHandler creates account for user
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

	// send verification code if user is not verified or not exist
	code := internal.GenerateRandomCode()
	subject, body := internal.SignUpMailContent(code, a.config.MailSender.Timeout, signUp.Name)
	err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, signUp.Email, subject, body)
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
		Name:           signUp.Name,
		Email:          signUp.Email,
		HashedPassword: hashedPassword,
		Code:           code,
		SSHKey:         user.SSHKey,
		TeamSize:       signUp.TeamSize,
		ProjectDesc:    signUp.ProjectDesc,
		College:        signUp.College,
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

		// create empty quota
		quota := models.Quota{
			UserID: u.ID.String(),
			Vms:    0,
		}
		err = a.db.CreateQuota(&quota)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	return ResponseMsg{
		Message: "Verification code has been sent to " + signUp.Email,
		Data:    map[string]int{"timeout": a.config.MailSender.Timeout},
	}, Created()
}

// VerifySignUpCodeHandler gets verification code to create user
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
	err = a.db.UpdateVerification(user.ID.String(), true)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}
	middlewares.UserCreations.WithLabelValues(user.ID.String(), user.Email, user.College, fmt.Sprint(user.TeamSize)).Inc()

	return ResponseMsg{
		Message: "Account is created successfully",
		Data:    map[string]string{"user_id": user.ID.String()},
	}, Ok()
}

// SignInHandler allows user to sign in to the system
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
		return nil, BadRequest(errors.New("password is not correct"))
	}

	token, err := internal.CreateJWT(user.ID.String(), user.Email, a.config.Token.Secret, a.config.Token.Timeout)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "You are signed in successfully",
		Data:    map[string]string{"access_token": token},
	}, Ok()
}

// RefreshJWTHandler refreshes the user's token
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
		}, Ok()
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
		Data:    map[string]string{"access_token": reqToken, "refresh_token": newToken},
	}, Ok()
}

// ForgotPasswordHandler sends user verification code
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
	subject, body := internal.ResetPasswordMailContent(code, a.config.MailSender.Timeout, user.Name)
	err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, email.Email, subject, body)

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
		Data:    map[string]int{"timeout": a.config.MailSender.Timeout},
	}, Ok()
}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (a *App) VerifyForgetPasswordCodeHandler(req *http.Request) (interface{}, Response) {
	data := VerifyCodeInput{}
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
		Data:    map[string]string{"access_token": token},
	}, Ok()
}

// ChangePasswordHandler changes password of user
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

	err = a.db.UpdatePassword(data.Email, hashedPassword)
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
func (a *App) UpdateUserHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	input := UpdateUserInput{}
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

	if len(strings.TrimSpace(input.Name)) != 0 {
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
			Name:           input.Name,
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
		Data:    map[string]string{"user_id": userID},
	}, Ok()
}

// GetUserHandler returns user by its idx
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
		Voucher:   v,
		UserID:    userID,
		VMs:       input.VMs,
		Reason:    input.Reason,
		PublicIPs: input.PublicIPs,
	}

	err = a.db.CreateVoucher(&voucher)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}
	middlewares.VoucherApplied.WithLabelValues(userID, voucher.Voucher, fmt.Sprint(voucher.VMs), fmt.Sprint(voucher.PublicIPs)).Inc()

	return ResponseMsg{
		Message: "Voucher request is being reviewed, you'll receive a confirmation mail soon",
		Data:    nil,
	}, Ok()
}

// ActivateVoucherHandler makes user adds voucher to his account
func (a *App) ActivateVoucherHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input AddVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read voucher data"))
	}

	oldQuota, err := a.db.GetUserQuota(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user quota is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	voucherQuota, err := a.db.GetVoucher(input.Voucher)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user voucher is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if voucherQuota.Rejected {
		return nil, BadRequest(errors.New("voucher is rejected"))
	}

	if !voucherQuota.Approved {
		return nil, BadRequest(errors.New("voucher is not approved yet"))
	}

	if voucherQuota.Used {
		return nil, BadRequest(errors.New("voucher is already used"))
	}

	err = a.db.DeactivateVoucher(userID, input.Voucher)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.db.UpdateUserQuota(userID, oldQuota.Vms+voucherQuota.VMs, oldQuota.PublicIPs+voucherQuota.PublicIPs)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}
	middlewares.VoucherActivated.WithLabelValues(userID, voucherQuota.Voucher, fmt.Sprint(voucherQuota.VMs), fmt.Sprint(voucherQuota.PublicIPs)).Inc()

	return ResponseMsg{
		Message: "Voucher is applied successfully",
		Data:    nil,
	}, Ok()
}
