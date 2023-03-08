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
		writeErrResponse(w, err)
		return
	}

	// validate mail
	err = validator.ValidateMail(signUp.Email)
	if err != nil {
		writeErrResponse(w, fmt.Errorf("email '%s' isn't valid: %v", signUp.Email, err))
		return
	}

	//validate password
	err = validator.ValidatePassword(signUp.Password)
	if err != nil {
		writeErrResponse(w, fmt.Errorf("error: %v password isn't valid", err))
		return
	}

	// password and confirm password should match
	if signUp.Password != signUp.ConfirmPassword {
		writeErrResponse(w, fmt.Errorf("password and confirm password don't match"))
		return
	}

	user, getErr := r.db.GetUserByEmail(signUp.Email)
	var code int
	// check if user already exists and verified
	if getErr == nil {
		if user.Verified {
			writeMsgResponse(w, "user already exists", "")
			return
		}
	}

	// send verification code if user is not verified or not exist
	code, err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.Password, signUp.Email, r.config.Token.Timeout)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// update code if user is not verified but exists
	if getErr == nil {
		if !user.Verified {
			_, err = r.db.UpdateUserByID(user.ID.String(), "", "", "", time.Now(), code)
			if err != nil {
				writeErrResponse(w, err)
				return
			}
		}
	}

	// check if user doesn't exist
	if getErr != nil {
		// hash password
		hashedPassword, err := internal.HashPassword(signUp.Password)
		if err != nil {
			writeErrResponse(w, err)
			return
		}

		u := models.User{
			Name:           signUp.Name,
			Email:          signUp.Email,
			HashedPassword: hashedPassword,
			Verified:       false,
			Code:           code,
		}

		fmt.Printf("code: %v\n", code) //TODO: to be removed
		err = r.db.CreateUser(&u)
		if err != nil {
			writeErrResponse(w, err)
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
			writeErrResponse(w, err)
			return
		}
	}

	writeMsgResponse(w, "verification code has been sent to "+signUp.Email, "")
}

// VerifySignUpCodeHandler gets verification code to create user
func (r *Router) VerifySignUpCodeHandler(w http.ResponseWriter, req *http.Request) {

	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(data.Email)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if user.Verified {
		writeMsgResponse(w, "account is already created", map[string]string{"user_id": user.ID.String()})
		return
	}

	if user.Code != data.Code {
		writeErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.Token.Timeout) * time.Minute).Before(time.Now()) {
		writeErrResponse(w, fmt.Errorf("time out"))
		return
	}
	err = r.db.UpdateVerification(user.ID.String(), true)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	writeMsgResponse(w, "account created successfully", map[string]string{"user_id": user.ID.String()})
}

// SignInHandler allows user to sign in to the system
func (r *Router) SignInHandler(w http.ResponseWriter, req *http.Request) {

	var input SignInInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(input.Email)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if !user.Verified {
		writeErrResponse(w, fmt.Errorf("user is not verified yet"))
		return
	}

	match := internal.VerifyPassword(user.HashedPassword, input.Password)
	if !match {
		writeErrResponse(w, fmt.Errorf("password is not correct"))
		return
	}

	token, err := internal.CreateJWT(user.ID.String(), user.Email, r.config.Token.Secret, r.config.Token.Timeout)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	if err != nil {
		writeErrResponse(w, err)
		return
	}
	writeMsgResponse(w, "signed in successfully", map[string]string{"access_token": token})
}

// RefreshJWTHandler refreshes the user's token
func (r *Router) RefreshJWTHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		writeErrResponse(w, fmt.Errorf("token is required"))
		return
	}
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.config.Token.Secret), nil
	})
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	if !tkn.Valid {
		writeErrResponse(w, fmt.Errorf("token '%s' is invalid", reqToken))
		return
	}

	// if token didn't expire
	if time.Until(claims.ExpiresAt.Time) < time.Duration(r.config.Token.Timeout)*time.Minute {
		writeMsgResponse(w, "access token is valid", map[string]string{"access_token": reqToken, "refresh_token": reqToken})
		return
	}

	expirationTime := time.Now().Add(time.Duration(r.config.Token.Timeout) * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newToken, err := token.SignedString([]byte(r.config.Token.Secret))
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	writeMsgResponse(w, "token is refreshed successfully", map[string]string{"access_token": reqToken, "refresh_token": newToken})
}

// SignOut allows user to logout from the system by expiring his token
func (r *Router) SignOut(w http.ResponseWriter, req *http.Request) {
	// TODO: Rawda: how you logout??
	/*expirationTime := time.Now()
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)*/
	writeMsgResponse(w, "user logged out successfully", "")
}

// ForgotPasswordHandler sends user verification code
func (r *Router) ForgotPasswordHandler(w http.ResponseWriter, req *http.Request) {

	var email EmailInput
	err := json.NewDecoder(req.Body).Decode(&email)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(email.Email)
	if err != nil {
		writeNotFoundResponse(w, fmt.Errorf("user not found %v", err))
		return
	}

	// send verification code
	code, err := internal.SendMail(r.config.MailSender.Email, r.config.MailSender.Password, email.Email, r.config.Token.Timeout)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	fmt.Printf("code: %v\n", code) //TODO: to be removed

	_, err = r.db.UpdateUserByID(user.ID.String(), "", "", "", time.Now(), code)
	if err != nil {
		writeErrResponse(w, err)
		return
	}
	writeMsgResponse(w, "verification code has been sent to "+email.Email, "")
}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (r *Router) VerifyForgetPasswordCodeHandler(w http.ResponseWriter, req *http.Request) {

	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(data.Email)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}

	if user.Code != data.Code {
		writeErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.Token.Timeout) * time.Minute).Before(time.Now()) {
		writeErrResponse(w, fmt.Errorf("time out"))
		return
	}

	writeMsgResponse(w, "code is verified", map[string]string{"user_id": user.ID.String()})
}

// ChangePasswordHandler changes password of user
func (r *Router) ChangePasswordHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: Rawda: change password for verify
	data := ChangePasswordInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	// hash password
	hashedPassword, err := internal.HashPassword(data.Password)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = r.db.UpdatePassword(data.Email, hashedPassword)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	writeMsgResponse(w, "Password Updated Successfully", "")
}

// UpdateUserHandler updates user's data
func (r *Router) UpdateUserHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("UserID").(string)
	input := UpdateUserInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	var hashedPassword string
	if len(strings.TrimSpace(input.Password)) != 0 {
		// password and confirm password should match
		if input.Password != input.ConfirmPassword {
			writeErrResponse(w, fmt.Errorf("password and confirm password don't match"))
			return
		}

		//validate passwords
		err = validator.ValidatePassword(input.Password)
		if err != nil {
			writeErrResponse(w, fmt.Errorf("error: %v password isn't valid", err))
			return
		}

		// hash password
		hashedPassword, err = internal.HashPassword(input.Password)
		if err != nil {
			writeErrResponse(w, err)
			return
		}
	}

	//TODO: validate ssh key

	userID, err = r.db.UpdateUserByID(userID, input.Name, hashedPassword, input.SSHKey, time.Time{}, 0)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	writeMsgResponse(w, "user updated successfully", map[string]string{"user_id": userID})
}

// GetUserHandler returns user by its idx
func (r *Router) GetUserHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("UserID").(string)
	user, err := r.db.GetUserByID(userID)
	if err != nil {
		writeNotFoundResponse(w, err)
		return
	}
	writeMsgResponse(w, "user exists", map[string]interface{}{"user": user})
}

// ActivateVoucherHandler makes user adds voucher to his account
func (r *Router) ActivateVoucherHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("UserID").(string)

	var input AddVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	voucherQuota, err := r.db.GetVoucher(input.Voucher)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = r.db.AddUserVoucher(userID, input.Voucher)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	err = r.db.UpdateUserQuota(userID, voucherQuota.VMs, voucherQuota.K8s)
	if err != nil {
		writeErrResponse(w, err)
		return
	}

	writeMsgResponse(w, "voucher is applied successfully", "")
}

// func (router *Router) GetAllUsersHandlers(w http.ResponseWriter, r *http.Request) { //TODO: to be removed for testing only
// 	users, err := router.db.GetAllUsers()
// 	if err != nil {
// 		routewriteErrResponse(w, err)
// 	}
// 	userBytes, err := json.Marshal(users)
// 	if err != nil {
// 		routewriteErrResponse(w, err)
// 	}
// 	w.Write(userBytes)
// }
