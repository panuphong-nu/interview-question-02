package config_test

import (
	"testing"

	"example.com/it02-auth/backend/internal/config"
)

const (
	validSecret      = "a-test-secret-of-at-least-32-bytes!!"
	validDatabaseURL = "postgresql://postgres:secret@localhost:5432/postgres?sslmode=disable"
)

func TestLoad(t *testing.T) {
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("DATABASE_URL", validDatabaseURL)

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.HTTPAddress != ":4201" {
		t.Errorf("HTTPAddress = %q, want :4201", loaded.HTTPAddress)
	}
	if len(loaded.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins = %q, want two local origins", loaded.AllowedOrigins)
	}
}

func TestLoadNormalizesPortOnlyHTTPAddress(t *testing.T) {
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("DATABASE_URL", validDatabaseURL)
	t.Setenv("HTTP_ADDRESS", "4201")

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.HTTPAddress != ":4201" {
		t.Errorf("HTTPAddress = %q, want :4201", loaded.HTTPAddress)
	}
}

func TestLoadRequiresSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", validDatabaseURL)
	if _, err := config.Load(); err == nil {
		t.Fatal("Load() error = nil, want missing JWT_SECRET error")
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("JWT_SECRET", validSecret)
	if _, err := config.Load(); err == nil {
		t.Fatal("Load() error = nil, want missing DATABASE_URL error")
	}
}
