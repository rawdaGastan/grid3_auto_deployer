package routes

import (
	"encoding/json"
	"net/http"

	"github.com/rawdaGastan/grid3_auto_deployer/internal"
	"github.com/rawdaGastan/grid3_auto_deployer/models"
)

type Router struct {
	config *internal.Configuration
	db     models.DB
}

func NewRouter(config internal.Configuration, db models.DB) (r Router) {
	return Router{&config, db}
}

type ErrorMsg struct {
	Error string `json:"err"`
}

type ResponeMsg struct {
	Message string      `json:"msg"`
	Data    interface{} `json:"data","omitempty"`
}

func (router *Router) WriteErrResponse(w http.ResponseWriter, err error) {
	jsonErrRes, _ := json.Marshal(ErrorMsg{Error: err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonErrRes)
}

func (router *Router) WriteMsgResponse(w http.ResponseWriter, message string, data interface{}) {
	contentJson, err := json.Marshal(ResponeMsg{Message: message, Data: data})
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contentJson)
}
