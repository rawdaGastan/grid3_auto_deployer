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
	port string //TODO: env variable
}

func NewServer(dbFile string) (server *Server, err error) { //TODO: graceful shutdown
	content, err := internal.ReadFile("./.env") //TODO: pass ./.env and env variable
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

	// port := os.Getenv("PORT")
	// log.Printf("PORT: %s", port)

	router := NewRouter(m, db)
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

	return &Server{port: ":3000"}, nil
}

func (s *Server) Start() (err error) { //TODO:
	srv := &http.Server{
		Addr: ":3000",
	}
	fmt.Println("Server is listening on " + ":3000")
	err = http.ListenAndServe(":3000", nil)

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

	if errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server closed: %v", err)
	} else if err != nil {
		return fmt.Errorf("error starting server:  %v", err)
	}

	return nil
}
