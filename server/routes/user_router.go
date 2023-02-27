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

type SignUpInput struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" gorm:"unique" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type VerifyCodeInput struct {
	Email string `json:"email"`
	Code  int    `json:"code"`
}

type ChangePasswordInput struct {
	Email           string `json:"email"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type UpdateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password"`
	Voucher  string `json:"voucher"`
}

type EmailInput struct {
	Email string `json:"email" binding:"required"`
}

type AddVoucherInput struct {
	Voucher string `json:"voucher" binding:"required"`
}

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
	}

	// check if user already exists
	_, err = r.db.GetUserByEmail(u.Email)
	if err == nil {
		r.WriteMsgResponse(w, "user already exists", u.Email)
	}

	// hash password
	hashedPassword, err := internal.HashPassword(u.HashedPassword)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	u.HashedPassword = hashedPassword

	// send verification code
	code, err := internal.SendMail(r.mailSender, r.password, u.Email, "Cloud4Students", "")
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	u.Code = code
	fmt.Printf("code: %v\n", code)
	msg := "Verfification Code has been sent to " + u.Email
	r.db.SetCache(u.Email, u)
	r.WriteMsgResponse(w, msg, "")
}

// get verification code to create user
func (r *Router) VerifySignUpCodeHandler(w http.ResponseWriter, req *http.Request) {
	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	cachedUser, err := r.db.GetCache(data.Email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	if cachedUser.Code != data.Code {
		r.WriteErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	u, err := r.db.CreateUser(&cachedUser)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "Account Created Successfully", u.Email)
}

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
	}

	err = internal.VerifyPassword(user.HashedPassword, u.HashedPassword)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		r.WriteErrResponse(w, fmt.Errorf("error %v, Password is not correct", err))
		return
	}

	token, err := internal.CreateJWT(&u, r.secret)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "token", token)
}

// func (r *Router) Home(w http.ResponseWriter, req *http.Request) {
// 	reqToken := req.Header.Get("Authorization")
// 	splitToken := strings.Split(reqToken, "Bearer ")
// 	reqToken = splitToken[1]

// 	claims := &models.Claims{}
// 	token, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(r.secret), nil
// 	})
// 	if err != nil {
// 		r.WriteErrResponse(w, err)
// 		return
// 	}
// 	if !token.Valid {
// 		r.WriteErrResponse(w, fmt.Errorf("token is invalid"))
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}
// 	r.WriteMsgResponse(w, "Welcome Home"+claims.Email, "")
// }

func (r *Router) RefreshJWTHandler(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.secret), nil
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
	newToken, err := token.SignedString([]byte(r.secret))
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	r.WriteMsgResponse(w, "Old Token", reqToken)
	r.WriteMsgResponse(w, "New Token", newToken)
}

func (r *Router) Logout(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.secret), nil
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

func (r *Router) ForgotPasswordHandler(w http.ResponseWriter, req *http.Request) {
	var email EmailInput
	err := json.NewDecoder(req.Body).Decode(&email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	// send verification code
	code, err := internal.SendMail(r.mailSender, r.password, email.Email, "Cloud4Students", "")
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}
	fmt.Printf("code: %v\n", code)

	msg := "Verfification Code has been sent to " + email.Email

	r.db.SetCache(email.Email, code)
	r.WriteMsgResponse(w, msg, "")
}

func (r *Router) VerifyForgetPasswordCodeHandler(w http.ResponseWriter, req *http.Request) { //TODO: Error
	data := VerifyCodeInput{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	cachedUser, err := r.db.GetCache(data.Email)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	if cachedUser.Code != data.Code {
		r.WriteErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	msg := "Code Verified"
	r.WriteMsgResponse(w, msg, "")
}

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

	r.WriteMsgResponse(w, "Password Changed", "")
}

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

	updatedUser, err := r.db.UpdateUserById(id, input.Name, input.Password, input.Voucher)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	r.WriteMsgResponse(w, "user updated successfully", updatedUser)
}

func (r *Router) GetUserHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	user, err := r.db.GetUserById(id)
	if err != nil {
		r.WriteErrResponse(w, err)
	}
	r.WriteMsgResponse(w, "", user)

}

func (router *Router) GetAllUsersHandlres(w http.ResponseWriter, r *http.Request) {
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

func (r *Router) AddVoucherHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	var voucher AddVoucherInput
	err := json.NewDecoder(req.Body).Decode(&voucher)
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	user := r.db.AddVoucher(id, voucher.Voucher)
	r.WriteMsgResponse(w, "Voucher Applied Successfuly", user)
}
