package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/caitlin615/nist-password-validator/password"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rawdaGastan/grid3_auto_deployer/internal"
	"github.com/rawdaGastan/grid3_auto_deployer/models"
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

func (router *Router) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var signUp SignUpInput
	err := json.NewDecoder(r.Body).Decode(&signUp)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	// validate mail
	bool := internal.ValidateMail(signUp.Email)
	if !bool {
		router.WriteErrResponse(w, fmt.Errorf("email isn't valid %v", err))
		return
	}

	u := models.User{
		Name:     signUp.Name,
		Email:    signUp.Email,
		Password: signUp.Password,
	}

	// check if user already exists
	_, err = router.db.GetUserByEmail(u.Email, router.secret)
	if err != nil {
		router.WriteErrResponse(w, fmt.Errorf("user already exists"))
	}

	// password should be ACII , min 5 , max 10
	validator := password.NewValidator(true, 5, 10)
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
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		router.WriteErrResponse(w, fmt.Errorf("error: %v", err))
	}

	router.db.SetCache(u.Email, u)
	router.WriteMsgResponse(w, msgBytes)

}

// get verification code to create user
func (router *Router) VerifyUser(w http.ResponseWriter, r *http.Request) {
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

	userBytes, err := json.Marshal(u)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	router.WriteMsgResponse(w, userBytes)
}

func (router *Router) SignInHandler(w http.ResponseWriter, r *http.Request) {
	u := models.User{}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	user, err := router.db.GetUserByEmail(u.Email, router.secret)
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

	router.WriteMsgResponse(w, "Token :"+token)
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
		if err == jwt.ErrSignatureInvalid {
			router.WriteErrResponse(w, err)
			return
		}
		router.WriteErrResponse(w, err)
		return
	}
	if !token.Valid {
		router.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	router.WriteMsgResponse(w, "Welcome Home "+claims.Email)
}

func (router *Router) RefreshJWT(w http.ResponseWriter, r *http.Request) {
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
	router.WriteMsgResponse(w, "Old Token: "+reqToken+"/n New Token: "+newToken)
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
		router.WriteErrResponse(w, fmt.Errorf("token is already invalid"))
		return
	}

	expirationTime := time.Now()
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	router.WriteMsgResponse(w, "Logged out successfully")

}

func (router *Router) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var email string
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	// send verification code
	code, err := internal.SendMail(email)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}
	fmt.Printf("code: %v\n", code)

	msg := "Verfification Code has been sent to " + email
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	router.db.SetCache(email, code)
	router.WriteMsgResponse(w, msgBytes)
}

func (router *Router) VerifyCode(w http.ResponseWriter, r *http.Request) {
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
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	router.WriteMsgResponse(w, msgBytes)
}

func (router *Router) ChangePassword(w http.ResponseWriter, r *http.Request) {
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
}

func (router *Router) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	u := models.User{}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	updatedUser, err := router.db.UpdateData(&u)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	userBytes, err := json.Marshal(updatedUser)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	w.WriteHeader(http.StatusCreated)
	router.WriteMsgResponse(w, userBytes)
}

func (router *Router) GetUser(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
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
	if !token.Valid {
		router.WriteErrResponse(w, fmt.Errorf("token is invalid"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, err := router.db.GetUserByEmail(claims.Email, router.secret)
	if err != nil {
		router.WriteErrResponse(w, err)
	}
	router.WriteMsgResponse(w, user)
}
