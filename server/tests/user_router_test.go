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
func SetUp(t testing.TB) (r *routes.Router, db models.DB, version string) {
	file := tempDBFile(t)
	data, err := internal.ReadConfFile("./config-temp.json")
	if err != nil {
		return
	}
	configuration, err := internal.ParseConf(data)
	if err != nil {
		return
	}

	db = models.NewDB()
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
	return &router, db, version

}

func TestSignUpHandler(t *testing.T) {
	router, _, version := SetUp(t)
	// json Body of request
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
		}
		if response.Code != 200 {
			t.Errorf("error got %d response code, want %d", response.Code, 200)
		}
	})

	t.Run("send empty data", func(t *testing.T) {
		request := httptest.NewRequest("POST", version+"/user/signup", nil)
		response := httptest.NewRecorder()
		router.SignUpHandler(response, request)
		if response.Code != 500 {
			t.Errorf("error got %d response code, want %d", response.Code, 500)
		}

	})

}

func TestVerifySignUpCodeHandler(t *testing.T) {
	router, db, version := SetUp(t)
	body := []byte(`{
		"name":"name",
		"email":"name@gmail.com",
		"password":"strongpass",
		"confirm_password":"strongpass"
	}`)
	request1 := httptest.NewRequest("POST", version+"/user/signup", bytes.NewBuffer(body))
	response1 := httptest.NewRecorder()
	router.SignUpHandler(response1, request1)
	if response1.Code != 200 {
		t.Errorf("error got %d response code, want %d", response1.Code, 200)
	}
	code, err := db.GetCodeByEmail("name@gmail.com")
	if err != nil {
		t.Error(err)
	}
	t.Run("verify code ", func(t *testing.T) {
		data := fmt.Sprintf(`{
			"email":"name@gmail.com",
			"code": %d
		}`, code)
		body = []byte(data)
		request2 := httptest.NewRequest("POST", version+"/user/signup/verify_email", bytes.NewBuffer(body))
		response2 := httptest.NewRecorder()
		router.VerifySignUpCodeHandler(response2, request2)
		if response2.Code != 200 {
			t.Errorf("error got %d response code, want %d", response2.Code, 200)
		}
	})

	t.Run("add empty code", func(t *testing.T) {
		request2 := httptest.NewRequest("POST", version+"/user/signup/verify_email", nil)
		response2 := httptest.NewRecorder()
		router.VerifySignUpCodeHandler(response2, request2)
		if response2.Code != 500 {
			t.Errorf("error got %d response code, want %d", response2.Code, 500)
		}
	})
}
