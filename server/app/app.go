// Package app for c4s backend app
package app

import (
	"context"
	"net/http"

	c4sDeployer "github.com/codescalers/cloud4students/deployer"
	_ "github.com/codescalers/cloud4students/docs"
	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stripe/stripe-go/v81"
	httpSwagger "github.com/swaggo/http-swagger"
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

	stripe.Key = config.StripeSecret

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

	tfPluginClient, err := deployer.NewTFPluginClient(
		config.Account.Mnemonics,
		deployer.WithNetwork(config.Account.Network),
	)
	if err != nil {
		return
	}

	newDeployer, err := c4sDeployer.NewDeployer(db, redis, tfPluginClient, config.PricesPerMonth)
	if err != nil {
		return
	}

	server := newServer(config.Server.Host, config.Server.Port)

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

	// Invoices
	go a.monthlyInvoices()
	go a.sendRemindersToPayInvoices()

	// periodic deployments
	go a.deployer.PeriodicRequests(ctx, substrateBlockDiffInSeconds)
	go a.deployer.PeriodicDeploy(ctx, substrateBlockDiffInSeconds)

	// check pending deployments
	a.deployer.ConsumeVMRequest(ctx, true)
	a.deployer.ConsumeK8sRequest(ctx, true)
}

func (a *App) registerHandlers() {
	r := mux.NewRouter()

	// Setup Swagger UI route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// version router
	versionRouter := r.PathPrefix("/" + a.config.Version).Subrouter()
	authRouter := versionRouter.NewRoute().Subrouter()
	adminRouter := authRouter.NewRoute().Subrouter()

	// sub routes with authorization
	userRouter := authRouter.PathPrefix("/user").Subrouter()
	invoiceRouter := authRouter.PathPrefix("/invoice").Subrouter()
	cardRouter := userRouter.PathPrefix("/card").Subrouter()
	notificationRouter := authRouter.PathPrefix("/notification").Subrouter()
	vmRouter := authRouter.PathPrefix("/vm").Subrouter()
	k8sRouter := authRouter.PathPrefix("/k8s").Subrouter()
	regionRouter := authRouter.PathPrefix("/region").Subrouter()

	// sub routes with no authorization
	unAuthUserRouter := versionRouter.PathPrefix("/user").Subrouter()
	unAuthMaintenanceRouter := versionRouter.PathPrefix("/maintenance").Subrouter()
	unauthNextLaunchRouter := versionRouter.PathPrefix("/nextlaunch").Subrouter()

	// sub routes with admin access
	voucherRouter := adminRouter.PathPrefix("/voucher").Subrouter()
	maintenanceRouter := adminRouter.PathPrefix("/maintenance").Subrouter()
	balanceRouter := adminRouter.PathPrefix("/balance").Subrouter()
	deploymentsRouter := adminRouter.PathPrefix("/deployments").Subrouter()
	nextLaunchRouter := adminRouter.PathPrefix("/nextlaunch").Subrouter()

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
	userRouter.HandleFunc("/charge_balance", WrapFunc(a.ChargeBalance)).Methods("PUT", "OPTIONS")

	cardRouter.HandleFunc("", WrapFunc(a.AddCardHandler)).Methods("POST", "OPTIONS")
	cardRouter.HandleFunc("/{id}", WrapFunc(a.DeleteCardHandler)).Methods("DELETE", "OPTIONS")
	cardRouter.HandleFunc("", WrapFunc(a.ListCardHandler)).Methods("GET", "OPTIONS")
	cardRouter.HandleFunc("/default", WrapFunc(a.SetDefaultCardHandler)).Methods("PUT", "OPTIONS")

	invoiceRouter.HandleFunc("", WrapFunc(a.ListInvoicesHandler)).Methods("GET", "OPTIONS")
	invoiceRouter.HandleFunc("/{id}", WrapFunc(a.GetInvoiceHandler)).Methods("GET", "OPTIONS")
	invoiceRouter.HandleFunc("/pay/{id}", WrapFunc(a.PayInvoiceHandler)).Methods("PUT", "OPTIONS")

	notificationRouter.HandleFunc("", WrapFunc(a.ListNotificationsHandler)).Methods("GET", "OPTIONS")
	notificationRouter.HandleFunc("/{id}", WrapFunc(a.UpdateNotificationsHandler)).Methods("PUT", "OPTIONS")

	regionRouter.HandleFunc("", WrapFunc(a.ListRegionsHandler)).Methods("GET", "OPTIONS")

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
	unauthNextLaunchRouter.HandleFunc("", WrapFunc(a.GetNextLaunchHandler)).Methods("GET", "OPTIONS")

	// ADMIN ACCESS
	adminRouter.HandleFunc("/user/all", WrapFunc(a.GetAllUsersHandler)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/invoice/all", WrapFunc(a.GetAllInvoicesHandler)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/announcement", WrapFunc(a.CreateNewAnnouncementHandler)).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/email", WrapFunc(a.SendEmailHandler)).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/set_admin", WrapFunc(a.SetAdminHandler)).Methods("PUT", "OPTIONS")
	adminRouter.HandleFunc("/set_prices", WrapFunc(a.SetPricesHandler)).Methods("PUT", "OPTIONS")
	balanceRouter.HandleFunc("", WrapFunc(a.GetBalanceHandler)).Methods("GET", "OPTIONS")
	maintenanceRouter.HandleFunc("", WrapFunc(a.UpdateMaintenanceHandler)).Methods("PUT", "OPTIONS")
	deploymentsRouter.HandleFunc("", WrapFunc(a.DeleteAllDeploymentsHandler)).Methods("DELETE", "OPTIONS")
	deploymentsRouter.HandleFunc("", WrapFunc(a.ListDeploymentsHandler)).Methods("GET", "OPTIONS")
	deploymentsRouter.HandleFunc("/count", WrapFunc(a.GetDlsCountHandler)).Methods("GET", "OPTIONS")
	nextLaunchRouter.HandleFunc("", WrapFunc(a.UpdateNextLaunchHandler)).Methods("PUT", "OPTIONS")

	voucherRouter.HandleFunc("", WrapFunc(a.GenerateVoucherHandler)).Methods("POST", "OPTIONS")
	voucherRouter.HandleFunc("", WrapFunc(a.ListVouchersHandler)).Methods("GET", "OPTIONS")
	voucherRouter.HandleFunc("/{id}", WrapFunc(a.UpdateVoucherHandler)).Methods("PUT", "OPTIONS")
	voucherRouter.HandleFunc("", WrapFunc(a.ApproveAllVouchersHandler)).Methods("PUT", "OPTIONS")
	voucherRouter.HandleFunc("/reset", WrapFunc(a.ResetUsersVoucherBalanceHandler)).Methods("PUT", "OPTIONS")

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
