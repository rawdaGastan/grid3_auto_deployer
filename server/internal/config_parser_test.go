package internal

import (
	"io/fs"
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
	t.Run("read config file ", func(t *testing.T) {
		dir := t.TempDir()
		configPath := dir + "/config.json"

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		data, err := ReadConfFile(configPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

	})

	t.Run("change permissions of file", func(t *testing.T) {
		dir := t.TempDir()
		configPath := dir + "/config.json"

		err := os.WriteFile(configPath, []byte(config), fs.FileMode(os.O_RDONLY))
		assert.NoError(t, err)

		data, err := ReadConfFile(configPath)
		assert.Error(t, err)
		assert.Empty(t, data)

	})

	t.Run("no file exists", func(t *testing.T) {

		err := os.WriteFile("./config.json", []byte(config), fs.FileMode(os.O_RDONLY))
		assert.Error(t, err)

		data, err := ReadConfFile("./config.json")
		assert.Error(t, err)
		assert.Empty(t, data)

	})

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
	"version": "v1",
	"salt": "salt"
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
	assert.Equal(t, got.Token, expected.Token)
	assert.Equal(t, got.Database, expected.Database)
	assert.Equal(t, got.Version, expected.Version)
}
