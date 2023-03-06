// package routes for API endpoints
package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rawdaGastan/grid3_auto_deployer/internal"
	"github.com/rawdaGastan/grid3_auto_deployer/models"
)

// Router struct holds db model and configurations
type Router struct {
	config *internal.Configuration
	db     models.DB
}

// NewRouter create new router with db
func NewRouter(config internal.Configuration, db models.DB) (r Router) {
	return Router{&config, db}
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

// WriteErrResponse wite error messages in api
func (router *Router) WriteErrResponse(w http.ResponseWriter, err error) {
	jsonErrRes, _ := json.Marshal(ErrorMsg{Error: err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write(jsonErrRes)
	log.Printf("write error response failed %v", err.Error())
}

// WriteMsgResponse write response messages for api
func (router *Router) WriteMsgResponse(w http.ResponseWriter, message string, data interface{}) {
	contentJSON, err := json.Marshal(ResponseMsg{Message: message, Data: data})
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(contentJSON)
	log.Printf("write message response failed %v", err.Error())
}
