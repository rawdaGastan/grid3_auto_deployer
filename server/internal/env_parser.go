package internal

import (
	"fmt"
	"os"
	"strings"

	env "github.com/hashicorp/go-envparse"
)

func ReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}

	return content, nil
}

func ParseEnv(content []byte) (map[string]string, error) {
	m := make(map[string]string)
	envMap, err := env.Parse(strings.NewReader(string(content)))
	if err != nil {
		return m, err
	}

	secret := envMap["JWT_SECRET"]
	time := envMap["ExpirationTimeout"]
	sender := envMap["MAIL_SENDER"]
	password := envMap["PASSWORD_SENDER"]
	m["secret"] = secret
	m["time"] = time
	m["mailSender"] = sender
	m["senderPassword"] = password
	if secret == "" {
		return m, fmt.Errorf("error, secret is missing")
	}
	return m, nil
}
