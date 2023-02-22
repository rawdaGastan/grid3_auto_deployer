package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rawdaGastan/grid3_auto_deployer/internal"
	"github.com/rawdaGastan/grid3_auto_deployer/models"
	"github.com/rawdaGastan/grid3_auto_deployer/routes"
)

//TODO: add middleware

type Server struct {
}

func NewServer(dbFile string) (server Server, err error) {
	content, err := internal.ReadFile("./.env")
	if err != nil {
		return
	}

	secret, err := internal.ParseEnv(content)
	if err != nil {
		return
	}

	db := models.NewDB()
	err = db.Connect(dbFile)
	if err != nil {
		return
	}
	err = db.Migrate()
	if err != nil {
		return
	}
	router := routes.NewRouter(secret, db) //TODO: Add rest of routers
	http.HandleFunc("/signup", router.SignUpHandler)
	http.HandleFunc("/verify", router.VerifyUser)
	http.HandleFunc("/signin", router.SignInHandler)
	http.HandleFunc("/home", router.Home)
	http.HandleFunc("/refresh", router.RefreshJWT)
	http.HandleFunc("/logout", router.Logout)
	http.HandleFunc("/forgotPassword", router.ForgotPasswordHandler)
	http.HandleFunc("/changePassword", router.ChangePassword)
	http.HandleFunc("/updateUser", router.UpdateAccount)
	

	return Server{}, nil
}

func (s *Server) Start() error {

	fmt.Println("Server is listening on 3000")
	err := http.ListenAndServe(":3000", nil)

	if errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server closed: %v", err)
	} else if err != nil {
		return fmt.Errorf("error starting server:  %v", err)
	}
	return nil
}
