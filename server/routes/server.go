// Package routes for API endpoints
package routes

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/grid3-go/deployer"
)

// Server struct holds port of server
type Server struct {
	host string
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

	tfPluginClient, err := deployer.NewTFPluginClient(configuration.Account.Mnemonics, "sr25519", configuration.Account.Network, "", "", "", 0, true, false)
	if err != nil {
		return
	}

	version := "/" + configuration.Version

	router, err := NewRouter(*configuration, db, tfPluginClient)
	if err != nil {
		return
	}
	r := mux.NewRouter()
	signUp := r.HandleFunc(version+"/user/signup", router.SignUpHandler).Methods("POST", "OPTIONS")
	signUpVerify := r.HandleFunc(version+"/user/signup/verify_email", router.VerifySignUpCodeHandler).Methods("POST", "OPTIONS")
	signIn := r.HandleFunc(version+"/user/signin", router.SignInHandler).Methods("POST", "OPTIONS")
	refreshToken := r.HandleFunc(version+"/user/refresh_token", router.RefreshJWTHandler).Methods("POST", "OPTIONS")
	forgetPass := r.HandleFunc(version+"/user/forgot_password", router.ForgotPasswordHandler).Methods("POST", "OPTIONS")
	forgetPassVerify := r.HandleFunc(version+"/user/forget_password/verify_email", router.VerifyForgetPasswordCodeHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/user/change_password", router.ChangePasswordHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc(version+"/user", router.UpdateUserHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc(version+"/user", router.GetUserHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/user/apply_voucher", router.ApplyForVoucherHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/user/activate_voucher", router.ActivateVoucherHandler).Methods("PUT", "OPTIONS")

	r.HandleFunc(version+"/quota", router.GetQuotaHandler).Methods("GET", "OPTIONS")

	r.HandleFunc(version+"/vm", router.DeployVMHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/vm/{id}", router.GetVMHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/vm", router.ListVMsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/vm/{id}", router.DeleteVM).Methods("DELETE", "OPTIONS")
	r.HandleFunc(version+"/vm", router.DeleteAllVMs).Methods("DELETE", "OPTIONS")

	r.HandleFunc(version+"/k8s", router.K8sDeployHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/k8s", router.K8sGetAllHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/k8s", router.K8sDeleteAllHandler).Methods("DELETE", "OPTIONS")
	r.HandleFunc(version+"/k8s/{id}", router.K8sGetHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/k8s/{id}", router.K8sDeleteHandler).Methods("DELETE", "OPTIONS")

	// ADMIN ACCESS
	r.HandleFunc(version+"/user/all", router.GetAllUsersHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/voucher", router.GenerateVoucherHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/voucher", router.ListVouchersHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/voucher/{id}", router.ApproveVoucherHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc(version+"/voucher", router.ApproveAllVouchers).Methods("PUT", "OPTIONS")

	r.Use(middlewares.LoggingMW)
	r.Use(middlewares.EnableCors)
	excludedRoutes := []*mux.Route{signUp, signUpVerify, signIn, refreshToken, forgetPass, forgetPassVerify}
	r.Use(middlewares.Authorization(excludedRoutes, configuration.Token.Secret, configuration.Token.Timeout))
	http.Handle("/", r)

	return &Server{port: configuration.Server.Port, host: configuration.Server.Host}, nil
}

// Start starts the server
func (s *Server) Start() (err error) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msgf("Server is listening on %s%s", s.host, s.port)

	srv := &http.Server{
		Addr: s.port,
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
		log.Info().Msg("Stopped serving new connections")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("HTTP shutdown error")
	}
	log.Info().Msg("Graceful shutdown complete")

	return nil
}
