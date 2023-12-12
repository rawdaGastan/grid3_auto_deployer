// Package app for c4s backend app
package app

import (
	"context"
	"net/http"

	c4sDeployer "github.com/codescalers/cloud4students/deployer"
	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

// App for all dependencies of backend server
type App struct {
	config   internal.Configuration
	server   server
	db       models.DB
	redis    streams.RedisClient
	deployer c4sDeployer.Deployer
}

// NewApp creates new server app all configurations
func NewApp(ctx context.Context, configFile string) (app *App, err error) {
	config, err := internal.ReadConfFile(configFile)
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

	server := newServer(config.Server.Host, config.Server.Port)
	if err != nil {
		return
	}

	return &App{
		config:   config,
		server:   *server,
		db:       db,
		redis:    redis,
		deployer: newDeployer,
	}, nil
}

// Start starts the app
func (a *App) Start(ctx context.Context) (err error) {
	a.registerHandlers()
	a.startBackgroundWorkers(ctx)

	return a.server.start()
}

func (a *App) startBackgroundWorkers(ctx context.Context) {
	// notify admins
	go a.notifyAdmins()

	// periodic deployments
	go a.deployer.PeriodicRequests(ctx, substrateBlockDiffInSeconds)
	go a.deployer.PeriodicDeploy(ctx, substrateBlockDiffInSeconds)

	// send notification about vms and k8s expiration
	go a.deployer.WarnUsersWithExpiredVMs(ctx)
	go a.deployer.WarnUsersWithExpiredK8s(ctx)

	// remove expired vms and k8s
	go a.deployer.CleanExpiredVMs(ctx)
	go a.deployer.CleanExpiredK8S(ctx)

	// check pending deployments
	a.deployer.ConsumeVMRequest(ctx, true)
	a.deployer.ConsumeK8sRequest(ctx, true)
}

func (a *App) registerHandlers() {
	r := mux.NewRouter()

	// version router
	versionRouter := r.PathPrefix("/" + a.config.Version).Subrouter()
	authRouter := versionRouter.NewRoute().Subrouter()
	adminRouter := authRouter.NewRoute().Subrouter()

	// sub routes with authorization
	userRouter := authRouter.PathPrefix("/user").Subrouter()
	quotaRouter := authRouter.PathPrefix("/quota").Subrouter()
	notificationRouter := authRouter.PathPrefix("/notification").Subrouter()
	vmRouter := authRouter.PathPrefix("/vm").Subrouter()
	k8sRouter := authRouter.PathPrefix("/k8s").Subrouter()

	// sub routes with no authorization
	unAuthUserRouter := versionRouter.PathPrefix("/user").Subrouter()
	unAuthMaintenanceRouter := versionRouter.PathPrefix("/maintenance").Subrouter()

	// sub routes with admin access
	voucherRouter := adminRouter.PathPrefix("/voucher").Subrouter()
	maintenanceRouter := adminRouter.PathPrefix("/maintenance").Subrouter()
	balanceRouter := adminRouter.PathPrefix("/balance").Subrouter()

	unAuthUserRouter.HandleFunc("/signup", WrapFunc(a.SignUpHandler)).Methods("POST", "OPTIONS")
	unAuthUserRouter.HandleFunc("/signup/verify_email", WrapFunc(a.VerifySignUpCodeHandler)).Methods("POST", "OPTIONS")
	unAuthUserRouter.HandleFunc("/signin", WrapFunc(a.SignInHandler)).Methods("POST", "OPTIONS")
	unAuthUserRouter.HandleFunc("/refresh_token", WrapFunc(a.RefreshJWTHandler)).Methods("POST", "OPTIONS")
	unAuthUserRouter.HandleFunc("/forgot_password", WrapFunc(a.ForgotPasswordHandler)).Methods("POST", "OPTIONS")
	unAuthUserRouter.HandleFunc("/forget_password/verify_email", WrapFunc(a.VerifyForgetPasswordCodeHandler)).Methods("POST", "OPTIONS")

	userRouter.HandleFunc("/change_password", WrapFunc(a.ChangePasswordHandler)).Methods("PUT", "OPTIONS")
	userRouter.HandleFunc("", WrapFunc(a.UpdateUserHandler)).Methods("PUT", "OPTIONS")
	userRouter.HandleFunc("", WrapFunc(a.GetUserHandler)).Methods("GET", "OPTIONS")
	userRouter.HandleFunc("/apply_voucher", WrapFunc(a.ApplyForVoucherHandler)).Methods("POST", "OPTIONS")
	userRouter.HandleFunc("/activate_voucher", WrapFunc(a.ActivateVoucherHandler)).Methods("PUT", "OPTIONS")

	quotaRouter.HandleFunc("", WrapFunc(a.GetQuotaHandler)).Methods("GET", "OPTIONS")

	notificationRouter.HandleFunc("", WrapFunc(a.ListNotificationsHandler)).Methods("GET", "OPTIONS")
	notificationRouter.HandleFunc("/{id}", WrapFunc(a.UpdateNotificationsHandler)).Methods("PUT", "OPTIONS")

	vmRouter.HandleFunc("", WrapFunc(a.DeployVMHandler)).Methods("POST", "OPTIONS")
	vmRouter.HandleFunc("/validate/{name}", WrapFunc(a.ValidateVMNameHandler)).Methods("Get", "OPTIONS")
	vmRouter.HandleFunc("/{id}", WrapFunc(a.GetVMHandler)).Methods("GET", "OPTIONS")
	vmRouter.HandleFunc("/{id}", WrapFunc(a.DeleteVMHandler)).Methods("DELETE", "OPTIONS")
	vmRouter.HandleFunc("", WrapFunc(a.ListVMsHandler)).Methods("GET", "OPTIONS")
	vmRouter.HandleFunc("", WrapFunc(a.DeleteAllVMsHandler)).Methods("DELETE", "OPTIONS")

	k8sRouter.HandleFunc("", WrapFunc(a.K8sDeployHandler)).Methods("POST", "OPTIONS")
	k8sRouter.HandleFunc("/validate/{name}", WrapFunc(a.ValidateK8sNameHandler)).Methods("Get", "OPTIONS")
	k8sRouter.HandleFunc("/{id}", WrapFunc(a.K8sGetHandler)).Methods("GET", "OPTIONS")
	k8sRouter.HandleFunc("/{id}", WrapFunc(a.K8sDeleteHandler)).Methods("DELETE", "OPTIONS")
	k8sRouter.HandleFunc("", WrapFunc(a.K8sGetAllHandler)).Methods("GET", "OPTIONS")
	k8sRouter.HandleFunc("", WrapFunc(a.K8sDeleteAllHandler)).Methods("DELETE", "OPTIONS")

	unAuthMaintenanceRouter.HandleFunc("", WrapFunc(a.GetMaintenanceHandler)).Methods("GET", "OPTIONS")

	// ADMIN ACCESS
	adminRouter.HandleFunc("/user/all", WrapFunc(a.GetAllUsersHandler)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/deployment/count", WrapFunc(a.GetDlsCountHandler)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/announcement", WrapFunc(a.CreateNewAnnouncement)).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/set_admin", WrapFunc(a.SetAdmin)).Methods("PUT", "OPTIONS")
	balanceRouter.HandleFunc("", WrapFunc(a.GetBalanceHandler)).Methods("GET", "OPTIONS")
	maintenanceRouter.HandleFunc("", WrapFunc(a.UpdateMaintenanceHandler)).Methods("PUT", "OPTIONS")

	voucherRouter.HandleFunc("", WrapFunc(a.GenerateVoucherHandler)).Methods("POST", "OPTIONS")
	voucherRouter.HandleFunc("", WrapFunc(a.ListVouchersHandler)).Methods("GET", "OPTIONS")
	voucherRouter.HandleFunc("/{id}", WrapFunc(a.UpdateVoucherHandler)).Methods("PUT", "OPTIONS")
	voucherRouter.HandleFunc("", WrapFunc(a.ApproveAllVouchersHandler)).Methods("PUT", "OPTIONS")

	// middlewares
	r.Use(middlewares.LoggingMW)
	r.Use(middlewares.EnableCors)

	authRouter.Use(middlewares.Authorization(a.db, a.config.Token.Secret, a.config.Token.Timeout))
	adminRouter.Use(middlewares.AdminAccess(a.db))

	// prometheus registration
	prometheus.MustRegister(middlewares.Requests, middlewares.UserCreations, middlewares.VoucherActivated, middlewares.VoucherApplied, middlewares.Deployments, middlewares.Deletions)
	http.Handle("/metrics", promhttp.Handler())

	http.Handle("/", r)
}
