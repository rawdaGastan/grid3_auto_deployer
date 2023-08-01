// Package internal for internal details
package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/validator.v2"
)

// Configuration struct to hold app configurations
type Configuration struct {
	Server                      Server      `json:"server"`
	MailSender                  MailSender  `json:"mailSender"`
	Database                    DB          `json:"database"`
	Token                       JwtToken    `json:"token"`
	Account                     GridAccount `json:"account"`
	Version                     string      `json:"version" validate:"nonzero"`
	Admins                      []string    `json:"admins"`
	NotifyAdminsIntervalHours   int         `json:"notifyAdminsIntervalHours"`
	AdminSSHKey                 string      `json:"adminSSHKey"`
	BalanceThreshold            int         `json:"balanceThreshold"`
	ExpirationToleranceInDays   int         `json:"expirationToleranceInDays"`
	NotifyUsersExpirationInDays int         `json:"notifyUsersExpirationInDays"`
	Prices                      Price       `json:"prices"`
	StripeSecret                string      `json:"stripe_secret" validate:"nonzero"`
}

// Server struct to hold server's information
type Server struct {
	Host string `json:"host" validate:"nonzero"`
	Port string `json:"port" validate:"nonzero"`

	RedisHost string `json:"redisHost" validate:"nonzero"`
	RedisPort string `json:"redisPort" validate:"nonzero"`
	RedisPass string `json:"redisPass"`
}

// MailSender struct to hold sender's email, password
type MailSender struct {
	Email       string `json:"email" validate:"nonzero"`
	SendGridKey string `json:"sendgrid_key" validate:"nonzero"`
	Timeout     int    `json:"timeout" validate:"min=30"`
}

// DB struct to hold database file
type DB struct {
	File string `json:"file" validate:"nonzero"`
}

// JwtToken struct to hold JWT information
type JwtToken struct {
	Secret  string `json:"secret" validate:"nonzero"`
	Timeout int    `json:"timeout" validate:"min=5"`
}

// Price struct to hold prices info
type Price struct {
	SmallVM              uint64 `json:"small_vm" validate:"nonzero"`
	SmallVMWithPublicIP  uint64 `json:"small_vm_with_public_ip" validate:"nonzero"`
	MediumVM             uint64 `json:"medium_vm" validate:"nonzero"`
	MediumVMWithPublicIP uint64 `json:"medium_vm_with_public_ip" validate:"nonzero"`
	LargeVM              uint64 `json:"large_vm" validate:"nonzero"`
	LargeVMWithPublicIP  uint64 `json:"large_vm_with_public_ip" validate:"nonzero"`
}

// GridAccount struct to hold grid account mnemonics
type GridAccount struct {
	Mnemonics string `json:"mnemonics" validate:"nonzero"`
	Network   string `json:"network" validate:"nonzero"`
}

// ReadConfFile read configurations of json file
func ReadConfFile(path string) (Configuration, error) {
	config := Configuration{NotifyAdminsIntervalHours: 6, BalanceThreshold: 2000, ExpirationToleranceInDays: 30, NotifyUsersExpirationInDays: 1}
	file, err := os.Open(path)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to open config file: %w", err)
	}

	dec := json.NewDecoder(file)
	if err := dec.Decode(&config); err != nil {
		return Configuration{}, fmt.Errorf("failed to load config: %w", err)
	}

	return config, validator.Validate(config)
}
