package models

type Quota struct {
	UserID uint64 `json:"userID"`
	Vms    int    `json:"vms"`
	K8s    int    `json:"k8s"`
}
