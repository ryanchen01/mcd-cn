package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const tokenEnvKey = "MCDCN_MCP_TOKEN"

func LoadToken() (string, error) {
	if value := strings.TrimSpace(os.Getenv(tokenEnvKey)); value != "" {
		return value, nil
	}

	file, err := os.Open(".env")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%s not set and .env not found", tokenEnvKey)
		}
		return "", fmt.Errorf("read .env: %w", err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if key != "" {
			values[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan .env: %w", err)
	}

	if value := strings.TrimSpace(values[tokenEnvKey]); value != "" {
		return value, nil
	}

	return "", fmt.Errorf("%s not set in environment or .env", tokenEnvKey)
}
