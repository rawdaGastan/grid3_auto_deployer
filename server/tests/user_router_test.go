package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
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
func SetUp(t testing.TB) (r *routes.Router, db models.DB, configurations *internal.Configuration, version string) {
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
	return &router, db, configuration, version

}

func TestSignUpHandler(t *testing.T) {
	router, _, _, version := SetUp(t)
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
		assert.Equal(t, response.Code, http.StatusOK)
	})

	t.Run("send empty data", func(t *testing.T) {
		request := httptest.NewRequest("POST", version+"/user/signup", nil)
		response := httptest.NewRecorder()
		router.SignUpHandler(response, request)
		assert.Equal(t, response.Code, http.StatusInternalServerError)
	})

}

func TestVerifySignUpCodeHandler(t *testing.T) {
	router, db, _, version := SetUp(t)
	body := []byte(`{
		"name":"name",
		"email":"name@gmail.com",
		"password":"strongpass",
		"confirm_password":"strongpass"
	}`)
	request1 := httptest.NewRequest("POST", version+"/user/signup", bytes.NewBuffer(body))
	response1 := httptest.NewRecorder()
	router.SignUpHandler(response1, request1)
	assert.Equal(t, response1.Code, http.StatusOK)

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
		assert.Equal(t, response2.Code, http.StatusOK)

	})

	t.Run("add empty code", func(t *testing.T) {
		request2 := httptest.NewRequest("POST", version+"/user/signup/verify_email", nil)
		response2 := httptest.NewRecorder()
		router.VerifySignUpCodeHandler(response2, request2)
		assert.Equal(t, response2.Code, http.StatusInternalServerError)

	})
}

func TestSignInHandler(t *testing.T) {
	router, db, _, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(u)
	if err != nil {
		t.Error(err)
	}

	t.Run("signIn successfully", func(t *testing.T) {
		body := []byte(`{
			"name":"name",
			"email":"name@gmail.com",
			"password":"strongpass"
		}`)
		request := httptest.NewRequest("POST", version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.SignInHandler(response, request)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	t.Run("signIn with wrong password", func(t *testing.T) {
		body := []byte(`{
			"name":"name",
			"email":"name@gmail.com",
			"password":"wrongpass"
		}`)
		request := httptest.NewRequest("POST", version+"/user/signin", bytes.NewBuffer(body))
		response := httptest.NewRecorder()
		router.SignInHandler(response, request)
		got := response.Body.String()
		want := `{"err":"Password is not correct"}`
		if got != want {
			t.Errorf("error got %q want %q", got, want)
		}
		assert.Equal(t, response.Code, http.StatusInternalServerError)

	})
}

func TestRefreshJWTHandler(t *testing.T) {
	router, db, config, version := SetUp(t)
	u := models.User{
		Name:           "name",
		Email:          "name@gmail.com",
		HashedPassword: "$2a$14$EJtkQHG54.wyFnBMBJn2lus5OkIZn3l/MtuqbaaX1U3KpttvxVGN6",
		Verified:       true,
	}
	err := db.CreateUser(u)
	if err != nil {
		t.Error(err)
	}

	t.Run("refresh jwt token", func(t *testing.T) {
		user, err := db.GetUserByEmail("name@gmail.com")
		if err != nil {
			t.Error(err)
		}
		token, err := internal.CreateJWT(user.ID.String(), user.Email, config.Token.Secret, config.Token.Timeout)
		if err != nil {
			t.Error(err)
		}
		request := httptest.NewRequest("POST", version+"/user/refresh_token", nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
		response := httptest.NewRecorder()
		router.RefreshJWTHandler(response, request)
		assert.Equal(t, response.Code, http.StatusOK)

	})

	

}
