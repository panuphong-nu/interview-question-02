package domain_test

import (
	"errors"
	"strings"
	"testing"

	"example.com/it02-auth/backend/internal/domain"
)

func TestNewRegistrationAcceptsValidInput(t *testing.T) {
	registration, err := domain.NewRegistration("  Somchai_01  ", "sup3r-secret", "sup3r-secret")
	if err != nil {
		t.Fatalf("NewRegistration() error = %v, want nil", err)
	}

	if registration.Username != "Somchai_01" {
		t.Errorf("Username = %q, want the trimmed form %q", registration.Username, "Somchai_01")
	}
	if registration.Canonical() != "somchai_01" {
		t.Errorf("Canonical() = %q, want %q", registration.Canonical(), "somchai_01")
	}
}

func TestNewRegistrationRejectsInvalidInput(t *testing.T) {
	tests := map[string]struct {
		username string
		password string
		confirm  string
		want     domain.Violation
	}{
		"empty username": {
			username: "   ", password: "sup3r-secret", confirm: "sup3r-secret",
			want: domain.Violation{Field: domain.FieldUsername, Code: domain.ViolationRequired},
		},
		"username too short": {
			username: "ab", password: "sup3r-secret", confirm: "sup3r-secret",
			want: domain.Violation{Field: domain.FieldUsername, Code: domain.ViolationMinLength, Limit: domain.UsernameMinLength},
		},
		"username too long": {
			username: strings.Repeat("a", domain.UsernameMaxLength+1), password: "sup3r-secret", confirm: "sup3r-secret",
			want: domain.Violation{Field: domain.FieldUsername, Code: domain.ViolationMaxLength, Limit: domain.UsernameMaxLength},
		},
		"username with a space": {
			username: "som chai", password: "sup3r-secret", confirm: "sup3r-secret",
			want: domain.Violation{Field: domain.FieldUsername, Code: domain.ViolationInvalidFormat},
		},
		"username with punctuation": {
			username: "somchai!", password: "sup3r-secret", confirm: "sup3r-secret",
			want: domain.Violation{Field: domain.FieldUsername, Code: domain.ViolationInvalidFormat},
		},
		"empty password": {
			username: "somchai", password: "", confirm: "",
			want: domain.Violation{Field: domain.FieldPassword, Code: domain.ViolationRequired},
		},
		"password too short": {
			username: "somchai", password: "short", confirm: "short",
			want: domain.Violation{Field: domain.FieldPassword, Code: domain.ViolationMinLength, Limit: domain.PasswordMinLength},
		},
		"password beyond the hasher's input limit": {
			username: "somchai",
			password: strings.Repeat("x", domain.PasswordMaxLength+1),
			confirm:  strings.Repeat("x", domain.PasswordMaxLength+1),
			want:     domain.Violation{Field: domain.FieldPassword, Code: domain.ViolationMaxLength, Limit: domain.PasswordMaxLength},
		},
		"confirmation does not match": {
			username: "somchai", password: "sup3r-secret", confirm: "sup3r-secrets",
			want: domain.Violation{Field: domain.FieldConfirmPassword, Code: domain.ViolationMismatch},
		},
		"confirmation missing": {
			username: "somchai", password: "sup3r-secret", confirm: "",
			want: domain.Violation{Field: domain.FieldConfirmPassword, Code: domain.ViolationRequired},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.NewRegistration(test.username, test.password, test.confirm)

			var validation *domain.ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("NewRegistration() error = %v, want a *ValidationError", err)
			}
			if !containsViolation(validation.Violations, test.want) {
				t.Errorf("violations = %+v, want one matching %+v", validation.Violations, test.want)
			}
		})
	}
}

// A form marks every bad field at once, so one call has to report them all.
func TestNewRegistrationReportsEveryViolationTogether(t *testing.T) {
	_, err := domain.NewRegistration("", "short", "different")

	var validation *domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("NewRegistration() error = %v, want a *ValidationError", err)
	}

	for _, want := range []domain.Violation{
		{Field: domain.FieldUsername, Code: domain.ViolationRequired},
		{Field: domain.FieldPassword, Code: domain.ViolationMinLength, Limit: domain.PasswordMinLength},
		{Field: domain.FieldConfirmPassword, Code: domain.ViolationMismatch},
	} {
		if !containsViolation(validation.Violations, want) {
			t.Errorf("violations = %+v, want one matching %+v", validation.Violations, want)
		}
	}
}

// Sign in checks presence only: an old password that no longer satisfies the
// current policy must still work.
func TestNewLoginAttemptChecksPresenceOnly(t *testing.T) {
	attempt, err := domain.NewLoginAttempt(" Somchai ", "old")
	if err != nil {
		t.Fatalf("NewLoginAttempt() error = %v, want nil", err)
	}
	if attempt.Username != "Somchai" || attempt.Canonical() != "somchai" {
		t.Errorf("attempt = %+v, want the trimmed and folded username", attempt)
	}
}

func TestNewLoginAttemptRejectsMissingFields(t *testing.T) {
	_, err := domain.NewLoginAttempt("", "")

	var validation *domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("NewLoginAttempt() error = %v, want a *ValidationError", err)
	}
	if len(validation.Violations) != 2 {
		t.Errorf("violations = %+v, want one for each empty field", validation.Violations)
	}
}

func TestCanonicalUsernameFoldsCase(t *testing.T) {
	if got := domain.CanonicalUsername("  SomChai  "); got != "somchai" {
		t.Errorf("CanonicalUsername() = %q, want %q", got, "somchai")
	}
}

func containsViolation(violations []domain.Violation, want domain.Violation) bool {
	for _, violation := range violations {
		if violation == want {
			return true
		}
	}
	return false
}
