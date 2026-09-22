// Package security adapts a password hashing algorithm to the application's
// PasswordHasher port. It is the only package that knows which algorithm is in
// use, so replacing it is a change in one file plus one line of wiring.
package security

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"example.com/it02-auth/backend/internal/domain"
)

// Cost bounds. bcrypt's cost is the base 2 logarithm of the number of rounds,
// so each step doubles the work an attacker has to repeat per guess. The
// default is one step above the library's own, which has not moved in years.
const (
	DefaultCost = bcrypt.DefaultCost + 1
	MinCost     = bcrypt.MinCost
	MaxCost     = bcrypt.MaxCost
)

// BcryptHasher hashes passwords with bcrypt: a deliberately slow, salted
// algorithm. The salt is generated per password and stored inside the encoded
// hash, so the same password never produces the same record twice.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher returns a hasher at the given cost. A cost outside the range
// bcrypt supports falls back to DefaultCost rather than failing at the first
// registration, which is the safer behaviour for a misconfigured deployment.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < MinCost || cost > MaxCost {
		cost = DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash implements service.PasswordHasher.
func (h *BcryptHasher) Hash(password string) (domain.PasswordHash, error) {
	// bcrypt refuses an input longer than 72 bytes. The domain's password
	// maximum keeps every accepted password under that, so reaching this
	// error means a caller bypassed validation.
	encoded, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return domain.PasswordHash(encoded), nil
}

// Verify implements service.PasswordHasher. A wrong password is reported
// as false with no error; an error means the stored hash could not be read at
// all, which is an operational fault rather than a failed sign in.
func (h *BcryptHasher) Verify(hash domain.PasswordHash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return false, nil
	case errors.Is(err, bcrypt.ErrHashTooShort):
		return false, fmt.Errorf("stored hash is unusable: %w", err)
	default:
		var version bcrypt.HashVersionTooNewError
		if errors.As(err, &version) {
			return false, fmt.Errorf("stored hash is unusable: %w", err)
		}
		// A password longer than bcrypt accepts cannot match any stored hash.
		return false, nil
	}
}
