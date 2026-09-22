package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/it02-auth/backend/internal/config"
	"example.com/it02-auth/backend/internal/httpapi"
	"example.com/it02-auth/backend/internal/infrastructure/id"
	"example.com/it02-auth/backend/internal/infrastructure/security"
	"example.com/it02-auth/backend/internal/infrastructure/token"
	"example.com/it02-auth/backend/internal/repository/postgres"
	"example.com/it02-auth/backend/internal/service"
)

const (
	tokenIssuer   = "it02-auth"
	tokenAudience = "it02-web"
	tokenLifetime = time.Hour
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// run is the composition root: dependencies are created and connected here.
func run() error {
	configuration, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database, err := postgres.Open(ctx, configuration.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	signer := token.NewSigner(token.Options{
		Secret:   configuration.JWTSecret,
		Issuer:   tokenIssuer,
		Audience: tokenAudience,
		Lifetime: tokenLifetime,
	})
	auth := service.NewAuthService(
		postgres.NewUserRepository(database),
		security.NewBcryptHasher(security.DefaultCost),
		signer,
		id.NewUUID,
		time.Now,
	)

	server := &http.Server{
		Addr: configuration.HTTPAddress,
		Handler: httpapi.NewRouter(httpapi.Options{
			Auth:           auth,
			Verifier:       signer,
			Health:         postgres.NewHealthChecker(database),
			Logger:         slog.Default(),
			AllowedOrigins: configuration.AllowedOrigins,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go shutdownOnSignal(ctx, server)
	slog.Info("api listening", "address", configuration.HTTPAddress)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func shutdownOnSignal(ctx context.Context, server *http.Server) {
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
