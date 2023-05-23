// Package app for c4s backend app
package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestK8sGetAllHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("no clusters for user", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", app.config.Version+"/k8s", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sGetAllHandler(newRequest)
		want := `{"msg":"Kubernetes clusters not found","data":[]}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)
	})
	t.Run("get all k8s ", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				{
					ClusterID: 1,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
				{
					ClusterID: 1,
					Name:      "worker2",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		cluster.ID = 2
		cluster.Master.Name = cluster.Master.Name + "3"

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", app.config.Version+"/k8s", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sGetAllHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

}

func TestK8sDeleteAllHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("no clusters found", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("DELETE", app.config.Version+"/k8s", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sDeleteAllHandler(newRequest)
		want := `{"msg":"Kubernetes clusters not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("delete all k8s of user ", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				{
					ClusterID: 1,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
				{
					ClusterID: 1,
					Name:      "worker2",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		cluster.ID = 2
		cluster.Master.Name = cluster.Master.Name + "3"

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		request := httptest.NewRequest("DELETE", app.config.Version+"/k8s", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sDeleteAllHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})
}

func TestK8sGetHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("get k8s of user ", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				{
					ClusterID: 1,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
				{
					ClusterID: 1,
					Name:      "worker2",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", app.config.Version+"/k8s/1", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "1",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sGetHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("failed to read cluster id", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", app.config.Version+"/k8s/", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sGetHandler(newRequest)
		want := `{"err":"Failed to read cluster id"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	t.Run("k8s cluster not found", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", app.config.Version+"/k8s/10", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "10",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sGetHandler(newRequest)
		want := `{"err":"Kubernetes cluster not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})
}

func TestK8sDeleteHandler(t *testing.T) {
	app := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: []byte{},
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := app.db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("delete k8s of user ", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				{
					ClusterID: 1,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
				{
					ClusterID: 1,
					Name:      "worker2",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		req := httptest.NewRequest("DELETE", app.config.Version+"/k8s/1", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "1",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sDeleteHandler(newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("failed to read k8s id", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		req := httptest.NewRequest("DELETE", app.config.Version+"/k8s/", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sDeleteHandler(newRequest)
		want := `{"err":"Failed to read cluster id"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

	t.Run("k8s cluster not found", func(t *testing.T) {
		user, err := app.db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, app.config.Token.Secret, app.config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     2,
			UserID: "userID",
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				{
					ClusterID: 2,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = app.db.CreateK8s(&cluster)
		assert.NoError(t, err)

		req := httptest.NewRequest("DELETE", app.config.Version+"/k8s/2", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "2",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		app.K8sDeleteHandler(newRequest)
		want := `{"err":"Kubernetes cluster not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})

}
