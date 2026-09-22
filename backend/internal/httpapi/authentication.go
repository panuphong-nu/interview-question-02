package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"example.com/it02-auth/backend/internal/domain"
)

// TokenVerifier checks a raw bearer token and returns its claims. Declaring
// the interface here, where it is consumed, keeps the transport layer testable
// with a stub and free of any particular token library.
type TokenVerifier interface {
	Verify(raw string) (domain.AuthenticatedUser, error)
}

type claimsKey struct{}

// authenticate rejects a request that does not carry a valid token and puts
// the verified claims in the context for the handler behind it.
//
// Only the signature and the standard claims are checked here. Whether the
// subject still names a real account is a question about application state,
// so the handler asks the service; this layer only decides whether the token
// itself can be trusted.
func authenticate(verifier TokenVerifier, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			raw, ok := bearerToken(request)
			if !ok {
				unauthenticated(writer, logger, "Bearer")
				return
			}

			claims, err := verifier.Verify(raw)
			if err != nil {
				// The reason is not passed to the client beyond the standard
				// "invalid_token": telling a caller exactly why a token was
				// refused helps them forge a better one.
				unauthenticated(writer, logger, `Bearer error="invalid_token"`)
				return
			}

			next.ServeHTTP(writer, request.WithContext(
				context.WithValue(request.Context(), claimsKey{}, claims),
			))
		})
	}
}

// claimsFrom returns the verified claims carried by a context. The second
// result is false outside an authenticated route.
func claimsFrom(ctx context.Context) (domain.AuthenticatedUser, bool) {
	claims, ok := ctx.Value(claimsKey{}).(domain.AuthenticatedUser)
	return claims, ok
}

// bearerToken extracts the credential from an Authorization header. The scheme
// is matched case insensitively, as RFC 7235 requires.
func bearerToken(request *http.Request) (string, bool) {
	header := request.Header.Get("Authorization")
	scheme, credential, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}

	credential = strings.TrimSpace(credential)
	if credential == "" {
		return "", false
	}
	return credential, true
}

// unauthenticated answers with 401 and the WWW-Authenticate header the status
// code requires, which is what tells a client how to authenticate.
func unauthenticated(writer http.ResponseWriter, logger *slog.Logger, challenge string) {
	writer.Header().Set("WWW-Authenticate", challenge)
	writeMessage(writer, logger, http.StatusUnauthorized, messageUnauthenticated)
}
