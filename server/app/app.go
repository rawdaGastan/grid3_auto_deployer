// Package app for c4s backend app
package app

import (
	"context"
	"net/http"
	"time"

	c4sDeployer "github.com/codescalers/cloud4students/deployer"
	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/routes"
	"github.com/codescalers/cloud4students/streams"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

// App for all dependencies of backend server
type App struct {
	config   internal.Configuration
	server   server
	router   routes.Router
	db       models.DB
	redis    streams.RedisClient
	deployer c4sDeployer.Deployer
}

// NewApp creates new server app all configurations
func NewApp(ctx context.Context, file string) (app App, err error) {
	data, err := internal.ReadConfFile(file)
	if err != nil {
		return
	}

	config, err := internal.ParseConf(data)
	if err != nil {
		return
	}

	db := models.NewDB()
	err = db.Connect(config.Database.File)
	if err != nil {
		return
	}
	err = db.Migrate()
	if err != nil {
		return
	}

	redis, err := streams.NewRedisClient(config)
	if err != nil {
		return
	}

	tfPluginClient, err := deployer.NewTFPluginClient(config.Account.Mnemonics, "sr25519", config.Account.Network, "", "", "", 0, false)
	if err != nil {
		return
	}

	newDeployer, err := c4sDeployer.NewDeployer(db, redis, tfPluginClient)
	if err != nil {
		return
	}

	router, err := routes.NewRouter(config, db, redis, newDeployer)
	if err != nil {
		return
	}

	server, err := newServer(ctx, config)
	if err != nil {
		return
	}

	return App{
		config:   config,
		server:   server,
		router:   router,
		db:       db,
		redis:    redis,
		deployer: newDeployer,
	}, nil
}

// Start starts the app
func (a *App) Start(ctx context.Context) (err error) {
	a.registerHandlers()
	a.startBackgroundWorkers(ctx)

	// check pending deployments
	a.deployer.ConsumeVMRequest(ctx, true)
	a.deployer.ConsumeK8sRequest(ctx, true)

	return a.server.start()
}

func (a *App) startBackgroundWorkers(ctx context.Context) {
	// notify admins
	go a.notifyAdmins()

	// periodic deployments
	go a.deployer.PeriodicRequests(ctx, substrateBlockDiffInSeconds)
	go a.deployer.PeriodicDeploy(ctx, substrateBlockDiffInSeconds)
}

// NotifyAdmins is used to notify admins that there are new vouchers requests
func (a *App) notifyAdmins() {
	ticker := time.NewTicker(time.Hour * time.Duration(a.config.NotifyAdminsIntervalHours))

	for range ticker.C {
		pending, err := a.db.GetAllPendingVouchers()
		if err != nil {
			log.Error().Err(err).Send()
		}

		if len(pending) > 0 {
			subject, body := internal.NotifyAdminsMailContent(len(pending))

			admins, err := a.db.ListAdmins()
			if err != nil {
				log.Error().Err(err).Send()
			}

			for _, admin := range admins {
				err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, admin.Email, subject, body)
				if err != nil {
					log.Error().Err(err).Send()
				}
			}
		}
	}
}

func (a *App) registerHandlers() {
	version := "/" + a.config.Version

	r := mux.NewRouter()
	signUp := r.HandleFunc(version+"/user/signup", a.router.SignUpHandler).Methods("POST", "OPTIONS")
	signUpVerify := r.HandleFunc(version+"/user/signup/verify_email", a.router.VerifySignUpCodeHandler).Methods("POST", "OPTIONS")
	signIn := r.HandleFunc(version+"/user/signin", a.router.SignInHandler).Methods("POST", "OPTIONS")
	refreshToken := r.HandleFunc(version+"/user/refresh_token", a.router.RefreshJWTHandler).Methods("POST", "OPTIONS")
	forgetPass := r.HandleFunc(version+"/user/forgot_password", a.router.ForgotPasswordHandler).Methods("POST", "OPTIONS")
	forgetPassVerify := r.HandleFunc(version+"/user/forget_password/verify_email", a.router.VerifyForgetPasswordCodeHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/user/change_password", a.router.ChangePasswordHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc(version+"/user", a.router.UpdateUserHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc(version+"/user", a.router.GetUserHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/user/apply_voucher", a.router.ApplyForVoucherHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/user/activate_voucher", a.router.ActivateVoucherHandler).Methods("PUT", "OPTIONS")

	r.HandleFunc(version+"/quota", a.router.GetQuotaHandler).Methods("GET", "OPTIONS")

	r.HandleFunc(version+"/notification", a.router.ListNotificationsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/notification/{id}", a.router.UpdateNotificationsHandler).Methods("PUT", "OPTIONS")

	r.HandleFunc(version+"/vm", a.router.DeployVMHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/vm/{id}", a.router.GetVMHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/vm", a.router.ListVMsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/vm/{id}", a.router.DeleteVM).Methods("DELETE", "OPTIONS")
	r.HandleFunc(version+"/vm", a.router.DeleteAllVMs).Methods("DELETE", "OPTIONS")

	r.HandleFunc(version+"/k8s", a.router.K8sDeployHandler).Methods("POST", "OPTIONS")
	r.HandleFunc(version+"/k8s", a.router.K8sGetAllHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/k8s", a.router.K8sDeleteAllHandler).Methods("DELETE", "OPTIONS")
	r.HandleFunc(version+"/k8s/{id}", a.router.K8sGetHandler).Methods("GET", "OPTIONS")
	r.HandleFunc(version+"/k8s/{id}", a.router.K8sDeleteHandler).Methods("DELETE", "OPTIONS")

	maintenance := r.HandleFunc(version+"/maintenance", a.router.GetMaintenanceHandler).Methods("GET", "OPTIONS")

	// ADMIN ACCESS
	listUsers := r.HandleFunc(version+"/user/all", a.router.GetAllUsersHandler).Methods("GET", "OPTIONS")
	generateVoucher := r.HandleFunc(version+"/voucher", a.router.GenerateVoucherHandler).Methods("POST", "OPTIONS")
	listVouchers := r.HandleFunc(version+"/voucher", a.router.ListVouchersHandler).Methods("GET", "OPTIONS")
	updateVoucherRequest := r.HandleFunc(version+"/voucher/{id}", a.router.UpdateVoucherHandler).Methods("PUT", "OPTIONS")
	approveAllVouchers := r.HandleFunc(version+"/voucher", a.router.ApproveAllVouchers).Methods("PUT", "OPTIONS")
	updateMaintenance := r.HandleFunc(version+"/maintenance", a.router.UpdateMaintenanceHandler).Methods("PUT", "OPTIONS")

	// middlewares
	r.Use(middlewares.LoggingMW)
	r.Use(middlewares.EnableCors)
	excludedRoutes := []*mux.Route{maintenance, signUp, signUpVerify, signIn, refreshToken, forgetPass, forgetPassVerify}
	r.Use(middlewares.Authorization(excludedRoutes, a.config.Token.Secret, a.config.Token.Timeout))
	includedRoutes := []*mux.Route{listUsers, generateVoucher, listVouchers, updateVoucherRequest, approveAllVouchers, updateMaintenance}
	r.Use(middlewares.AdminAccess(includedRoutes, a.db))

	// prometheus registration
	prometheus.MustRegister(middlewares.Requests, middlewares.UserCreations, middlewares.VoucherActivated, middlewares.VoucherApplied, middlewares.Deployments, middlewares.Deletions)
	http.Handle("/metrics", promhttp.Handler())

	http.Handle("/", r)
}
