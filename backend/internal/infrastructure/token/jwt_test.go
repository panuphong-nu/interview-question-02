package token_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/infrastructure/token"
)

var (
	secret = []byte("a-test-secret-of-at-least-32-bytes!!")
	// Verification checks exp against the real clock, so a token under test
	// has to be minted relative to now rather than at a fixed date.
	issuedAt = time.Now().Truncate(time.Second)
	user     = domain.User{ID: "user-1", Username: "Somchai"}
)

func newSigner() *token.Signer {
	return token.NewSigner(token.Options{
		Secret:   secret,
		Issuer:   "it02-auth",
		Audience: "it02-web",
		Lifetime: time.Hour,
	})
}

func TestIssuedTokenVerifies(t *testing.T) {
	signer := newSigner()

	issued, err := signer.Issue(user, issuedAt)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
	if want := issuedAt.Add(time.Hour); !issued.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", issued.ExpiresAt, want)
	}

	claims, err := signer.Verify(issued.Value)
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("UserID = %q, want %q", claims.UserID, user.ID)
	}
	if claims.Username != user.Username {
		t.Errorf("Username = %q, want %q", claims.Username, user.Username)
	}
}

// The payload is signed, not encrypted, so anyone holding a token can read it.
// This pins that nothing secret is put in there.
func TestTokenPayloadCarriesNoSecret(t *testing.T) {
	signer := newSigner()

	issued, err := signer.Issue(domain.User{
		ID:           "user-1",
		Username:     "Somchai",
		PasswordHash: "$2a$10$notarealhash",
	}, issuedAt)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	parts := strings.Split(issued.Value, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d parts, want the three of a JWS", len(parts))
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if strings.Contains(string(raw), "notarealhash") {
		t.Errorf("payload = %s, which carries the password hash", raw)
	}
	for _, claim := range []string{"sub", "iss", "aud", "exp", "iat", "jti"} {
		if _, ok := payload[claim]; !ok {
			t.Errorf("payload is missing the %q claim", claim)
		}
	}
}

func TestVerifyRejectsAnExpiredToken(t *testing.T) {
	signer := newSigner()

	// Issued far enough in the past that the lifetime, and the library's
	// small leeway for clock skew, have both elapsed.
	issued, err := signer.Issue(user, time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := signer.Verify(issued.Value); !errors.Is(err, token.ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyRejectsATokenSignedWithAnotherSecret(t *testing.T) {
	other := token.NewSigner(token.Options{
		Secret:   []byte("a-different-secret-of-at-least-32-bytes"),
		Issuer:   "it02-auth",
		Audience: "it02-web",
		Lifetime: time.Hour,
	})

	issued, err := other.Issue(user, issuedAt)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := newSigner().Verify(issued.Value); !errors.Is(err, token.ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
	}
}

// The "alg: none" forgery: a token whose signature is simply omitted. It is
// refused because the verifier pins HS256 rather than trusting the header.
func TestVerifyRejectsAnUnsignedToken(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(
		`{"sub":"user-1","iss":"it02-auth","aud":"it02-web","exp":4102444800}`,
	))
	forged := header + "." + claims + "."

	if _, err := newSigner().Verify(forged); !errors.Is(err, token.ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
	}
}

// A token minted by, or for, another service must not be accepted here even
// when the secret happens to match.
func TestVerifyRejectsAForeignIssuerOrAudience(t *testing.T) {
	tests := map[string]token.Options{
		"another issuer":   {Secret: secret, Issuer: "somewhere-else", Audience: "it02-web", Lifetime: time.Hour},
		"another audience": {Secret: secret, Issuer: "it02-auth", Audience: "another-app", Lifetime: time.Hour},
	}

	for name, options := range tests {
		t.Run(name, func(t *testing.T) {
			issued, err := token.NewSigner(options).Issue(user, issuedAt)
			if err != nil {
				t.Fatalf("Issue() error = %v", err)
			}

			if _, err := newSigner().Verify(issued.Value); !errors.Is(err, token.ErrInvalidToken) {
				t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
			}
		})
	}
}

func TestVerifyRejectsMalformedInput(t *testing.T) {
	signer := newSigner()

	for name, raw := range map[string]string{
		"empty":        "",
		"not a jwt":    "definitely-not-a-token",
		"two segments": "aaa.bbb",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := signer.Verify(raw); !errors.Is(err, token.ErrInvalidToken) {
				t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
			}
		})
	}
}

// Two tokens for the same account at the same instant still differ, so a
// future revocation list can key on the identifier.
func TestIssuedTokensCarryDistinctIdentifiers(t *testing.T) {
	signer := newSigner()

	first, err := signer.Issue(user, issuedAt)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	second, err := signer.Issue(user, issuedAt)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if first.Value == second.Value {
		t.Error("two tokens issued at the same instant are identical")
	}
}
