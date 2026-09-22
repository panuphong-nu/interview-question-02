package security_test

import (
	"strings"
	"testing"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/infrastructure/security"
)

// testCost keeps the suite fast. Production uses security.DefaultCost, which
// is deliberately slow and would add seconds to every test that hashes.
const testCost = security.MinCost

func TestHashVerifiesTheOriginalPassword(t *testing.T) {
	hasher := security.NewBcryptHasher(testCost)

	hash, err := hasher.Hash("sup3r-secret")
	if err != nil {
		t.Fatalf("Hash() error = %v, want nil", err)
	}

	matches, err := hasher.Verify(hash, "sup3r-secret")
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
	if !matches {
		t.Error("Verify() = false, want true for the password that produced the hash")
	}
}

func TestHashIsNotReversible(t *testing.T) {
	hasher := security.NewBcryptHasher(testCost)

	hash, err := hasher.Hash("sup3r-secret")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if strings.Contains(string(hash), "sup3r-secret") {
		t.Errorf("Hash() = %q, which contains the plaintext password", hash)
	}
	if !strings.HasPrefix(string(hash), "$2") {
		t.Errorf("Hash() = %q, want the bcrypt modular crypt format", hash)
	}
}

// A per password salt is what stops one leaked table being cracked in bulk.
func TestHashIsSaltedPerPassword(t *testing.T) {
	hasher := security.NewBcryptHasher(testCost)

	first, err := hasher.Hash("sup3r-secret")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	second, err := hasher.Hash("sup3r-secret")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if first == second {
		t.Error("the same password hashed twice produced the same value, so it is unsalted")
	}
	for _, hash := range []domain.PasswordHash{first, second} {
		matches, err := hasher.Verify(hash, "sup3r-secret")
		if err != nil || !matches {
			t.Errorf("Verify(%q) = %v, %v; both hashes must verify", hash, matches, err)
		}
	}
}

// A wrong password is an ordinary false, never an error: the caller decides
// what to do with it, and an error would be logged as a fault.
func TestVerifyRejectsAWrongPasswordWithoutAnError(t *testing.T) {
	hasher := security.NewBcryptHasher(testCost)

	hash, err := hasher.Hash("sup3r-secret")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	matches, err := hasher.Verify(hash, "not-the-password")
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil for a simple mismatch", err)
	}
	if matches {
		t.Error("Verify() = true, want false for a different password")
	}
}

// A corrupt column is an operational fault, and reporting it as a plain
// mismatch would hide a broken database behind a stream of failed sign ins.
func TestVerifyReportsAnUnusableStoredHash(t *testing.T) {
	hasher := security.NewBcryptHasher(testCost)

	matches, err := hasher.Verify("not-a-bcrypt-hash", "sup3r-secret")
	if err == nil {
		t.Fatal("Verify() error = nil, want an error for a corrupt stored hash")
	}
	if matches {
		t.Error("Verify() = true, want false")
	}
}

func TestNewBcryptHasherFallsBackOnAnUnusableCost(t *testing.T) {
	// Rather than failing at the first registration, a misconfigured cost
	// falls back to the safe default.
	hasher := security.NewBcryptHasher(security.MaxCost + 1)

	hash, err := hasher.Hash("sup3r-secret")
	if err != nil {
		t.Fatalf("Hash() error = %v, want the hasher to fall back to the default cost", err)
	}
	if hash == "" {
		t.Error("Hash() = empty, want a hash")
	}
}
