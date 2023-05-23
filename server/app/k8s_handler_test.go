// Package app for c4s backend app
package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/models"
	"github.com/stretchr/testify/assert"
)

func TestK8sGetAllHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("Get all k8s: no clusters for user", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sGetAllHandler,
				api:         fmt.Sprintf("/%s/k8s", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"msg":"Kubernetes clusters are not found","data":[]}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Get all k8s: success", func(t *testing.T) {
		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{},
		}

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		cluster.ID = 2
		cluster.Master.Name = cluster.Master.Name + "3"

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sGetAllHandler,
				api:         fmt.Sprintf("/%s/k8s", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)

		var res ResponseMsg
		err = json.Unmarshal(response.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, len(res.Data.([]interface{})), 2)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}

func TestK8sDeleteAllHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("Delete all k8s: no clusters found", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sDeleteAllHandler,
				api:         fmt.Sprintf("/%s/k8s", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"msg":"Kubernetes clusters are not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Delete all k8s: success", func(t *testing.T) {
		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{},
		}

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		cluster.ID = 2
		cluster.Master.Name = cluster.Master.Name + "3"

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sDeleteAllHandler,
				api:         fmt.Sprintf("/%s/k8s", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"msg":"All kubernetes clusters are deleted successfully"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}

func TestK8sGetAndDeleteHandler(t *testing.T) {
	app := SetUp(t)

	user.Verified = true
	err := app.db.CreateUser(user)
	assert.NoError(t, err)

	token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
	assert.NoError(t, err)

	t.Run("Get k8s: cluster not found", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sGetHandler,
				api:         fmt.Sprintf("/%s/k8s/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"kubernetes cluster is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Delete k8s: cluster not found", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sDeleteHandler,
				api:         fmt.Sprintf("/%s/k8s/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"kubernetes cluster is not found"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)
	})

	t.Run("Get k8s: success", func(t *testing.T) {
		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{},
		}

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sGetHandler,
				api:         fmt.Sprintf("/%s/k8s/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Delete k8s: success", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sDeleteHandler,
				api:         fmt.Sprintf("/%s/k8s/1", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("Get k8s: invalid cluster id", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sGetHandler,
				api:         fmt.Sprintf("/%s/k8s/", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read cluster id"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Delete k8s: invalid cluster id", func(t *testing.T) {
		req := authHandlerConfig{
			unAuthHandlerConfig: unAuthHandlerConfig{
				body:        nil,
				handlerFunc: app.K8sDeleteHandler,
				api:         fmt.Sprintf("/%s/k8s/", app.config.Version),
			},
			userID: user.ID.String(),
			token:  token,
			config: app.config,
			db:     app.db,
		}

		response := authorizedHandler(req)
		want := `{"err":"failed to read cluster id"}` + "\n"
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}
