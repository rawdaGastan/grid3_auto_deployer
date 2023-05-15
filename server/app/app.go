// Package app for c4s backend app
package app

import (
	"context"
	"time"

	c4sDeployer "github.com/codescalers/cloud4students/deployer"
	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/routes"
	"github.com/codescalers/cloud4students/streams"
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

	server, err := newServer(ctx, config, router, db)
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
