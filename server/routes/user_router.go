// Package routes for API endpoints
package routes

import (
	"encoding/json"
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
func (r *Router) SignUpHandler(w http.ResponseWriter, req *http.Request) {
	var signUp SignUpInput
	err := json.NewDecoder(req.Body).Decode(&signUp)

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read sign up data")
		return
	}

	err = validator.Validate(signUp)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Invalid sign up data")
		return
	}

	// password and confirm password should match
	if signUp.Password != signUp.ConfirmPassword {
		writeErrResponse(req, w, http.StatusBadRequest, "Password and confirm password don't match")
		return
	}

	user, getErr := r.db.GetUserByEmail(signUp.Email)
	// check if user already exists and verified
	if getErr != gorm.ErrRecordNotFound {
		if user.Verified {
			writeErrResponse(req, w, http.StatusBadRequest, "User already exists")
			return
		}
	}

	// send verification code if user is not verified or not exist
	code := internal.GenerateRandomCode()
	subject, body := internal.SignUpMailContent(code, r.config.MailSender.Timeout, signUp.Name)
	err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.SendGridKey, signUp.Email, subject, body)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	hashedPassword, err := internal.HashAndSaltPassword([]byte(signUp.Password))
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
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
		Admin:          internal.Contains(r.config.Admins, signUp.Email),
	}

	// update code if user is not verified but exists
	if getErr != gorm.ErrRecordNotFound {
		if !user.Verified {
			u.ID = user.ID
			u.UpdatedAt = time.Now()
			err = r.db.UpdateUserByID(u)
			if err != nil {
				log.Error().Err(err).Send()
				writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
				return
			}
		}
	}

	// check if user doesn't exist
	if getErr != nil {
		err = r.db.CreateUser(&u)
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}

		// create empty quota
		quota := models.Quota{
			UserID: u.ID.String(),
			Vms:    0,
		}
		err = r.db.CreateQuota(&quota)
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}
	}

	writeMsgResponse(req, w, "Verification code has been sent to "+signUp.Email, map[string]int{"timeout": r.config.MailSender.Timeout})
}

// VerifySignUpCodeHandler gets verification code to create user
func (r *Router) VerifySignUpCodeHandler(w http.ResponseWriter, req *http.Request) {
	var data VerifyCodeInput
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read sign up code data")
		return
	}

	user, err := r.db.GetUserByEmail(data.Email)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if user.Verified {
		writeErrResponse(req, w, http.StatusBadRequest, "Account is already created")
		return
	}

	if user.Code != data.Code {
		writeErrResponse(req, w, http.StatusBadRequest, "Wrong code")
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		writeErrResponse(req, w, http.StatusBadRequest, "Code has expired")
		return
	}
	err = r.db.UpdateVerification(user.ID.String(), true)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	middlewares.UserCreations.WithLabelValues(user.ID.String(), user.Email, user.College, fmt.Sprint(user.TeamSize)).Inc()
	writeMsgResponse(req, w, "Account is created successfully", map[string]string{"user_id": user.ID.String()})
}

// SignInHandler allows user to sign in to the system
func (r *Router) SignInHandler(w http.ResponseWriter, req *http.Request) {
	var input SignInInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read sign in data")
		return
	}

	user, err := r.db.GetUserByEmail(input.Email)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if !user.Verified {
		writeErrResponse(req, w, http.StatusBadRequest, "Email is not verified yet, please check the verification email in your inbox")
		return
	}

	match := internal.VerifyPassword(user.HashedPassword, input.Password)
	if !match {
		writeErrResponse(req, w, http.StatusBadRequest, "Password is not correct")
		return
	}

	token, err := internal.CreateJWT(user.ID.String(), user.Email, r.config.Token.Secret, r.config.Token.Timeout)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "You are signed in successfully", map[string]string{"access_token": token})
}

// RefreshJWTHandler refreshes the user's token
func (r *Router) RefreshJWTHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		writeErrResponse(req, w, http.StatusBadRequest, "Token is required")
		return
	}
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.config.Token.Secret), nil
	})

	// if user doesn't exist
	if _, err := r.db.GetUserByID(claims.UserID); err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}

	// if token didn't expire
	if err == nil && time.Until(claims.ExpiresAt.Time) < time.Duration(r.config.Token.Timeout)*time.Minute && tkn.Valid {
		writeMsgResponse(req, w, "Access Token still valid", map[string]string{"access_token": reqToken, "refresh_token": reqToken})
		return
	}

	expirationTime := time.Now().Add(time.Duration(r.config.Token.Timeout) * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newToken, err := token.SignedString([]byte(r.config.Token.Secret))
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	writeMsgResponse(req, w, "Token is refreshed successfully", map[string]string{"access_token": reqToken, "refresh_token": newToken})
}

// ForgotPasswordHandler sends user verification code
func (r *Router) ForgotPasswordHandler(w http.ResponseWriter, req *http.Request) {

	var email EmailInput
	err := json.NewDecoder(req.Body).Decode(&email)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read email data")
		return
	}

	user, err := r.db.GetUserByEmail(email.Email)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	// send verification code
	code := internal.GenerateRandomCode()
	subject, body := internal.ResetPasswordMailContent(code, r.config.MailSender.Timeout, user.Name)
	err = internal.SendMail(r.config.MailSender.Email, r.config.MailSender.SendGridKey, email.Email, subject, body)

	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.db.UpdateUserByID(
		models.User{
			ID:        user.ID,
			UpdatedAt: time.Now(),
			Code:      code,
		},
	)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	writeMsgResponse(req, w, "Verification code has been sent to "+email.Email, map[string]int{"timeout": r.config.MailSender.Timeout})
}

// VerifyForgetPasswordCodeHandler verifies code sent to user when forgetting password
func (r *Router) VerifyForgetPasswordCodeHandler(w http.ResponseWriter, req *http.Request) {
	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read password code")
		return
	}

	user, err := r.db.GetUserByEmail(data.Email)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if user.Code != data.Code {
		writeErrResponse(req, w, http.StatusBadRequest, "Wrong code")
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.MailSender.Timeout) * time.Second).Before(time.Now()) {
		writeErrResponse(req, w, http.StatusBadRequest, "Code has expired")
		return
	}

	// token
	token, err := internal.CreateJWT(user.ID.String(), user.Email, r.config.Token.Secret, r.config.Token.Timeout)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Code is verified", map[string]string{"access_token": token})
}

// ChangePasswordHandler changes password of user
func (r *Router) ChangePasswordHandler(w http.ResponseWriter, req *http.Request) {
	var data ChangePasswordInput
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read password data")
		return
	}

	err = validator.Validate(data)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Invalid password data")
		return
	}

	if data.ConfirmPassword != data.Password {
		writeErrResponse(req, w, http.StatusBadRequest, "Password does not match confirm password")
		return
	}

	hashedPassword, err := internal.HashAndSaltPassword([]byte(data.Password))
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.db.UpdatePassword(data.Email, hashedPassword)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "Password is updated successfully", "")
}

// UpdateUserHandler updates user's data
func (r *Router) UpdateUserHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	input := UpdateUserInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read user data")
		return
	}
	updates := 0

	var hashedPassword []byte
	if len(strings.TrimSpace(input.Password)) != 0 {
		updates++
		// password and confirm password should match
		if input.Password != input.ConfirmPassword {
			writeErrResponse(req, w, http.StatusBadRequest, "Password and confirm password don't match")
			return
		}

		err = validators.ValidatePass(input.Password)
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusBadRequest, "Invalid password")
			return
		}

		// hash password
		hashedPassword, err = internal.HashAndSaltPassword([]byte(input.Password))
		if err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
			return
		}
	}

	if len(strings.TrimSpace(input.SSHKey)) != 0 {
		updates++
		if err := validators.ValidateSSH(input.SSHKey); err != nil {
			log.Error().Err(err).Send()
			writeErrResponse(req, w, http.StatusBadRequest, "Invalid sshKey")
			return
		}
	}

	if len(strings.TrimSpace(input.Name)) != 0 {
		updates++
	}

	if updates == 0 {
		writeMsgResponse(req, w, "Nothing to update", "")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	err = r.db.UpdateUserByID(
		models.User{
			ID:             userUUID,
			Name:           input.Name,
			HashedPassword: hashedPassword,
			SSHKey:         input.SSHKey,
			UpdatedAt:      time.Now(),
		},
	)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	writeMsgResponse(req, w, "User is updated successfully", map[string]string{"user_id": userID})
}

// GetUserHandler returns user by its idx
func (r *Router) GetUserHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	user, err := r.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User is not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	writeMsgResponse(req, w, "User exists", map[string]interface{}{"user": user})
}

// ApplyForVoucherHandler makes user apply for voucher that would be accepted by admin
func (r *Router) ApplyForVoucherHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)
	userVoucher, err := r.db.GetNotUsedVoucherByUserID(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "Voucher is not found")
		return
	}
	if userVoucher.Voucher != "" && !userVoucher.Approved && !userVoucher.Rejected {
		writeErrResponse(req, w, http.StatusBadRequest, "You have already a voucher request, please wait for the confirmation mail")
		return
	}

	var input ApplyForVoucherInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read voucher data")
		return
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

	err = r.db.CreateVoucher(&voucher)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	middlewares.VoucherApplied.WithLabelValues(userID, voucher.Voucher, fmt.Sprint(voucher.VMs), fmt.Sprint(voucher.PublicIPs)).Inc()
	writeMsgResponse(req, w, "Voucher request is being reviewed, you'll receive a confirmation mail soon", "")
}

// ActivateVoucherHandler makes user adds voucher to his account
func (r *Router) ActivateVoucherHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input AddVoucherInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusBadRequest, "Failed to read voucher data")
		return
	}

	oldQuota, err := r.db.GetUserQuota(userID)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User quota not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	voucherQuota, err := r.db.GetVoucher(input.Voucher)
	if err == gorm.ErrRecordNotFound {
		writeErrResponse(req, w, http.StatusNotFound, "User voucher not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	if voucherQuota.Rejected {
		writeErrResponse(req, w, http.StatusBadRequest, "Voucher is rejected")
		return
	}

	if !voucherQuota.Approved {
		writeErrResponse(req, w, http.StatusBadRequest, "Voucher is not approved yet")
		return
	}

	if voucherQuota.Used {
		writeErrResponse(req, w, http.StatusBadRequest, "Voucher is already used")
		return
	}

	err = r.db.DeactivateVoucher(userID, input.Voucher)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	err = r.db.UpdateUserQuota(userID, oldQuota.Vms+voucherQuota.VMs, oldQuota.PublicIPs+voucherQuota.PublicIPs)
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(req, w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	middlewares.VoucherActivated.WithLabelValues(userID, voucherQuota.Voucher, fmt.Sprint(voucherQuota.VMs), fmt.Sprint(voucherQuota.PublicIPs)).Inc()
	writeMsgResponse(req, w, "Voucher is applied successfully", "")
}
