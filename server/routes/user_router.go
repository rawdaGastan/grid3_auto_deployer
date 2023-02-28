package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/rawdaGastan/grid3_auto_deployer/internal"
	"github.com/rawdaGastan/grid3_auto_deployer/models"
	"github.com/rawdaGastan/grid3_auto_deployer/validator"
)

// SignUpInput struct for data needed when user creates account
type SignUpInput struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" gorm:"unique" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// VerifyCodeInput struct takes verification code from user
type VerifyCodeInput struct {
	Email     string    `json:"email"`
	Code      int       `json:"code"`
	CreatedAt time.Time `json:"timestamp"`
}

// ChangePasswordInput struct for user to change password
type ChangePasswordInput struct {
	Email           string `json:"email"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// UpdateUserInput struct for user to updates his data
type UpdateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password"`
	Voucher  string `json:"voucher"`
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

	u := models.User{
		Name:           signUp.Name,
		Email:          signUp.Email,
		HashedPassword: signUp.Password,
		Verified:       false,
	}

	// check if user already exists
	_, err = r.db.GetUserByEmail(u.Email)
	if err == nil {
		r.WriteMsgResponse(w, "user already exists", u.Email)
		return
	}

	// hash password
	hashedPassword, err := internal.HashPassword(signUp.Password)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	u.HashedPassword = hashedPassword

	// send verification code
	code, err := internal.SendMail(r.config.MailSender.Email, r.config.MailSender.Password, u.Email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	u.Code = code
	fmt.Printf("code: %v\n", code) //TODO: to be removed
	msg := "Verification Code has been sent to " + u.Email
	err = r.db.CreateUser(&u)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, msg, "")
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

	if user.Code != data.Code {
		r.WriteErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	if user.UpdatedAt.Add(time.Duration(r.config.Token.Timeout) * time.Minute).Before(time.Now()) {
		r.WriteErrResponse(w, fmt.Errorf("time out"))
		return
	}
	err = r.db.UpdateVerification(user.ID, true)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "Account Created Successfully", map[string]string{"user_id": user.ID})
}

// SignInHandler allows user to sign in to the system
func (r *Router) SignInHandler(w http.ResponseWriter, req *http.Request) {
	u := models.User{}
	err := json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	user, err := r.db.GetUserByEmail(u.Email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	if !user.Verified {
		r.WriteErrResponse(w, fmt.Errorf("user not verified yet"))
		return
	}

	match := internal.VerifyPassword(user.HashedPassword, u.HashedPassword)
	if match {
		r.WriteErrResponse(w, fmt.Errorf("password is not correct"))
		return
	}

	token, err := internal.CreateJWT(&u, r.config.Token.Secret, r.config.Token.Timeout)
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
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.config.Token.Secret), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			r.WriteErrResponse(w, err)
			return
		}
		r.WriteErrResponse(w, err)
		return
	}
	if !tkn.Valid {
		r.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		return
	}

	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		r.WriteErrResponse(w, fmt.Errorf("token is expired"))
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newToken, err := token.SignedString([]byte(r.config.Token.Secret))
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "token is refreshed successfully", map[string]string{"access_token": reqToken, "refresh_token": newToken})
}

// Logout allows user to logout from the system by expiring his token
func (r *Router) Logout(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.config.Token.Secret), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			r.WriteErrResponse(w, err)
			return
		}
		r.WriteErrResponse(w, err)
		return
	}
	if !tkn.Valid {
		r.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		return
	}

	expirationTime := time.Now()
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	r.WriteMsgResponse(w, "Logged out successfully", "")

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
	code, err := internal.SendMail(r.config.MailSender.Email, r.config.MailSender.Password, email.Email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	fmt.Printf("code: %v\n", code) //TODO: to be removed

	msg := "Verification Code has been sent to " + email.Email
	_, err = r.db.UpdateUserByID(user.ID, "", "", "", time.Now(), code)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, msg, "")
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

	if user.UpdatedAt.Add(time.Duration(r.config.Token.Timeout) * time.Minute).Before(data.CreatedAt) {
		r.WriteErrResponse(w, fmt.Errorf("time out"))
		return
	}

	msg := "Code Verified"
	r.WriteMsgResponse(w, msg, map[string]string{"user_id": user.ID})
}

// ChangePasswordHandler changes password of user
func (r *Router) ChangePasswordHandler(w http.ResponseWriter, req *http.Request) {
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
	input := UpdateUserInput{}
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	//validate password
	err = validator.ValidatePassword(input.Password)
	if err != nil {
		r.WriteErrResponse(w, fmt.Errorf("error: %v password isn't valid", err))
		return
	}

	userID, err := r.db.UpdateUserByID(id, input.Name, input.Password, input.Voucher, time.Time{}, 0)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	r.WriteMsgResponse(w, "user updated successfully", map[string]string{"user_id": userID})
}

// GetUserHandler returns user by its idx
func (r *Router) GetUserHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	user, err := r.db.GetUserByID(id)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "user exists", user)

}

// AddVoucherHandler makes user adds voucher to his account
func (r *Router) AddVoucherHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	var voucher AddVoucherInput
	err := json.NewDecoder(req.Body).Decode(&voucher)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	err = r.db.AddVoucher(id, voucher.Voucher)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "Voucher Applied Successfully", "")
}

// func (router *Router) GetAllUsersHandlers(w http.ResponseWriter, r *http.Request) { //TODO: to be removed for testing only
// 	users, err := router.db.GetAllUsers()
// 	if err != nil {
// 		router.WriteErrResponse(w, err)
// 	}
// 	userBytes, err := json.Marshal(users)
// 	if err != nil {
// 		router.WriteErrResponse(w, err)
// 	}
// 	w.Write(userBytes)
// }
