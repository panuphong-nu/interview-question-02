package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"slices"

	"example.com/it02-auth/backend/internal/service"
)

type HealthChecker interface {
	Check(context.Context) error
}

// Options are the collaborators the HTTP surface needs. A struct rather than a
// positional list keeps the call site readable as routes are added.
type Options struct {
	Auth           *service.AuthService
	Verifier       TokenVerifier
	Health         HealthChecker
	Logger         *slog.Logger
	AllowedOrigins []string
}

func NewRouter(options Options) http.Handler {
	auth := newAuthHandler(options.Auth, options.Logger)
	health := newHealthHandler(options.Health, options.Logger)
	requireToken := authenticate(options.Verifier, options.Logger)

	router := http.NewServeMux()
	router.HandleFunc("GET /health", health.check)
	router.HandleFunc("POST /api/auth/register", auth.register)
	router.HandleFunc("POST /api/auth/login", auth.login)
	router.Handle("GET /api/auth/me", requireToken(http.HandlerFunc(auth.me)))

	return withCORS(router, options.AllowedOrigins)
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if slices.Contains(allowedOrigins, origin) {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Vary", "Origin")
		}

		if request.Method == http.MethodOptions {
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
