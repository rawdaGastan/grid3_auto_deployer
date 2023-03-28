package internal

import (
	"os"
	"testing"
)

func TestReadConfFile(t *testing.T) {
	config :=
		`
{
	"server": {
		"host": "localhost",
		"port": ":3000"
	}
}
	`
	dir := t.TempDir()
	configPath := dir + "/config.json"
	os.WriteFile(configPath, []byte(config), 0644)
	data, err := ReadConfFile(configPath)
	if err != nil {
		t.Error(err)
	}
	if data == nil {
		t.Errorf("File is empty!")
	}
}

func TestParseConf(t *testing.T) {
	config :=
		`
{
	"server": {
		"host": "localhost",
		"port": ":3000"
	},
	"mailSender": {
        "email": "email",
        "sendgrid_key": "my sendgrid_key",
        "timeout": 60 
    },
    "account": {
        "mnemonics": "my mnemonics"
    }
}
	`
	dir := t.TempDir()
	configPath := dir + "/config.json"
	os.WriteFile(configPath, []byte(config), 0644)
	data, err := ReadConfFile(configPath)
	if err != nil {
		t.Error(err)
	}
	expected := Configuration{
		Server: Server{
			Host: "localhost",
			Port: ":3000",
		},
		MailSender: MailSender{
			Email:       "email",
			SendGridKey: "my sendgrid_key",
			Timeout:     60,
		},
		Account: GridAccount{
			Mnemonics: "my mnemonics",
		},
	}
	got, err := ParseConf(data)
	if err != nil {
		t.Error(err)
	}
	if got.Server != expected.Server {
		t.Errorf("incorrect data, got %v, want %v", got.Server, expected.Server)
	}
	if got.MailSender.Email != expected.MailSender.Email {
		t.Errorf("incorrect data, got %v, want %v", got.MailSender.Email, expected.MailSender.Email)
	}
	if got.Account.Mnemonics != expected.Account.Mnemonics {
		t.Errorf("incorrect data, got %s, want %s", got.Account.Mnemonics, expected.Account.Mnemonics)
	}

}
