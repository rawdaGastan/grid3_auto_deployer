// Package app for c4s backend app
package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/routes"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var substrateBlockDiffInSeconds = 6

// Server struct holds port of server
type server struct {
	host string
	port string
}

// NewServer create new server with all configurations
func newServer(ctx context.Context, config internal.Configuration, router routes.Router, db models.DB) (server server, err error) {
	version := "/" + config.Version

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

	r.HandleFunc(version+"/notification", router.ListNotificationsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/notification/{id}", router.UpdateNotificationsHandler).Methods("PUT", "OPTIONS")

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

	maintenance := r.HandleFunc(version+"/maintenance", router.GetMaintenanceHandler).Methods("GET", "OPTIONS")

	// ADMIN ACCESS
	listUsers := r.HandleFunc(version+"/user/all", router.GetAllUsersHandler).Methods("GET", "OPTIONS")
	generateVoucher := r.HandleFunc(version+"/voucher", router.GenerateVoucherHandler).Methods("POST", "OPTIONS")
	listVouchers := r.HandleFunc(version+"/voucher", router.ListVouchersHandler).Methods("GET", "OPTIONS")
	updateVoucherRequest := r.HandleFunc(version+"/voucher/{id}", router.UpdateVoucherHandler).Methods("PUT", "OPTIONS")
	approveAllVouchers := r.HandleFunc(version+"/voucher", router.ApproveAllVouchers).Methods("PUT", "OPTIONS")
	updateMaintenance := r.HandleFunc(version+"/maintenance", router.UpdateMaintenanceHandler).Methods("PUT", "OPTIONS")

	// middlewares
	r.Use(middlewares.LoggingMW)
	r.Use(middlewares.EnableCors)
	excludedRoutes := []*mux.Route{maintenance, signUp, signUpVerify, signIn, refreshToken, forgetPass, forgetPassVerify}
	r.Use(middlewares.Authorization(excludedRoutes, config.Token.Secret, config.Token.Timeout))
	includedRoutes := []*mux.Route{listUsers, generateVoucher, listVouchers, updateVoucherRequest, approveAllVouchers, updateMaintenance}
	r.Use(middlewares.AdminAccess(includedRoutes, db))

	// prometheus registration
	prometheus.MustRegister(middlewares.Requests, middlewares.UserCreations, middlewares.VoucherActivated, middlewares.VoucherApplied, middlewares.Deployments, middlewares.Deletions)
	http.Handle("/metrics", promhttp.Handler())

	http.Handle("/", r)

	// check pending deployments
	router.Deployer.ConsumeVMRequest(ctx, true)
	router.Deployer.ConsumeK8sRequest(ctx, true)

	server.port = config.Server.Port
	server.host = config.Server.Host
	return
}

// Start starts the server
func (s *server) start() (err error) {
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
