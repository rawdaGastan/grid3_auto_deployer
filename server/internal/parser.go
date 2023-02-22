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

func ParseEnv(content []byte) (string, error) {
	envMap, err := env.Parse(strings.NewReader(string(content)))
	if err != nil {
		return "", err
	}

	secret := envMap["JWT_SECRET"]
	if secret == "" {
		return "", fmt.Errorf("error, secret is missing")
	}
	return secret, nil
}
