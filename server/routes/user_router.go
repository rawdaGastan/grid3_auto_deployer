// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rawdaGastan/cloud4students/validator"
)

// SignUpInput struct for data needed when user creates account
type SignUpInput struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" gorm:"unique" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	SSHKey          string `json:"sshKey" binding:"required"`
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
	SSHKey          string `json:"sshKey" binding:"required"`
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
		r.WriteErrResponse(w, err)
		return
	}

	// validate mail
	valid := validator.ValidateMail(signUp.Email)
	if !valid {
		r.WriteErrResponse(w, fmt.Errorf("email isn't valid %v", err))
		return
	}

	//validate password
	err = validator.ValidatePassword(signUp.Password)
	if err != nil {
		r.WriteErrResponse(w, fmt.Errorf("error: %v password isn't valid", err))
		return
	}

	// password and confirm password should match
	if signUp.Password != signUp.ConfirmPassword {
		r.WriteErrResponse(w, fmt.Errorf("password and confirm password don't match"))
		return
	}

	user, getErr := r.db.GetUserByEmail(signUp.Email)
	var code int
	// check if user already exists and verified
	if getErr == nil {
		if user.Verified {
			r.WriteMsgResponse(w, "user already exists", "")
			return
		}
	}

	// send verification code if user is not verified or not exist
	code, err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.Password, signUp.Email, r.config.Token.Timeout)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// update code if user is not verified but exists
	if getErr == nil {
		if !user.Verified {
			_, err = r.db.UpdateUserByID(user.ID.String(), "", "", time.Now(), code, "")
			if err != nil {
				r.WriteErrResponse(w, err)
				return
			}
		}
	}

	// check if user doesn't exist
	if getErr != nil {
		// hash password
		hashedPassword, err := internal.HashPassword(signUp.Password)
		if err != nil {
			r.WriteErrResponse(w, err)
			return
		}

		u := models.User{
			Name:           signUp.Name,
			Email:          signUp.Email,
			HashedPassword: hashedPassword,
			Verified:       false,
			Code:           code,
			SSHKey:         signUp.SSHKey,
		}

		fmt.Printf("code: %v\n", code) //TODO: to be removed
		err = r.db.CreateUser(&u)
		if err != nil {
			r.WriteErrResponse(w, err)
			return
		}

		// create empty quota
		quota := models.Quota{
			UserID: u.ID.String(),
			Vms:    0,
			K8s:    0,
		}
		err = r.db.CreateQuota(quota)
		if err != nil {
			r.WriteErrResponse(w, err)
			return
		}
	}

	r.WriteMsgResponse(w, "verification code has been sent to "+signUp.Email, "")
}

// VerifySignUpCodeHandler gets verification code to create user
func (r *Router) VerifySignUpCodeHandler(w http.ResponseWriter, req *http.Request) {
	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(data.Email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	if user.Verified {
		r.WriteMsgResponse(w, "account is already created", map[string]string{"user_id": user.ID.String()})
		return
	}

	if user.Code != data.Code {
		r.WriteErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.Token.Timeout) * time.Minute).Before(time.Now()) {
		r.WriteErrResponse(w, fmt.Errorf("time out"))
		return
	}
	err = r.db.UpdateVerification(user.ID.String(), true)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "account created successfully", map[string]string{"user_id": user.ID.String()})
}

// SignInHandler allows user to sign in to the system
func (r *Router) SignInHandler(w http.ResponseWriter, req *http.Request) {
	var input SignInInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(input.Email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	if !user.Verified {
		r.WriteErrResponse(w, fmt.Errorf("user is not verified yet"))
		return
	}

	match := internal.VerifyPassword(user.HashedPassword, input.Password)
	if !match {
		r.WriteErrResponse(w, fmt.Errorf("password is not correct"))
		return
	}

	token, err := internal.CreateJWT(user.ID.String(), user.Email, r.config.Token.Secret, r.config.Token.Timeout)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "signed in successfully", map[string]string{"access_token": token})
}

// RefreshJWTHandler refreshes the user's token
func (r *Router) RefreshJWTHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	claims, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	expirationTime := time.Now().Add(time.Duration(r.config.Token.Timeout) * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newToken, err := token.SignedString([]byte(r.config.Token.Secret))
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "token is refreshed successfully", map[string]string{"access_token": reqToken, "refresh_token": newToken})
}

// SignOut allows user to logout from the system by expiring his token
func (r *Router) SignOut(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	claims, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// TODO: Rawda: how you logout??
	expirationTime := time.Now()
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	r.WriteMsgResponse(w, "user logged out successfully", "")
}

// ForgotPasswordHandler sends user verification code
func (r *Router) ForgotPasswordHandler(w http.ResponseWriter, req *http.Request) {
	var email EmailInput
	err := json.NewDecoder(req.Body).Decode(&email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(email.Email)
	if err != nil {
		r.WriteErrResponse(w, fmt.Errorf("user not found %v", err))
		return
	}

	// send verification code
	code, err := internal.SendMail(r.config.MailSender.Email, r.config.MailSender.Password, email.Email, r.config.Token.Timeout)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	fmt.Printf("code: %v\n", code) //TODO: to be removed

	_, err = r.db.UpdateUserByID(user.ID.String(), "", "", time.Now(), code, "")
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "verification code has been sent to "+email.Email, "")
}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (r *Router) VerifyForgetPasswordCodeHandler(w http.ResponseWriter, req *http.Request) {
	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(data.Email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	if user.Code != data.Code {
		r.WriteErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.Token.Timeout) * time.Minute).Before(time.Now()) {
		r.WriteErrResponse(w, fmt.Errorf("time out"))
		return
	}

	r.WriteMsgResponse(w, "code is verified", map[string]string{"user_id": user.ID.String()})
}

// ChangePasswordHandler changes password of user
func (r *Router) ChangePasswordHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: Rawda: change password for verify - settings
	data := ChangePasswordInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// hash password
	hashedPassword, err := internal.HashPassword(data.Password)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	err = r.db.UpdatePassword(data.Email, hashedPassword)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	r.WriteMsgResponse(w, "Password Updated Successfully", "")
}

// UpdateUserHandler updates user's data
func (r *Router) UpdateUserHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]

	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	_, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	input := UpdateUserInput{}
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	var hashedPassword string
	if len(strings.TrimSpace(input.Password)) != 0 {
		// password and confirm password should match
		if input.Password != input.ConfirmPassword {
			r.WriteErrResponse(w, fmt.Errorf("password and confirm password don't match"))
			return
		}

		//validate passwords
		err = validator.ValidatePassword(input.Password)
		if err != nil {
			r.WriteErrResponse(w, fmt.Errorf("error: %v password isn't valid", err))
			return
		}

		// hash password
		hashedPassword, err = internal.HashPassword(input.Password)
		if err != nil {
			r.WriteErrResponse(w, err)
			return
		}
	}

	userID, err := r.db.UpdateUserByID(id, input.Name, hashedPassword, time.Time{}, 0, input.SSHKey)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	r.WriteMsgResponse(w, "user updated successfully", map[string]string{"user_id": userID})
}

// GetUserHandler returns user by its idx
func (r *Router) GetUserHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]

	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	_, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "user exists", map[string]interface{}{"user": user})
}

// AddVoucherHandler makes user adds voucher to his account
func (r *Router) ActivateVoucherHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]

	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		r.WriteErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	_, err := r.validateToken(false, reqToken, r.config.Token.Secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	var input AddVoucherInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	voucherQuota, err := r.db.GetVoucher(input.Voucher)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	err = r.db.AddUserVoucher(id, input.Voucher)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	err = r.db.UpdateUserQuota(id, voucherQuota.VMs, voucherQuota.K8s)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	r.WriteMsgResponse(w, "voucher is applied successfully", "")
}

func (router *Router) GetAllUsersHandlers(w http.ResponseWriter, r *http.Request) { //TODO: to be removed for testing only
	users, err := router.db.GetAllUsers()
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	userBytes, err := json.Marshal(users)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	w.Write(userBytes)
}
