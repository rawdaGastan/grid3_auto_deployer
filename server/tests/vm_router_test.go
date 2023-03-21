package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rawdaGastan/cloud4students/models"
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
	if err != nil {
		t.Error(err)
	}

	t.Run("deploy medium vm successfully", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		if err != nil {
			t.Error(err)
		}
		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		if err != nil {
			t.Error(err)
		}
		v := models.Voucher{
			Voucher: "voucher",
			K8s:     10,
			VMs:     10,
		}
		err = db.CreateVoucher(&v)
		if err != nil {
			t.Error(err)
		}
		err = db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    10,
				K8s:    10,
			},
		)
		if err != nil {
			t.Error(err)
		}
		err = db.AddUserVoucher(user.ID.String(), v.Voucher)
		if err != nil {
			t.Error(err)
		}
		body := []byte(`{
		"name" : "vm",
		"resources" : "medium"
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
		if err != nil {
			t.Error(err)
		}
		err = router.CancelDeployment(vm.ContractID, vm.NetworkContractID)
		if err != nil {
			t.Error(err)
		}
	})

}

// TODO: Error
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
	if err != nil {
		t.Error(err)
	}
	t.Run("get vm of user", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		if err != nil {
			t.Error(err)
		}
		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		if err != nil {
			t.Error(err)
		}
		vm := models.VM{
			ID:        1,
			UserID:    user.ID.String(),
			Name:      "vm",
			IP:        "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}
		err = db.CreateVM(&vm)
		if err != nil {
			t.Error(err)
			fmt.Printf("err: %v\n", err)
		}
		request := httptest.NewRequest("GET", version+"/vm/1", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		fmt.Printf("user.ID.String(): %v\n", user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.GetVMHandler(response, newRequest)
		body := response.Body.String()
		fmt.Printf("body: %v\n", body)
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
	if err != nil {
		t.Error(err)
	}
	t.Run("list all vms of user", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		if err != nil {
			t.Error(err)
		}
		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		if err != nil {
			t.Error(err)
		}
		vm := models.VM{
			ID:        1,
			UserID:    user.ID.String(),
			Name:      "vm",
			IP:        "10.1.0.0",
			Resources: "small",
			SRU:       5,
			CRU:       2,
			MRU:       2,
		}
		err = db.CreateVM(&vm)
		if err != nil {
			t.Error(err)
		}
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
		if err != nil {
			t.Error(err)
		}
		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		if err != nil {
			t.Error(err)
		}
		request := httptest.NewRequest("GET", version+"/vm", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		fmt.Printf("user.ID.String(): %v\n", user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.ListVMsHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})
}

// TODO:
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
	if err != nil {
		t.Error(err)
	}

	t.Run("deploy vm then delete it", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		if err != nil {
			t.Error(err)
		}
		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		if err != nil {
			t.Error(err)
		}
		v := models.Voucher{
			Voucher: "voucher",
			K8s:     10,
			VMs:     10,
		}
		err = db.CreateVoucher(&v)
		if err != nil {
			t.Error(err)
		}
		err = db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    10,
				K8s:    10,
			},
		)
		if err != nil {
			t.Error(err)
		}
		err = db.AddUserVoucher(user.ID.String(), v.Voucher)
		if err != nil {
			t.Error(err)
		}
		body := []byte(`{
		"name" : "vm",
		"resources" : "medium"
		}`)
		// deploy vm
		request1 := httptest.NewRequest("POST", version+"/vm", bytes.NewBuffer(body))
		request1.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request1.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest1 := request1.WithContext(ctx)
		response1 := httptest.NewRecorder()
		router.DeployVMHandler(response1, newRequest1)
		assert.Equal(t, response1.Code, http.StatusOK)

		// delete vm
		request2 := httptest.NewRequest("DELETE", version+"/vm/1", bytes.NewBuffer(body))
		request2.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx = context.WithValue(request2.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest2 := request2.WithContext(ctx)
		response2 := httptest.NewRecorder()
		router.DeleteVM(response2, newRequest2)
		assert.Equal(t, response2.Code, http.StatusOK)

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
	if err != nil {
		t.Error(err)
	}

	t.Run("deploy vms then delete them", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		if err != nil {
			t.Error(err)
		}
		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		if err != nil {
			t.Error(err)
		}
		v := models.Voucher{
			Voucher: "voucher",
			K8s:     10,
			VMs:     10,
		}
		err = db.CreateVoucher(&v)
		if err != nil {
			t.Error(err)
		}
		err = db.CreateQuota(
			&models.Quota{
				UserID: user.ID.String(),
				Vms:    10,
				K8s:    10,
			},
		)
		if err != nil {
			t.Error(err)
		}
		err = db.AddUserVoucher(user.ID.String(), v.Voucher)
		if err != nil {
			t.Error(err)
		}
		body := []byte(`{
		"name" : "vm1",
		"resources" : "medium"
		}`)
		// deploy vm1
		request1 := httptest.NewRequest("POST", version+"/vm", bytes.NewBuffer(body))
		request1.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request1.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest1 := request1.WithContext(ctx)
		response1 := httptest.NewRecorder()
		router.DeployVMHandler(response1, newRequest1)
		fmt.Printf("response1.Body.String(): %v\n", response1.Body.String())
		assert.Equal(t, response1.Code, http.StatusOK)

		time.Sleep(20 * time.Second)
		// deploy vm2
		request2 := httptest.NewRequest("POST", version+"/vm", bytes.NewBuffer(body))
		request2.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx2 := context.WithValue(request2.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest2 := request2.WithContext(ctx2)
		response2 := httptest.NewRecorder()
		router.DeployVMHandler(response2, newRequest2)
		fmt.Printf("response2.Body.String(): %v\n", response2.Body.String())
		assert.Equal(t, response2.Code, http.StatusOK)

		// delete vms
		request3 := httptest.NewRequest("DELETE", version+"/vm", bytes.NewBuffer(body))
		request3.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx3 := context.WithValue(request3.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest3 := request3.WithContext(ctx3)
		response3 := httptest.NewRecorder()
		router.DeleteAllVMs(response3, newRequest3)
		fmt.Printf("response3.Body.String(): %v\n", response3.Body.String())
		assert.Equal(t, response3.Code, http.StatusOK)

	})

}
