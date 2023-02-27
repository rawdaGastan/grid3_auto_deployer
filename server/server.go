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
	port string
}

func NewServer(dbFile string) (server Server, err error) {
	content, err := internal.ReadFile("./.env")
	if err != nil {
		return
	}
	m, err := internal.ParseEnv(content)
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

	router := routes.NewRouter(m, db)
	r := mux.NewRouter()
	r.Use(middlewares.Middleware)
	r.HandleFunc("/user/signup", router.SignUpHandler).Methods("POST")
	r.HandleFunc("/user/verify", router.VerifySignUpCodeHandler).Methods("POST")
	r.HandleFunc("/user/signin", router.SignInHandler).Methods("POST")
	r.HandleFunc("/user/home", router.Home).Methods("GET")
	r.HandleFunc("/user/refresh", router.RefreshJWTHandler).Methods("GET")
	r.HandleFunc("/user/logout", router.Logout).Methods("GET")
	r.HandleFunc("/user/forgotPassword", router.ForgotPasswordHandler).Methods("GET")
	r.HandleFunc("/user/forgetpassword/verify", router.VerifyForgetPasswordCodeHandler).Methods("POST")
	r.HandleFunc("/user/changePassword", router.ChangePasswordHandler).Methods("POST")
	r.HandleFunc("/user/update/{id}", router.UpdateUserHandler).Methods("POST")
	r.HandleFunc("/user/get/{id}", router.GetUserHandler).Methods("GET")
	r.HandleFunc("/user/get", router.GetAllUsers).Methods("GET")

	// var port string
	fmt.Print("Enter the port: ")
	fmt.Scan(&server.port)

	err = http.ListenAndServe(server.port, r)
	if err != nil {
		log.Fatalln("There's an error with the server,", err)
	}
	return Server{}, nil
}

func (s *Server) Start() error {

	fmt.Println("Server is listening on " + s.port)
	err := http.ListenAndServe(s.port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server closed: %v", err)
	} else if err != nil {
		return fmt.Errorf("error starting server:  %v", err)
	}
	return nil
}
