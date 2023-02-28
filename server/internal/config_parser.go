package internal

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Configuration struct {
	Server     Server     `json:"server"`
	MailSender MailSender `json:"mailSender"`
	Database   DB         `json:"database"`
	Token      JwtToken   `json:"token"`
}

type Server struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type MailSender struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DB struct {
	File string `json:"file"`
}

type JwtToken struct {
	Secret  string `json:"secret"`
	Timeout int    `json:"timeout"`
}

func ReadConfFile(path string) ([]byte, error) {
	confFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer confFile.Close()
	conf, err := ioutil.ReadAll(confFile)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

func ParseConf(conf []byte) (*Configuration, error) {
	myConf := Configuration{}
	err := json.Unmarshal(conf, &myConf)
	if err != nil {
		return &myConf, err
	}
	return &myConf, nil
}
