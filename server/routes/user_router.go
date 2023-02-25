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

type AuthDataInput struct {
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

func (router *Router) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var signUp SignUpInput
	err := json.NewDecoder(r.Body).Decode(&signUp)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	// validate mail
	valid := validator.ValidateMail(signUp.Email)
	if !valid {
		router.WriteErrResponse(w, fmt.Errorf("email isn't valid %v", err))
		return
	}

	//validate password
	err = validator.ValidatePassword(signUp.Password)
	if err != nil {
		router.WriteErrResponse(w, fmt.Errorf("error: %v password isn't valid", err))
		return
	}

	// password and confirm password should match
	if signUp.Password != signUp.ConfirmPassword {
		router.WriteErrResponse(w, fmt.Errorf("password and confirm password don't match"))
		return
	}

	u := models.User{
		Name:     signUp.Name,
		Email:    signUp.Email,
		Password: signUp.Password,
	}

	// check if user already exists
	_, err = router.db.GetUserByEmail(u.Email)
	if err == nil {
		router.WriteMsgResponse(w, "user already exists", u.Email)
	}

	// hash password
	hashedPassword, err := internal.HashPassword(u.Password)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	u.Password = hashedPassword

	// send verification code
	code, err := internal.SendMail(u.Email)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	u.Code = code
	fmt.Printf("code: %v\n", code)
	msg := "Verfification Code has been sent to " + u.Email
	router.db.SetCache(u.Email, u)
	router.WriteMsgResponse(w, msg, "") //TODO:

}

// get verification code to create user
func (router *Router) VerifySignUpCodeHandler(w http.ResponseWriter, r *http.Request) {
	data := AuthDataInput{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	cachedUser, err := router.db.GetCache(data.Email)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	if cachedUser.Code != data.Code {
		router.WriteErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	u, err := router.db.CreateUser(&cachedUser)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	fmt.Printf("u: %v\n", u)
	router.WriteMsgResponse(w, "Account Created Successfully", u.Email)
}

func (router *Router) SignInHandler(w http.ResponseWriter, r *http.Request) {
	u := models.User{}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	user, err := router.db.GetUserByEmail(u.Email)
	if err != nil {
		router.WriteErrResponse(w, err)
	}

	err = internal.VerifyPassword(user.Password, u.Password)
	if err != nil {
		router.WriteErrResponse(w, fmt.Errorf("error %v, Password is not correct", err))
		return
	}

	token, err := internal.CreateJWT(&u, router.secret)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	json.NewEncoder(w).Encode("token :" + token)
}

func (router *Router) Home(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(router.secret), nil
	})
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	if !token.Valid {
		router.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	router.WriteMsgResponse(w, "Welcome Home "+claims.Email, "") //TODO: ID empty ??
}

func (router *Router) RefreshJWTHandler(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(router.secret), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			router.WriteErrResponse(w, err)
			return
		}
		router.WriteErrResponse(w, err)
		return
	}
	if !tkn.Valid {
		router.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		return
	}

	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		router.WriteErrResponse(w, fmt.Errorf("token is expired"))
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newToken, err := token.SignedString([]byte(router.secret))
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	router.WriteMsgResponse(w, "Old Token", reqToken)
	router.WriteMsgResponse(w, "New Token", newToken)
}

func (router *Router) Logout(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(router.secret), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			router.WriteErrResponse(w, err)
			return
		}
		router.WriteErrResponse(w, err)
		return
	}
	if !tkn.Valid {
		router.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		return
	}

	expirationTime := time.Now()
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	router.WriteMsgResponse(w, "Logged out successfully", "")

}

func (router *Router) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var email EmailInput
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	// send verification code
	code, err := internal.SendMail(email.Email)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	fmt.Printf("code: %v\n", code)

	msg := "Verfification Code has been sent to " + email.Email

	router.db.SetCache(email.Email, code)
	router.WriteMsgResponse(w, msg, "")
}

func (router *Router) VerifyForgetPasswordCodeHandler(w http.ResponseWriter, r *http.Request) { //TODO: Error
	data := AuthDataInput{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	cachedUser, err := router.db.GetCache(data.Email)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	if cachedUser.Code != data.Code {
		router.WriteErrResponse(w, fmt.Errorf("wrong code"))
		return
	}

	msg := "Code Verified"
	router.WriteMsgResponse(w, msg, "")
	//TODO: should output msg
}

func (router *Router) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	data := ChangePasswordInput{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	err = router.db.UpdatePassword(data.Email, data.Password)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	//TODO: hash password before changing it
	//TODO: should output msg
}

func (router *Router) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: will the user be able to update its email or not ??
	id := mux.Vars(r)["id"]

	input := UpdateUserInput{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	//validate password
	err = validator.ValidatePassword(input.Password)
	if err != nil {
		router.WriteErrResponse(w, fmt.Errorf("error: %v password isn't valid", err))
		return
	}

	updatedUser, err := router.db.UpdateUserById(id, input.Name, input.Password, input.Voucher)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	userBytes, err := json.Marshal(updatedUser)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(userBytes)
}

func (router *Router) GetUserHandler(w http.ResponseWriter, r *http.Request) { //TODO: error
	id := mux.Vars(r)["id"]
	user, err := router.db.GetUserById(id)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	w.Write(userBytes)
}
