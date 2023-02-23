package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/grid3_auto_deployer/internal"
	"github.com/rawdaGastan/grid3_auto_deployer/middlewares"
	"github.com/rawdaGastan/grid3_auto_deployer/models"
	"github.com/rawdaGastan/grid3_auto_deployer/routes"
)

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

	router := routes.NewRouter(secret, db)
	r := mux.NewRouter()
	r.Use(middlewares.Middleware)
	r.HandleFunc("/signup", router.SignUpHandler).Methods("POST")
	r.HandleFunc("/verify", router.VerifyUser).Methods("POST")
	r.HandleFunc("/signin", router.SignInHandler).Methods("GET")
	r.HandleFunc("/home", router.Home).Methods("GET")
	r.HandleFunc("/refresh", router.RefreshJWT).Methods("GET")
	r.HandleFunc("/logout", router.Logout).Methods("GET")
	r.HandleFunc("/forgotPassword", router.ForgotPasswordHandler).Methods("GET")
	r.HandleFunc("/verifycode", router.VerifyCode).Methods("POST")
	r.HandleFunc("/changePassword", router.ChangePassword).Methods("POST")
	r.HandleFunc("/updateAccount", router.UpdateAccount).Methods("POST")
	r.HandleFunc("/getUser", router.GetUser).Methods("GET")

	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatalln("There's an error with the server,", err)
	}
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
