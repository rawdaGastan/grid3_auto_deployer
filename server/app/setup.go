// Package app for c4s backend app
package app

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"

	"testing"

	c4sDeployer "github.com/codescalers/cloud4students/deployer"
	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type authHandlerConfig struct {
	unAuthHandlerConfig
	userID string
	token  string
	config internal.Configuration
	db     models.DB
	varID  int
}

type unAuthHandlerConfig struct {
	body        io.Reader
	handlerFunc Handler
	api         string
}

// SetUp sets the needed configuration for testing
func SetUp(t testing.TB) *App {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.json")
	dbPath := filepath.Join(dir, "testing.db")

	config := fmt.Sprintf(`
{
	"server": {
		"host": "localhost",
		"port": ":3000",
		"redisHost": "localhost",
		"redisPort": "6379",
		"redisPass": ""		
	},
	"mailSender": {
      "email": "email",
      "sendgrid_key": "my sendgrid_key",
      "timeout": 60 
    },
    "account": {
      "mnemonics": "winner giant reward damage expose pulse recipe manual brand volcano dry avoid",
			"network": "dev"
    },
	"token": {
      "secret": "secret",
      "timeout": 10
    },
	"database": {
      "file": "%s"
    },
	"version": "v1",
	"currency": "eur",
	"prices": {
		"public_ip": 2,
		"small_vm": 10,
		"medium_vm": 20,
		"large_vm": 30
	},
	"stripe_secret": "sk_test"
}
	`, dbPath)

	err := os.WriteFile(configPath, []byte(config), 0644)
	assert.NoError(t, err)

	configuration, err := internal.ReadConfFile(configPath)
	assert.NoError(t, err)

	db := models.NewDB()
	err = db.Connect(configuration.Database.File)
	assert.NoError(t, err)

	err = db.Migrate()
	assert.NoError(t, err)

	tfPluginClient, err := deployer.NewTFPluginClient(configuration.Account.Mnemonics, deployer.WithNetwork(configuration.Account.Network))
	assert.NoError(t, err)

	newDeployer, err := c4sDeployer.NewDeployer(db, streams.RedisClient{}, tfPluginClient, configuration.PricesPerMonth)
	assert.NoError(t, err)

	app := &App{
		config:   configuration,
		server:   server{},
		db:       db,
		redis:    streams.RedisClient{},
		deployer: newDeployer,
	}

	return app
}

func unAuthorizedHandler(req unAuthHandlerConfig) (response *httptest.ResponseRecorder) {
	request := httptest.NewRequest("GET", req.api, req.body)
	response = httptest.NewRecorder()

	WrapFunc(req.handlerFunc).ServeHTTP(response, request)
	return
}

func authorizedNoMiddlewareHandler(req authHandlerConfig) (response *httptest.ResponseRecorder) {
	request := httptest.NewRequest("GET", req.api, req.body)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", req.token))
	ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), req.userID)
	newRequest := request.WithContext(ctx)
	response = httptest.NewRecorder()

	WrapFunc(req.handlerFunc).ServeHTTP(response, newRequest)
	return
}

func authorizedHandler(req authHandlerConfig) (response *httptest.ResponseRecorder) {
	request := httptest.NewRequest("GET", req.api, req.body)

	// add id to url vars if it has id as last index in the api request
	if req.varID != 0 {
		request = mux.SetURLVars(request, map[string]string{
			"id": fmt.Sprint(req.varID),
		})
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", req.token))
	response = httptest.NewRecorder()

	handler := WrapFunc(req.handlerFunc)
	handlerWithAuth := middlewares.Authorization(req.db, req.config.Token.Secret, req.config.Token.Timeout)(handler)
	handlerWithAuth.ServeHTTP(response, request)
	return
}

func adminHandler(req authHandlerConfig) (response *httptest.ResponseRecorder) {
	request := httptest.NewRequest("GET", req.api, req.body)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", req.token))
	response = httptest.NewRecorder()

	handler := WrapFunc(req.handlerFunc)
	handlerWithAdmin := middlewares.AdminAccess(req.db)(handler)
	handlerWithAuth := middlewares.Authorization(req.db, req.config.Token.Secret, req.config.Token.Timeout)(handlerWithAdmin)
	handlerWithAuth.ServeHTTP(response, request)
	return
}
