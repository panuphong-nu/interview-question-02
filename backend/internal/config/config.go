package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	defaultHTTPAddress    = ":4201"
	defaultAllowedOrigins = "http://localhost:4200,http://127.0.0.1:4200"
	minJWTSecretLength    = 32
)

type Config struct {
	HTTPAddress    string
	DatabaseURL    string
	JWTSecret      []byte
	AllowedOrigins []string
}

func Load() (Config, error) {
	configuration := Config{
		HTTPAddress:    normalizeHTTPAddress(valueOr("HTTP_ADDRESS", defaultHTTPAddress)),
		DatabaseURL:    strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:      []byte(strings.TrimSpace(os.Getenv("JWT_SECRET"))),
		AllowedOrigins: splitList(valueOr("CORS_ALLOWED_ORIGINS", defaultAllowedOrigins)),
	}

	if len(configuration.JWTSecret) < minJWTSecretLength {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least %d bytes", minJWTSecretLength)
	}
	if err := validateDatabaseURL(configuration.DatabaseURL); err != nil {
		return Config{}, err
	}
	if len(configuration.AllowedOrigins) == 0 {
		return Config{}, fmt.Errorf("CORS_ALLOWED_ORIGINS must not be empty")
	}
	return configuration, nil
}

// normalizeHTTPAddress accepts the convenient port-only form used by secret
// managers (for example "4201") and converts it to the form net/http expects.
func normalizeHTTPAddress(address string) string {
	if _, err := strconv.ParseUint(address, 10, 16); err == nil {
		return ":" + address
	}
	return address
}

func validateDatabaseURL(connectionString string) error {
	parsed, err := url.Parse(connectionString)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" {
		return fmt.Errorf("DATABASE_URL must be a valid PostgreSQL connection string")
	}
	return nil
}

func valueOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}
