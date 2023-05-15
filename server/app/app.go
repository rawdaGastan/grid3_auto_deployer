// Package app for c4s backend app
package app

import (
	"context"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/routes"
	"github.com/codescalers/cloud4students/streams"
)

// App for all dependencies of backend server
type App struct {
	server server
	router routes.Router
	db     models.DB
	redis  streams.RedisClient
}

// NewApp creates new server app all configurations
func NewApp(ctx context.Context, file string) (app App, err error) {
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

	redis, err := streams.NewRedisClient(configuration)
	if err != nil {
		return
	}

	router, err := routes.NewRouter(configuration, db, redis)
	if err != nil {
		return
	}

	server, err := newServer(ctx, configuration, router, db)
	if err != nil {
		return
	}

	return App{
		server: server,
		router: router,
		db:     db,
		redis:  redis,
	}, nil
}

// Start starts the app
func (a *App) Start(ctx context.Context) (err error) {
	a.startBackgroundWorkers(ctx)
	return a.server.start()
}

func (a *App) startBackgroundWorkers(ctx context.Context) {
	// notify admins
	go a.router.NotifyAdmins()

	// periodic deployments
	go a.router.Deployer.PeriodicRequests(ctx, substrateBlockDiffInSeconds)
	go a.router.Deployer.PeriodicDeploy(ctx, substrateBlockDiffInSeconds)
}
