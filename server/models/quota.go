package models

import "github.com/google/uuid"

type Quota struct {
	UserID uuid.UUID `json:"userID"`
	Vms    int       `json:"vms"`
	K8s    int       `json:"k8s"`
}
