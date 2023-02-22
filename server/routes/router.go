package routes

import (
	"encoding/json"
	"net/http"

	"github.com/rawdaGastan/grid3_auto_deployer/models"
)

type Router struct {
	secret string
	db     models.DB
}

func NewRouter(secret string, db models.DB) (r Router) {
	return Router{secret, db}
}

type ErrorMsg struct {
	Message string `json:"message"`
}

func (router *Router) WriteErrResponse(w http.ResponseWriter, err error) {
	jsonErrRes, _ := json.Marshal(ErrorMsg{Message: err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonErrRes)
}

func (router *Router) WriteMsgResponse(w http.ResponseWriter, content interface{}) {
	contentJson, err := json.Marshal(content)
	if err != nil {
		router.WriteErrResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contentJson)
}
