// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"net/http"

	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/models"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/grid3-go/deployer"
)

const internalServerErrorMsg = "Something Went Wrong"

// Router struct holds db model and configurations
type Router struct {
	config         *internal.Configuration
	db             models.DB
	tfPluginClient deployer.TFPluginClient
}

// NewRouter create new router with db
func NewRouter(config internal.Configuration, db models.DB, tfPluginClient deployer.TFPluginClient) (r Router) {
	return Router{&config, db, tfPluginClient}
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
func writeErrResponse(w http.ResponseWriter, statusCode int, errStr string) {
	jsonErrRes, _ := json.Marshal(ErrorMsg{Error: errStr})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err := w.Write(jsonErrRes)
	if err != nil {
		log.Error().Err(err).Msg("write error response failed")
	}
}

// writeMsgResponse write response messages for api
func writeMsgResponse(w http.ResponseWriter, message string, data interface{}) {
	contentJSON, err := json.Marshal(ResponseMsg{Message: message, Data: data})
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(contentJSON)
	if err != nil {
		log.Error().Err(err).Msg("write error response failed")
		writeErrResponse(w, http.StatusInternalServerError, internalServerErrorMsg)
	}
}
