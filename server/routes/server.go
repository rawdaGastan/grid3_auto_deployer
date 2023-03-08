// Package routes for API endpoints
package routes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rawdaGastan/cloud4students/models"
)

// Server struct holds port of server
type Server struct {
	port string
}

// NewServer create new server with all configurations
func NewServer(file string) (server *Server, err error) {

	data, err := internal.ReadConfFile(file)
	if err != nil {
		return
	}
	configuration, err := internal.ParseConf(data)
	if err != nil {
		return
	}

	db := models.NewDB()
	err = db.Connect(configuration.Database.File)
	if err != nil {
		return
	}
	err = db.Migrate()
	if err != nil {
		return
	}
	//TODO: add version

	router := NewRouter(*configuration, db)
	r := mux.NewRouter()
	signUp := r.HandleFunc("/user/signup", router.SignUpHandler).Methods("POST", "OPTIONS")
	signUpVerify := r.HandleFunc("/user/signup/verify_email", router.VerifySignUpCodeHandler).Methods("POST", "OPTIONS")
	signIn := r.HandleFunc("/user/signin", router.SignInHandler).Methods("POST", "OPTIONS")
	refreshToken := r.HandleFunc("/user/refresh_token", router.RefreshJWTHandler).Methods("POST", "OPTIONS")
	forgetPass := r.HandleFunc("/user/forgot_password", router.ForgotPasswordHandler).Methods("POST", "OPTIONS")
	forgetPassVerify := r.HandleFunc("/user/forget_password/verify_email", router.VerifyForgetPasswordCodeHandler).Methods("POST", "OPTIONS")
	changePassword := r.HandleFunc("/user/change_password", router.ChangePasswordHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc("/user", router.UpdateUserHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc("/user", router.GetUserHandler).Methods("GET", "OPTIONS")
	//r.HandleFunc("/user", router.GetAllUsersHandlers).Methods("GET", "OPTIONS") //TODO:for testing only
	r.HandleFunc("/user/activate_voucher", router.ActivateVoucherHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc("/user/signout", router.SignOut).Methods("POST", "OPTIONS")

	// ADMIN ACCESS
	r.HandleFunc("/voucher/generate", router.GenerateVoucherHandler).Methods("POST")

	r.Use(middlewares.LoggingMW)
	r.Use(middlewares.EnableCors)
	excludedRoutes := []*mux.Route{signUp, signUpVerify, signIn, refreshToken, forgetPass, forgetPassVerify, changePassword}
	r.Use(middlewares.Authorization(excludedRoutes, configuration.Token.Secret, configuration.Token.Timeout))
	http.Handle("/", r)

	return &Server{port: configuration.Server.Port}, nil
}

// Start starts the server
func (s *Server) Start() (err error) {

	fmt.Println("Server is listening on " + s.port)

	srv := &http.Server{
		Addr: s.port,
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")

	return nil
}
