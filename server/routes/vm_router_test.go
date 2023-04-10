// Package routes for API endpoints
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

func TestDeployVMHandler(t *testing.T) {
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

	t.Run("deploy small vm successfully", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		v := models.Voucher{
			UserID:    user.ID.String(),
			Voucher:   "voucher",
			VMs:       10,
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
				Vms:       10,
				PublicIPs: 1,
			},
		)
		assert.NoError(t, err)
		body := []byte(`{
		"name": "name",
		"resources": "small",
		"public": false
		}`)
		request := httptest.NewRequest("POST", version+"/vm", bytes.NewBuffer(body))
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.DeployVMHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

		// delete deployed vm
		vm, err := db.GetVMByID(1)
		assert.NoError(t, err)

		err = router.cancelDeployment(vm.ContractID, vm.NetworkContractID)
		assert.NoError(t, err)
	})
}

func TestGetVMHandler(t *testing.T) {
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

	t.Run("get vm of user", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		vm := models.VM{
			ID:        1,
			UserID:    user.ID.String(),
			Name:      "vm",
			YggIP:     "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}
		err = db.CreateVM(&vm)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", version+"/vm/1", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "1",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.GetVMHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}

func TestListVMsHandler(t *testing.T) {
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

	t.Run("list all vms of user", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		vm := models.VM{
			UserID:    user.ID.String(),
			Name:      "vm",
			YggIP:     "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}
		vm.Name = vm.Name + "0"
		err = db.CreateVM(&vm)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", version+"/vm", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.ListVMsHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("no vms for user", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", version+"/vm", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.ListVMsHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}

func TestDeleteVM(t *testing.T) {
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

	t.Run("create vm then delete it", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		vm := models.VM{
			UserID:    user.ID.String(),
			Name:      "vm",
			YggIP:     "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}
		vm.Name = vm.Name + "1"
		err = db.CreateVM(&vm)
		assert.NoError(t, err)

		// delete vm
		req := httptest.NewRequest("DELETE", version+"/vm/1", nil)
		request := mux.SetURLVars(req, map[string]string{
			"id": "1",
		})

		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.DeleteVM(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

		vms, err := db.GetAllVms(user.ID.String())
		assert.Empty(t, vms)
		assert.NoError(t, err)
	})

}

func TestDeleteDeleteAllVMs(t *testing.T) {
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

	t.Run("create vms then delete them", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		assert.NoError(t, err)

		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		vm := models.VM{
			ID:        1,
			UserID:    user.ID.String(),
			Name:      "vm2",
			YggIP:     "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}
		err = db.CreateVM(&vm)
		assert.NoError(t, err)

		vm.ID = 2
		vm.Name = vm.Name + "3"
		err = db.CreateVM(&vm)
		assert.NoError(t, err)

		// delete vms
		request := httptest.NewRequest("DELETE", version+"/vm", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx3 := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx3)
		response := httptest.NewRecorder()
		router.DeleteAllVMs(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)

		vms, err := db.GetAllVms(user.ID.String())
		assert.Empty(t, vms)
		assert.NoError(t, err)
	})

}
