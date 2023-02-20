package routes

import "github.com/rawdaGastan/grid3_auto_deployer/models"

type Router struct { //TODO: HAndlers && http.Client
	db models.DB
}

func NewRouter(db models.DB) (r Router) {
	return Router{db}
}

type ErrorMsg struct {
	Message string `json:"message"`
}
