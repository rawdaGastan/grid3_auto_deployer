package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rightConfig = `
{
	"server": {
		"host": "localhost",
		"port": ":3000",
		"redisHost": "localhost",
		"redisPort": "6379",
		"redisPass": ""		
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
	"currency": "eur",
	"prices": {
		"public_ip": 2,
		"small_vm": 10,
		"medium_vm": 20,
		"large_vm": 30
	},
	"stripe_secret": "sk_test"
}
	`

func TestReadConfFile(t *testing.T) {
	t.Run("read config file ", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(rightConfig), 0644)
		assert.NoError(t, err)

		data, err := ReadConfFile(configPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

	})

	t.Run("change permissions of file", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(rightConfig), fs.FileMode(os.O_RDONLY))
		assert.NoError(t, err)

		data, err := ReadConfFile(configPath)
		assert.Error(t, err)
		assert.Empty(t, data)

	})

	t.Run("no file exists", func(t *testing.T) {
		data, err := ReadConfFile("./testing.json")
		assert.Error(t, err)
		assert.Empty(t, data)

	})

}

func TestParseConf(t *testing.T) {

	t.Run("can't unmarshal", func(t *testing.T) {
		config := `{testing}`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err)

	})

	t.Run("parse config file", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(rightConfig), 0644)
		assert.NoError(t, err)

		got, err := ReadConfFile(configPath)
		assert.NoError(t, err)

		expected := Configuration{
			Server: Server{
				Host:      "localhost",
				Port:      ":3000",
				RedisHost: "localhost",
				RedisPort: "6379",
				RedisPass: "",
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

		assert.NoError(t, err)
		assert.Equal(t, got.Server, expected.Server)
		assert.Equal(t, got.MailSender.Email, expected.MailSender.Email)
		assert.Equal(t, got.Account.Mnemonics, expected.Account.Mnemonics)
		assert.Equal(t, got.Token, expected.Token)
		assert.Equal(t, got.Database, expected.Database)
		assert.Equal(t, got.Version, expected.Version)
	})

	t.Run("no file", func(t *testing.T) {
		_, err := ReadConfFile("config.json")
		assert.Error(t, err)

	})

	t.Run("no server configuration", func(t *testing.T) {
		config :=
			`
{
	"server": {
		"host": "",
		"port": ""
	}
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "server configuration is required")
	})

	t.Run("no mail sender configuration", func(t *testing.T) {
		config :=
			`
{
	"server": {
		"host": "localhost",
		"port": ":3000"
	},
	"mailSender": {
        "email": "",
        "sendgrid_key": "",
        "timeout": 0
    }
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "mail sender configuration is required")

	})

	t.Run("no database configuration", func(t *testing.T) {
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
	"database": {
        "file": ""
    }
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "database configuration is required")

	})

	t.Run("no account configuration", func(t *testing.T) {
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
	"database": {
        "file": "testing.db"
    },
    "account": {
        "mnemonics": "",
		"network": ""
    }
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "account configuration is required")

	})

	t.Run("no jwt token configuration", func(t *testing.T) {
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
	"database": {
        "file": "testing.db"
    },
    "account": {
        "mnemonics": "my mnemonics",
		"network": "my network"
    },	
	"token": {
        "secret": "",
        "timeout": 0
    }
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "jwt token configuration is required")

	})

	t.Run("no version configuration", func(t *testing.T) {
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
	"database": {
        "file": "testing.db"
    },
	"token": {
        "secret": "secret",
        "timeout": 10
    },	
	"version": ""
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "version is required")

	})

	t.Run("no currency configuration", func(t *testing.T) {
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
	"database": {
        "file": "testing.db"
    },
	"token": {
        "secret": "secret",
        "timeout": 10
    },	
	"version": "v1",	
}
	`
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "currency is required")
	})

	t.Run("no prices configuration", func(t *testing.T) {
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
	"database": {
        "file": "testing.db"
    },
	"token": {
        "secret": "secret",
        "timeout": 10
    },	
	"version": "v1",	
	"currency": "eur",
}
	`
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "prices is required")
	})

	t.Run("no stripe secret configuration", func(t *testing.T) {
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
	"database": {
        "file": "testing.db"
    },
	"token": {
        "secret": "secret",
        "timeout": 10
    },	
	"version": "v1",	
	"currency": "eur",
	"prices": {
		"public_ip": 2,
		"small_vm": 10,
		"medium_vm": 20,
		"large_vm": 30
	},
}
	`
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "stripe_secret is required")

	})
}
