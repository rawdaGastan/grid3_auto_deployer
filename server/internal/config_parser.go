// Package internal for internal details
package internal

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

// Configuration struct to hold app configurations
type Configuration struct {
	Server                    Server      `json:"server"`
	MailSender                MailSender  `json:"mailSender"`
	Database                  DB          `json:"database"`
	Token                     JwtToken    `json:"token"`
	Account                   GridAccount `json:"account"`
	Version                   string      `json:"version"`
	Salt                      string      `json:"salt"`
	Admins                    []string    `json:"admins"`
	NotifyAdminsIntervalHours int         `json:"notifyAdminsIntervalHours"`
}

// Server struct to hold server's information
type Server struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// MailSender struct to hold sender's email, password
type MailSender struct {
	Email       string `json:"email"`
	SendGridKey string `json:"sendgrid_key"`
	Timeout     int    `json:"timeout"`
}

// DB struct to hold database file
type DB struct {
	File string `json:"file"`
}

// JwtToken struct to hold JWT information
type JwtToken struct {
	Secret  string `json:"secret"`
	Timeout int    `json:"timeout"`
}

// GridAccount struct to hold grid account mnemonics
type GridAccount struct {
	Mnemonics string `json:"mnemonics"`
	Network   string `json:"network"`
}

// ReadConfFile read configurations of json file
func ReadConfFile(path string) ([]byte, error) {
	confFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer confFile.Close()
	conf, err := io.ReadAll(confFile)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

// ParseConf parses content of file to Configurations struct
func ParseConf(conf []byte) (Configuration, error) {
	var myConf Configuration
	err := json.Unmarshal(conf, &myConf)
	if err != nil {
		return myConf, err
	}

	if myConf.Server.Host == "" || myConf.Server.Port == "" {
		return myConf, errors.New("server configuration is required")
	}

	if myConf.MailSender.Email == "" || myConf.MailSender.SendGridKey == "" || myConf.MailSender.Timeout == 0 {
		return myConf, errors.New("mail sender configuration is required")
	}

	if myConf.Database.File == "" {
		return myConf, errors.New("database configuration is required")
	}

	if myConf.Account.Mnemonics == "" || myConf.Account.Network == "" {
		return myConf, errors.New("account configuration is required")
	}

	if myConf.Token.Secret == "" || myConf.Token.Timeout == 0 {
		return myConf, errors.New("jwt token configuration is required")
	}

	if myConf.Version == "" {
		return myConf, errors.New("version is required")
	}

	if myConf.Salt == "" {
		return myConf, errors.New("salt is required")
	}

	if myConf.NotifyAdminsIntervalHours == 0 {
		myConf.NotifyAdminsIntervalHours = 6
	}

	return myConf, nil
}
