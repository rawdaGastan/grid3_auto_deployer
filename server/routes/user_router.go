package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rawdaGastan/grid3_auto_deployer/models"
)

type SignUpInput struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" gorm:"unique" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type AuthData struct {
	Email string `json:"email"`
	Code  int    `json:"code"`
}

type ChangePassword struct {
	Email           string `json:"email"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

func (router *Router) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	signUp := SignUpInput{}
	err := json.NewDecoder(r.Body).Decode(&signUp)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	if signUp.Password != signUp.ConfirmPassword {
		http.Error(w, "Password and Confirm Password dosen't match", http.StatusInternalServerError)
		return
	}

	u := models.User{
		Name:     signUp.Name,
		Email:    signUp.Email,
		Password: signUp.Password,
	}

	hashedPassword, err := HashPassword(u.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	u.Password = hashedPassword

	// send verification code
	code, err := SendMail(u.Email)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}
	u.Code = code
	fmt.Printf("code: %v\n", code)
	msg := "Verfification Code has been sent to " + u.Email
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}

	router.db.SetCache(u.Email, u)

	w.WriteHeader(http.StatusCreated)
	w.Write(msgBytes)
}

func (router *Router) VerifyUser(w http.ResponseWriter, r *http.Request) {
	data := AuthData{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	cachedUser, err := router.db.GetCache(data.Email)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	if cachedUser.Code != data.Code {
		errJSON, _ := json.Marshal("wrong code")
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	u, err := router.db.SignUp(&cachedUser)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	userBytes, err := json.Marshal(u)
	if err != nil {
		fmt.Print(err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(userBytes)

}

// TODO: retrun token str in json
func (router *Router) SignInHandler(w http.ResponseWriter, r *http.Request) {
	db := models.NewDB()
	u := models.User{}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}
	cached, err := db.SignIn(&u)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusFound)
	w.Write(cached)
}

func (router *Router) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var email string
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	// send verification code
	code, err := SendMail(email)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}
	fmt.Printf("code: %v\n", code)
	msg := "Verfification Code has been sent to " + email
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}

	router.db.SetCache(email, code)
	w.WriteHeader(http.StatusCreated)
	w.Write(msgBytes)
}

func (router *Router) VerifyCode(w http.ResponseWriter, r *http.Request) {
	data := AuthData{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	cachedUser, err := router.db.GetCache(data.Email)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	if cachedUser.Code != data.Code {
		errJSON, _ := json.Marshal("wrong code")
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	msg := "Code Verified"
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(msgBytes)
}

func (router *Router) ChangePassword(w http.ResponseWriter, r *http.Request) {
	data := ChangePassword{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	err = router.db.ChangePassword(data.Email, data.Password)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}
}

func (router *Router) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	u := models.User{}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	updatedUser, err := router.db.UpdateData(&u)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	userBytes, err := json.Marshal(updatedUser)
	if err != nil {
		fmt.Print(err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(userBytes)
}
