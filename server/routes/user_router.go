// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rawdaGastan/cloud4students/validator"
)

// SignUpInput struct for data needed when user creates account
type SignUpInput struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" gorm:"unique" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// VerifyCodeInput struct takes verification code from user
type VerifyCodeInput struct {
	Email string `json:"email" binding:"required"`
	Code  int    `json:"code" binding:"required"`
}

// SignInInput struct for data needed when user sign in
type SignInInput struct {
	Email    string `json:"email" gorm:"unique" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
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

// AddVoucherInput struct for voucher applied by user
type AddVoucherInput struct {
	Voucher string `json:"voucher" binding:"required"`
}

// SignUpHandler creates account for user
func (r *Router) SignUpHandler(w http.ResponseWriter, req *http.Request) {
	var signUp SignUpInput
	err := json.NewDecoder(req.Body).Decode(&signUp)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	// validate mail
	err = validator.ValidateMail(signUp.Email)
	if err != nil {
		writeErrResponse(w, fmt.Sprintf("Email '%s' isn't valid: %v", signUp.Email, err))
		return
	}

	//validate password
	err = validator.ValidatePassword(signUp.Password)
	if err != nil {
		writeErrResponse(w, fmt.Sprintf("Password isn't valid, error: %v", err))
		return
	}

	// password and confirm password should match
	if signUp.Password != signUp.ConfirmPassword {
		writeErrResponse(w, "Password and confirm password don't match")
		return
	}

	user, getErr := r.db.GetUserByEmail(signUp.Email)
	var code int
	// check if user already exists and verified
	if getErr == nil {
		if user.Verified {
			writeErrResponse(w, "User already exists")
			return
		}
	}

	// send verification code if user is not verified or not exist
	code, err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.Password, signUp.Email, r.config.MailSender.Timeout)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	// update code if user is not verified but exists
	if getErr == nil {
		if !user.Verified {
			_, err = r.db.UpdateUserByID(user.ID.String(), "", "", "", time.Now(), code)
			if err != nil {
				writeErrResponse(w, err.Error())
				return
			}
		}
	}

	// check if user doesn't exist
	if getErr != nil {
		// hash password
		hashedPassword, err := internal.HashPassword(signUp.Password)
		if err != nil {
			writeErrResponse(w, err.Error())
			return
		}

		u := models.User{
			Name:           signUp.Name,
			Email:          signUp.Email,
			HashedPassword: hashedPassword,
			Verified:       false,
			Code:           code,
			SSHKey:         user.SSHKey,
		}

		err = r.db.CreateUser(&u)
		if err != nil {
			writeErrResponse(w, err.Error())
			return
		}

		// create empty quota
		quota := models.Quota{
			UserID: u.ID.String(),
			Vms:    0,
			K8s:    0,
		}
		err = r.db.CreateQuota(&quota)
		if err != nil {
			writeErrResponse(w, err.Error())
			return
		}
	}

	writeMsgResponse(w, "Verification code has been sent to "+signUp.Email, "")
}

// VerifySignUpCodeHandler gets verification code to create user
func (r *Router) VerifySignUpCodeHandler(w http.ResponseWriter, req *http.Request) {

	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	user, err := r.db.GetUserByEmail(data.Email)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}

	if user.Verified {
		writeErrResponse(w, "Account is already created")
		return
	}

	if user.Code != data.Code {
		writeErrResponse(w, "Wrong code")
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		writeErrResponse(w, "Code has expired")
		return
	}
	err = r.db.UpdateVerification(user.ID.String(), true)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}
	writeMsgResponse(w, "Account is created successfully", map[string]string{"user_id": user.ID.String()})
}

// SignInHandler allows user to sign in to the system
func (r *Router) SignInHandler(w http.ResponseWriter, req *http.Request) {

	var input SignInInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	user, err := r.db.GetUserByEmail(input.Email)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}

	if !user.Verified {
		writeErrResponse(w, "User is not verified yet")
		return
	}

	match := internal.VerifyPassword(user.HashedPassword, input.Password)
	if !match {
		writeErrResponse(w, "Password is not correct")
		return
	}

	token, err := internal.CreateJWT(user.ID.String(), user.Email, r.config.Token.Secret, r.config.Token.Timeout)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}
	writeMsgResponse(w, "User is signed in successfully", map[string]string{"access_token": token})
}

// RefreshJWTHandler refreshes the user's token
func (r *Router) RefreshJWTHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		writeErrResponse(w, "Token is required")
		return
	}
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.config.Token.Secret), nil
	})
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}
	if !tkn.Valid {
		writeErrResponse(w, fmt.Sprintf("Token '%s' is invalid", reqToken))
		return
	}

	// if token didn't expire
	if time.Until(claims.ExpiresAt.Time) < time.Duration(r.config.Token.Timeout)*time.Minute {
		writeMsgResponse(w, "Access Token still valid", map[string]string{"access_token": reqToken, "refresh_token": reqToken})
		return
	}

	expirationTime := time.Now().Add(time.Duration(r.config.Token.Timeout) * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newToken, err := token.SignedString([]byte(r.config.Token.Secret))
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}
	writeMsgResponse(w, "Token is refreshed successfully", map[string]string{"access_token": reqToken, "refresh_token": newToken})
}

// ForgotPasswordHandler sends user verification code
func (r *Router) ForgotPasswordHandler(w http.ResponseWriter, req *http.Request) {

	var email EmailInput
	err := json.NewDecoder(req.Body).Decode(&email)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	user, err := r.db.GetUserByEmail(email.Email)
	if err != nil {
		writeNotFoundResponse(w, fmt.Sprintf("User is not found, error occurred %v", err))
		return
	}

	// send verification code
	code, err := internal.SendMail(r.config.MailSender.Email, r.config.MailSender.Password, email.Email, r.config.MailSender.Timeout)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	_, err = r.db.UpdateUserByID(user.ID.String(), "", "", "", time.Now(), code)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}
	writeMsgResponse(w, "Verification code has been sent to "+email.Email, "")
}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (r *Router) VerifyForgetPasswordCodeHandler(w http.ResponseWriter, req *http.Request) {

	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	user, err := r.db.GetUserByEmail(data.Email)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}

	if user.Code != data.Code {
		writeErrResponse(w, "Wrong code")
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.MailSender.Timeout) * time.Minute).Before(time.Now()) {
		writeErrResponse(w, "Code has expired")
		return
	}

	writeMsgResponse(w, "Code is verified", map[string]string{"user_id": user.ID.String()})
}

// ChangePasswordHandler changes password of user
func (r *Router) ChangePasswordHandler(w http.ResponseWriter, req *http.Request) {
	var data ChangePasswordInput
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	if data.ConfirmPassword != data.Password {
		writeErrResponse(w, "Password does not match confirm password")
		return
	}

	// hash password
	hashedPassword, err := internal.HashPassword(data.Password)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	err = r.db.UpdatePassword(data.Email, hashedPassword)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	writeMsgResponse(w, "Password is updated successfully", "")
}

// UpdateUserHandler updates user's data
func (r *Router) UpdateUserHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	input := UpdateUserInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	updates := 0

	var hashedPassword string
	if len(strings.TrimSpace(input.Password)) != 0 {
		updates++
		// password and confirm password should match
		if input.Password != input.ConfirmPassword {
			writeErrResponse(w, "Password and confirm password don't match")
			return
		}

		//validate passwords
		err = validator.ValidatePassword(input.Password)
		if err != nil {
			writeErrResponse(w, fmt.Sprintf("error: %v password isn't valid", err))
			return
		}

		// hash password
		hashedPassword, err = internal.HashPassword(input.Password)
		if err != nil {
			writeErrResponse(w, err.Error())
			return
		}
	}

	if len(strings.TrimSpace(input.SSHKey)) != 0 {
		updates++
		/*if err := validator.ValidateSSHKey(input.SSHKey); err != nil {
			writeErrResponse(w, err.Error())
			return
		}*/
	}

	if len(strings.TrimSpace(input.Name)) != 0 {
		updates++
	}

	if updates == 0 {
		writeMsgResponse(w, "Nothing to update", "")
	}

	userID, err = r.db.UpdateUserByID(userID, input.Name, hashedPassword, input.SSHKey, time.Time{}, 0)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	writeMsgResponse(w, "User is updated successfully", map[string]string{"user_id": userID})
}

// GetUserHandler returns user by its idx
func (r *Router) GetUserHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeNotFoundResponse(w, err.Error())
		return
	}
	writeMsgResponse(w, "User exists", map[string]interface{}{"user": user})
}

// ActivateVoucherHandler makes user adds voucher to his account
func (r *Router) ActivateVoucherHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input AddVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	oldQuota, err := r.db.GetUserQuota(userID)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	voucherQuota, err := r.db.GetVoucher(input.Voucher)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	if voucherQuota.Used {
		writeErrResponse(w, "Voucher is already used")
		return
	}

	err = r.db.AddUserVoucher(userID, input.Voucher)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	err = r.db.UpdateUserQuota(userID, oldQuota.Vms+voucherQuota.VMs, oldQuota.K8s+voucherQuota.K8s)
	if err != nil {
		writeErrResponse(w, err.Error())
		return
	}

	writeMsgResponse(w, "Voucher is applied successfully", "")
}
