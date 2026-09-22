// Package token adapts JSON Web Tokens to the application's TokenIssuer port
// and provides the matching verifier the HTTP layer uses to authenticate a
// request. It is the only package that knows the token format.
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/service"
)

// ErrInvalidToken is returned for every rejected token: a bad signature, an
// expired one, the wrong algorithm, a wrong issuer or audience, or a malformed
// string. The caller only ever needs to know that it cannot be trusted, and
// collapsing the reasons keeps that decision from leaking into a response.
var ErrInvalidToken = errors.New("invalid token")

// Signer both issues and verifies tokens with one HMAC secret. HS256 is a
// symmetric algorithm: the service that signs is the same service that
// verifies, which is exactly this program's situation. An asymmetric algorithm
// would only be needed if a separate service had to verify without being able
// to mint tokens of its own.
type Signer struct {
	secret   []byte
	issuer   string
	audience string
	lifetime time.Duration
}

// Options configures a Signer.
type Options struct {
	Secret   []byte
	Issuer   string
	Audience string
	Lifetime time.Duration
}

// NewSigner returns a Signer for the given options.
func NewSigner(options Options) *Signer {
	return &Signer{
		secret:   options.Secret,
		issuer:   options.Issuer,
		audience: options.Audience,
		lifetime: options.Lifetime,
	}
}

// customClaims is the on the wire shape. Registered claim names are used
// wherever one exists, so any standard JWT tool can read the token.
type customClaims struct {
	Username string `json:"preferred_username"`
	jwt.RegisteredClaims
}

// Issue implements service.TokenIssuer.
func (s *Signer) Issue(user domain.User, issuedAt time.Time) (service.IssuedToken, error) {
	expiresAt := issuedAt.Add(s.lifetime)

	claims := customClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			// A unique token identifier is what a future revocation list would
			// key on; it costs nothing to include now.
			ID: newTokenID(issuedAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return service.IssuedToken{}, fmt.Errorf("sign token: %w", err)
	}
	return service.IssuedToken{Value: signed, ExpiresAt: expiresAt}, nil
}

// Verify checks the signature and every time based and identity claim, then
// returns the payload. It wraps ErrInvalidToken for any failure.
func (s *Signer) Verify(raw string) (domain.AuthenticatedUser, error) {
	parsed, err := jwt.ParseWithClaims(raw, &customClaims{},
		func(*jwt.Token) (any, error) { return s.secret, nil },
		// Pinning the algorithm is what stops the classic attack of handing
		// the server a token signed with "none", or one signed with the public
		// half of an asymmetric key.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return domain.AuthenticatedUser{}, fmt.Errorf("%w: %s", ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(*customClaims)
	if !ok || claims.Subject == "" {
		return domain.AuthenticatedUser{}, fmt.Errorf("%w: missing subject", ErrInvalidToken)
	}
	return domain.AuthenticatedUser{UserID: claims.Subject, Username: claims.Username}, nil
}
