package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/middlewares"
	"github.com/rawdaGastan/cloud4students/models"
)

// TODO:
func TestDeployVMHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(&u)
	if err != nil {
		t.Error(err)
	}

	t.Run("deploy medium vm successfully", func(t *testing.T) { // create voucher && activate it
		user, err := db.GetUserByEmail("name@gmail.com")
		fmt.Printf("user: %v\n", user)
		if err != nil {
			t.Error(err)
		}
		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
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
			t.Error(t)
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

	})

}
