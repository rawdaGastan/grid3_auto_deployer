// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rawdaGastan/cloud4students/internal"
	"github.com/rawdaGastan/cloud4students/models"
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

func setupCorsResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")

	if req.Method == "OPTIONS" {
		(*w).WriteHeader(http.StatusOK)
		return
	}
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
func (r *Router) WriteErrResponse(w http.ResponseWriter, err error) {
	jsonErrRes, _ := json.Marshal(ErrorMsg{Error: err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write(jsonErrRes)
	if err != nil {
		log.Printf("write error response failed %v", err.Error())
	}
}

// WriteMsgResponse write response messages for api
func (r *Router) WriteMsgResponse(w http.ResponseWriter, message string, data interface{}) {
	contentJSON, err := json.Marshal(ResponseMsg{Message: message, Data: data})
	if err != nil {
		r.WriteErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(contentJSON)
	if err != nil {
		r.WriteErrResponse(w, fmt.Errorf("write message response failed %v", err))
	}
}
