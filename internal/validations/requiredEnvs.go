package validations

import (
	"fmt"
	"os"
	"strings"
)

func RequiredEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s not found in environment variables", name)
	}
	return value, nil
}
