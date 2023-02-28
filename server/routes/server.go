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
	"github.com/rawdaGastan/grid3_auto_deployer/internal"
	"github.com/rawdaGastan/grid3_auto_deployer/middlewares"
	"github.com/rawdaGastan/grid3_auto_deployer/models"
)

type Server struct {
	port string
}

func NewServer(file string) (server *Server, err error) {

	data, err := internal.ReadConfFile("./config.json")
	if err != nil {
		return
	}
	configuration, err := internal.ParseConf(data)
	if err != nil {
		return
	}

	db := models.NewDB()
	err = db.Connect(file)
	if err != nil {
		return
	}
	err = db.Migrate()
	if err != nil {
		return
	}

	router := NewRouter(*configuration, db)
	r := mux.NewRouter()
	r.Use(middlewares.LoggingMW)
	r.HandleFunc("/user/signup", router.SignUpHandler).Methods("POST")
	r.HandleFunc("/user/verify", router.VerifySignUpCodeHandler).Methods("POST")
	r.HandleFunc("/user/signin", router.SignInHandler).Methods("POST")
	// r.HandleFunc("/user/home", router.Home).Methods("GET")
	r.HandleFunc("/user/refresh", router.RefreshJWTHandler).Methods("GET")
	r.HandleFunc("/user/logout", router.Logout).Methods("GET")
	r.HandleFunc("/user/forgotPassword", router.ForgotPasswordHandler).Methods("GET")
	r.HandleFunc("/user/forgetpassword/verify", router.VerifyForgetPasswordCodeHandler).Methods("POST")
	r.HandleFunc("/user/changePassword", router.ChangePasswordHandler).Methods("POST")
	r.HandleFunc("/user/update/{id}", router.UpdateUserHandler).Methods("POST")
	r.HandleFunc("/user/get/{id}", router.GetUserHandler).Methods("GET")
	r.HandleFunc("/user/get", router.GetAllUsersHandlres).Methods("GET") //for testing only
	r.HandleFunc("/user/addvoucher/{id}", router.AddVoucherHandler).Methods("POST")
	http.Handle("/", r)

	return &Server{port: configuration.Server.Port}, nil
}

func (s *Server) Start() (err error) {

	fmt.Println("Server is listening on " + s.port)

	srv := &http.Server{
		Addr: ":3000",
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
