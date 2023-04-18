package routes

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/stretchr/testify/assert"
)

func TestQuotaRouter(t *testing.T) {
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

	t.Run("get quota of user", func(t *testing.T) {
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

		request := httptest.NewRequest("GET", version+"/quota", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), user.ID.String())
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.GetQuotaHandler(response, newRequest)
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("user quota not found", func(t *testing.T) {
		token, err := internal.CreateJWT("wrongID", "wrong@gmail.com", config.Token.Secret, config.Token.Timeout)
		assert.NoError(t, err)

		request := httptest.NewRequest("GET", version+"/quota", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		ctx := context.WithValue(request.Context(), middlewares.UserIDKey("UserID"), "wrongID")
		newRequest := request.WithContext(ctx)
		response := httptest.NewRecorder()
		router.GetQuotaHandler(response, newRequest)
		want := `{"err":"user quota not found"}`
		assert.Equal(t, response.Body.String(), want)
		assert.Equal(t, response.Code, http.StatusNotFound)

	})
}
