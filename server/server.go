package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rawdaGastan/grid3_auto_deployer/models"
	"github.com/rawdaGastan/grid3_auto_deployer/routes"
)

//TODO: add middleware

type Server struct {
}

func NewServer(dbFile string) (server Server, err error) {
	db := models.NewDB()
	err = db.Connect(dbFile)
	if err != nil {
		return
	}
	err = db.Migrate()
	if err != nil {
		return
	}
	router := routes.NewRouter(db) //TODO: Add rest of routers
	http.HandleFunc("/signup", router.SignUpHandler)
	http.HandleFunc("/signin", router.SignInHandler)
	http.HandleFunc("/verify", router.VerifyUser)
	http.HandleFunc("/forgotPassword", router.ForgotPasswordHandler)
	http.HandleFunc("/changePassword",router.ChangePassword)

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
