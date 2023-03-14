package tests

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rawdaGastan/cloud4students/routes"
	"github.com/threefoldtech/grid3-go/deployer"
)

// tempDBFile create temporary DB file for testing
func tempDBFile(t testing.TB) string {
	file, err := os.CreateTemp("", "testing")
	defer file.Close()
	if err != nil {
		t.Fatalf("can't create temp file %q", err.Error())
	}
	return file.Name()
}

// SetUp sets the needed configuration for testing
func SetUp(t testing.TB) (r *routes.Router, version string) {
	file := tempDBFile(t)
	data, err := internal.ReadConfFile("/home/alaa/codescalers/cloud4students/server/config.json") //TODO:
	if err != nil {
		return
	}
	configuration, err := internal.ParseConf(data)
	if err != nil {
		return
	}

	db := models.NewDB()
	err = db.Connect(file)
	if err != nil {
		return
	}
	err = db.Migrate()
	if err != nil {
		return
	}

	tfPluginClient, err := deployer.NewTFPluginClient(configuration.Account.Mnemonics, "sr25519", configuration.Account.Network, "", "", "", true, false)
	if err != nil {
		return
	}

	version = "/" + configuration.Version
	router := routes.NewRouter(*configuration, db, tfPluginClient)
	return &router, version

}

func TestSignUpHandler(t *testing.T) {
	router, version := SetUp(t)
	// Json Body
	body := []byte(`{
		"name":"name",
		"email":"name@gmail.com",
		"password":"strongpass",
		"confirm_password":"strongpass"
	}`)
	t.Run("signup successfully", func(t *testing.T) {
		request := httptest.NewRequest("POST", version+"/user/signup", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.SignUpHandler(response, request)
		got := response.Body.String()
		want := `{"msg":"Verification code has been sent to name@gmail.com","data":""}`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
			fmt.Printf("got: %v\n", got)
		}
		code := response.Code
		if code != 200 {
			t.Errorf("error got %d response code, want %d", code, 200)
		}
	})

	t.Run("send empty data", func(t *testing.T) {
		request := httptest.NewRequest("POST", version+"/user/signup", nil)
		response := httptest.NewRecorder()
		router.SignUpHandler(response, request)
		code := response.Code
		if code != 500 {
			t.Errorf("error got %d response code, want %d", code, 200)
		}

	})

}
