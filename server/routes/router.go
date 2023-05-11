// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/codescalers/cloud4students/validators"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"gopkg.in/validator.v2"
)

const internalServerErrorMsg = "Something went wrong"

// Router struct holds db model and configurations
type Router struct {
	config         *internal.Configuration
	db             models.DB
	redis          streams.RedisClient
	tfPluginClient deployer.TFPluginClient

	vmDeployed  bool
	k8sDeployed bool

	// response is a map of deployment name as a key and (error code and error if exists) as a value
	vmRequestResponse  map[string]streams.ErrResponse
	k8sRequestResponse map[string]streams.ErrResponse

	mutex sync.Mutex
}

// NewRouter create new router with db
func NewRouter(config internal.Configuration, db models.DB, redis streams.RedisClient, tfPluginClient deployer.TFPluginClient) (Router, error) {
	// validations
	err := validator.SetValidationFunc("ssh", validators.ValidateSSHKey)
	if err != nil {
		return Router{}, err
	}
	err = validator.SetValidationFunc("password", validators.ValidatePassword)
	if err != nil {
		return Router{}, err
	}
	err = validator.SetValidationFunc("mail", validators.ValidateMail)
	if err != nil {
		return Router{}, err
	}

	return Router{
		&config,
		db,
		redis,
		tfPluginClient,
		false,
		false,
		map[string]streams.ErrResponse{},
		map[string]streams.ErrResponse{},
		sync.Mutex{},
	}, nil
}

// ErrorMsg holds errors
type ErrorMsg struct {
	Error string `json:"err"`
}

// ResponseMsg holds messages and needed data
type ResponseMsg struct {
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

// writeErrResponse write error messages in api
func writeErrResponse(r *http.Request, w http.ResponseWriter, statusCode int, errStr string) {
	jsonErrRes, _ := json.Marshal(ErrorMsg{Error: errStr})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err := w.Write(jsonErrRes)
	if err != nil {
		log.Error().Err(err).Msg("write error response failed")
	}
	middlewares.Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(statusCode)).Inc()
}

// writeMsgResponse write response messages for api
func writeMsgResponse(r *http.Request, w http.ResponseWriter, message string, data interface{}) {
	contentJSON, err := json.Marshal(ResponseMsg{Message: message, Data: data})
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(r, w, http.StatusInternalServerError, internalServerErrorMsg)
		middlewares.Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(http.StatusInternalServerError)).Inc()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(contentJSON)
	if err != nil {
		log.Error().Err(err).Msg("write error response failed")
		writeErrResponse(r, w, http.StatusInternalServerError, internalServerErrorMsg)
		middlewares.Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(http.StatusInternalServerError)).Inc()
		return
	}

	middlewares.Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(http.StatusOK)).Inc()
}
