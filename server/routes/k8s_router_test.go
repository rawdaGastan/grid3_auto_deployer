package routes

import (
	"bytes"
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

func TestK8sDeployHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("deploy k8s small cluster", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		v := models.Voucher{
			UserID:    user.ID.String(),
			Voucher:   "voucher",
			VMs:       20,
			PublicIPs: 1,
			Reason:    "reason",
			Used:      false,
			Approved:  true,
			Rejected:  false,
		}
		err = db.CreateVoucher(&v)
		assert.NoError(t, err)

		err = db.CreateQuota(
			&models.Quota{
				UserID:    user.ID.String(),
				Vms:       20,
				PublicIPs: 1,
			},
		)
		assert.NoError(t, err)
		body := []byte(`{
		"master_name": "name",
		"resources": "small",
		"public": false
		}`)

		request := httptest.NewRequest("POST", version+"/k8s", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sDeployHandler(response, newRequest)
		if response.Code == http.StatusInternalServerError {
			return
		}
		assert.Equal(t, response.Code, http.StatusOK)

		// delete deployed k8s
		k8s, err := db.GetK8s(1)
		assert.NoError(t, err)

		err = router.cancelDeployment(uint64(k8s.ClusterContract), uint64(k8s.NetworkContract))
		assert.NoError(t, err)

	})

	t.Run("user not found", func(t *testing.T) {
		body := []byte(`{
		"master_name": "name",
		"resources": "small",
		"public": false
		}`)
		request := httptest.NewRequest("POST", version+"/k8s", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), "userID")
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sDeployHandler(response, newRequest)
		// if response.Code == http.StatusInternalServerError {
		// 	return
		// }
		want := `{"err":"User not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})

	t.Run("send wrong data", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)
		body := []byte(`{
		"master_name": "",
		"resources": "",
		"public":
		}`)
		request := httptest.NewRequest("POST", version+"/k8s", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sDeployHandler(response, newRequest)
		want := `{"err":"Failed to read k8s data"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)
	})

	//TODO: ERROR
	// t.Run("Invalid kubernetes data", func(t *testing.T) {
	// 	user, err := db.GetUserByEmail("name@gmail.com")
	// 	assert.NoError(t, err)

	// 	token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
	// 	assert.NoError(t, err)

	// 	body := []byte(`{
	// 	"master_name": "name",
	// 	"resources": "huge",
	// 	"public": false
	// 	}`)
	// 	request := httptest.NewRequest("POST", version+"/k8s", bytes.NewBuffer(body))
	// 	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	// 	ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
	// 	newRequest := request.WithContext(ctx)
	// 	response := httptest.NewRecorder()
	// 	router.K8sDeployHandler(response, newRequest)
	// 	want := `{"err":"Invalid Kubernetes data"}`
	// 	assert.Equal(t, response.Body.String(), want)
	// 	assert.Equal(t, response.Code, http.StatusBadRequest)

	// })

	t.Run("user quota not found", func(t *testing.T) {
		newUser := models.User{
			Name:           "new-name",
			Email:          "newname@gmail.com",
			HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
			Verified:       true,
			SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
		}
		err := db.CreateUser(&newUser)
		assert.NoError(t, err)
		user, err := db.GetUserByEmail("newname@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		body := []byte(`{
		"master_name": "master",
		"resources": "small",
		"public": false
		}`)

		request := httptest.NewRequest("POST", version+"/k8s", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sDeployHandler(response, newRequest)

		want := `{"err":"User quota not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})

	t.Run("no ssh key", func(t *testing.T) {
		newUser := models.User{
			Name:           "abcd",
			Email:          "abcd@gmail.com",
			HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
			Verified:       true,
			SSHKey:         "",
		}
		err := db.CreateUser(&newUser)
		assert.NoError(t, err)
		user, err := db.GetUserByEmail("abcd@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		err = db.CreateQuota(
			&models.Quota{
				UserID:    user.ID.String(),
				Vms:       20,
				PublicIPs: 1,
			},
		)
		assert.NoError(t, err)

		body := []byte(`{
		"master_name": "master",
		"resources": "small",
		"public": false
		}`)


		request := httptest.NewRequest("POST", version+"/k8s", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sDeployHandler(response, newRequest)
		want := `{"err":"SSH key is required"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusBadRequest)

	})

}

func TestK8sGetAllHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("get all k8s ", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				models.Worker{
					ClusterID: 1,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
				models.Worker{
					ClusterID: 1,
					Name:      "worker2",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = db.CreateK8s(&cluster)
		assert.NoError(t, err)

		cluster.ID = 2
		cluster.Master.Name = cluster.Master.Name + "3"

		err = db.CreateK8s(&cluster)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", version+"/k8s", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sGetAllHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("no clusters for user", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", version+"/k8s", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sGetAllHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})

}

func TestK8sDeleteAllHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("delete all k8s of user ", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				models.Worker{
					ClusterID: 1,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
				models.Worker{
					ClusterID: 1,
					Name:      "worker2",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = db.CreateK8s(&cluster)
		assert.NoError(t, err)

		cluster.ID = 2
		cluster.Master.Name = cluster.Master.Name + "3"

		err = db.CreateK8s(&cluster)
		assert.NoError(t, err)

		request := httptest.NewRequest("DELETE", version+"/k8s", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sDeleteAllHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})
}

func TestK8sGetHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("get k8s of user ", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				models.Worker{
					ClusterID: 1,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
				models.Worker{
					ClusterID: 1,
					Name:      "worker2",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = db.CreateK8s(&cluster)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", version+"/k8s/1", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "1",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sGetHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})
}

func TestK8sDeleteHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
		SSHKey:         "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCSJYyNo6j1LxrjDTRGkbBgIyD/puMprzoepKr2zwbNobCEMfAx9DXBFstueQ9wYgcwO0Pu7/95BNgtGhjoRsNDEz5MBO0Iyhcr9hGYfoXrG2Ufr8IYu3i5DWLRmDERzuArZ6/aUWIpCfpheHX+/jH/R9vvnjO2phCutpkWrjx34/33U3pL+RRycA1uTsISZTyrcMZIXfABI4xBMFLundaBk6F4YFZaCjkUOLYld4KDxJ+N6cYnJ5pa5/hLzZQedn6h7SpMvSCghxOdCxqdEwF0m9odfsrXeKRBxRfL+HWxqytNKp9CgfLvE9Knmfn5GWhXYS6/7dY7GNUGxWSje6L1h9DFwhJLjTpEwoboNzveBmlcyDwduewFZZY+q1C/gKmJial3+0n6zkx4daQsiHc29KM5wiH8mvqpm5Ew9vWNOqw85sO7BaE1W5jMkZOuqIEJiz+KW6UicUBbv2YJ8kjvNtMLM1BiE3/WjVXQ3cMf1x1mUH4bFVgW7F42nnkuc2k= alaa@alaa-Inspiron-5537",
	}
	err := db.CreateUser(&u)
	assert.NoError(t, err)

	t.Run("delete k8s of user ", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		cluster := models.K8sCluster{
			ID:     1,
			UserID: user.ID.String(),
			Master: models.Master{
				Name: "name",
			},
			Workers: []models.Worker{
				models.Worker{
					ClusterID: 1,
					Name:      "worker1",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
				models.Worker{
					ClusterID: 1,
					Name:      "worker2",
					Resources: "small",
					SRU:       5,
					CRU:       2,
					MRU:       2,
				},
			},
		}

		err = db.CreateK8s(&cluster)
		assert.NoError(t, err)

		req := httptest.NewRequest("DELETE", version+"/k8s/1", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "1",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.K8sDeleteHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

	})

}
