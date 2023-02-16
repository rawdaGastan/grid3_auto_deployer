package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type App struct {
	db *gorm.DB
}

func newApp(db *gorm.DB) *App {
	return &App{db}
}

type ErrorMsg struct {
	Message string `json:"message"`
}

type SignUpInput struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" gorm:"unique" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashedPassword string, candidatePassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidatePassword))
}

func (app *App) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	u := &models.User{}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	userCreated, err := u.CreateUser()
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}
	userBytes, err := json.Marshal(userCreated)
	if err != nil {
		fmt.Print(err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(userBytes)
}

func (app *App) SignInHandler(w http.ResponseWriter, r *http.Request) { //TODO:
	u := &models.User{}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}

	data, err := u.SignIn(u.Email, u.Password)
	if err != nil {
		errJSON, _ := json.Marshal(ErrorMsg{Message: err.Error()})
		http.Error(w, string(errJSON), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusFound)
	w.Write(data)
}
