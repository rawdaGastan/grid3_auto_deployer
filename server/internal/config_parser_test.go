package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

	err := os.WriteFile(configPath, []byte(config), 0644)
	assert.NoError(t, err)

	data, err := ReadConfFile(configPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
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
        "mnemonics": "my mnemonics",
		"network": "my network"
    },
	"token": {
        "secret": "secret",
        "timeout": 10
    },
	"database": {
        "file": "testing.db"
    },
	"version": "v1"
}
	`
	dir := t.TempDir()
	configPath := dir + "/config.json"

	err := os.WriteFile(configPath, []byte(config), 0644)
	assert.NoError(t, err)

	data, err := ReadConfFile(configPath)
	assert.NoError(t, err)

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
		Token: JwtToken{
			Secret:  "secret",
			Timeout: 10,
		},
		Database: DB{
			File: "testing.db",
		},
		Version: "v1",
	}

	got, err := ParseConf(data)
	assert.NoError(t, err)
	assert.Equal(t, got.Server, expected.Server)
	assert.Equal(t, got.MailSender.Email, expected.MailSender.Email)
	assert.Equal(t, got.Account.Mnemonics, expected.Account.Mnemonics)
}
