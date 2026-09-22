// Package domain holds the entities and the rules that are true regardless of
// how the program is delivered. Nothing here imports a database driver, an
// HTTP package or a hashing library: those are decisions of the outer layers.
package domain

import (
	"strings"
	"time"
	"unicode"
)

// Credential policy. The limits are the application's own rules rather than a
// storage detail; the password maximum also keeps every secret inside the 72
// byte input bcrypt accepts, so no password is ever silently truncated.
const (
	UsernameMinLength = 3
	UsernameMaxLength = 32
	PasswordMinLength = 8
	PasswordMaxLength = 72
)

// User is a registered account. The plaintext password is never a field: the
// entity only ever carries the encoded hash, so there is no representation in
// which a readable password can reach a log or a response body.
type User struct {
	ID string
	// Username is kept as the person typed it, so IT 02-3 can greet them the
	// way they wrote their name.
	Username string
	// UsernameCanonical is the case folded form used for uniqueness and
	// lookup, so "Somchai" and "somchai" cannot both be registered.
	UsernameCanonical string
	PasswordHash      PasswordHash
	CreatedAt         time.Time
}

// AuthenticatedUser is the identity proven by an authentication mechanism.
// It lives in the domain so the HTTP layer does not need to import the JWT
// adapter just to read a verified user ID.
type AuthenticatedUser struct {
	UserID   string
	Username string
}

// PasswordHash is an already encoded secret. It is a distinct type so a
// plaintext password cannot be assigned to it by accident, and it has no
// String method, which keeps it out of formatted output by default.
type PasswordHash string

// Registration is a validated sign up request: a username that satisfies every
// rule together with the plaintext password that is about to be hashed. It is
// the only way a caller can express "these inputs are good", so no unchecked
// value can reach the service layer.
type Registration struct {
	Username string
	// Password is plaintext and exists for exactly as long as it takes the
	// application layer to hand it to a hasher.
	Password string
}

// Canonical returns the case folded username of this registration.
func (r Registration) Canonical() string { return CanonicalUsername(r.Username) }

// NewRegistration applies every sign up rule at once and reports all of the
// failures together, which is what a form needs in order to mark each offending
// field in a single round trip.
func NewRegistration(username, password, confirmPassword string) (Registration, error) {
	username = strings.TrimSpace(username)

	var broken violations
	checkUsername(&broken, username)
	checkPassword(&broken, password)

	switch {
	case confirmPassword == "":
		broken.add(FieldConfirmPassword, ViolationRequired)
	case password != confirmPassword:
		// The mismatch is reported against the confirmation field because that
		// is the one the form asks the person to correct.
		broken.add(FieldConfirmPassword, ViolationMismatch)
	}

	if len(broken) > 0 {
		return Registration{}, &ValidationError{Violations: broken}
	}
	return Registration{Username: username, Password: password}, nil
}

// LoginAttempt is a well formed sign in request. Sign in checks only that the
// fields are present: applying the full policy here would tell an attacker
// which stored passwords are short, and would lock out accounts created before
// a rule was tightened.
type LoginAttempt struct {
	Username string
	Password string
}

// Canonical returns the case folded username of this attempt.
func (a LoginAttempt) Canonical() string { return CanonicalUsername(a.Username) }

// NewLoginAttempt validates presence only, for the reason given on LoginAttempt.
func NewLoginAttempt(username, password string) (LoginAttempt, error) {
	username = strings.TrimSpace(username)

	var broken violations
	if username == "" {
		broken.add(FieldUsername, ViolationRequired)
	}
	if password == "" {
		broken.add(FieldPassword, ViolationRequired)
	}

	if len(broken) > 0 {
		return LoginAttempt{}, &ValidationError{Violations: broken}
	}
	return LoginAttempt{Username: username, Password: password}, nil
}

// CanonicalUsername folds a username to the form used for uniqueness and
// lookup. Every layer that compares usernames goes through this function, so
// the rule cannot drift between the repository and the service.
func CanonicalUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func checkUsername(broken *violations, username string) {
	switch {
	case username == "":
		broken.add(FieldUsername, ViolationRequired)
		return
	case len([]rune(username)) < UsernameMinLength:
		broken.addLength(FieldUsername, ViolationMinLength, UsernameMinLength)
		return
	case len([]rune(username)) > UsernameMaxLength:
		broken.addLength(FieldUsername, ViolationMaxLength, UsernameMaxLength)
		return
	}

	// Letters, digits and a small set of separators. The allowed set is
	// deliberately narrow: a username is an identifier, and keeping spaces and
	// punctuation out of it removes a whole class of lookalike accounts.
	for _, symbol := range username {
		if unicode.IsLetter(symbol) || unicode.IsDigit(symbol) {
			continue
		}
		if symbol == '.' || symbol == '_' || symbol == '-' {
			continue
		}
		broken.add(FieldUsername, ViolationInvalidFormat)
		return
	}
}

func checkPassword(broken *violations, password string) {
	switch {
	case password == "":
		broken.add(FieldPassword, ViolationRequired)
	case len(password) < PasswordMinLength:
		// Measured in bytes, matching the limit the hasher works in.
		broken.addLength(FieldPassword, ViolationMinLength, PasswordMinLength)
	case len(password) > PasswordMaxLength:
		broken.addLength(FieldPassword, ViolationMaxLength, PasswordMaxLength)
	}
}
