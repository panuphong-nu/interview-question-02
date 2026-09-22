package service_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/service"
)

var (
	fixedNow    = time.Date(2026, time.September, 21, 10, 0, 0, 0, time.UTC)
	errUpstream = errors.New("database unreachable")
)

func TestRegisterStoresOnlyTheHashedPassword(t *testing.T) {
	users := newFakeRepository()
	service := newService(users, newFakeHasher(), newFakeIssuer())

	registration, err := domain.NewRegistration("Somchai", "sup3r-secret", "sup3r-secret")
	if err != nil {
		t.Fatalf("NewRegistration() error = %v", err)
	}

	user, err := service.Register(context.Background(), registration)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if user.ID != "generated-id" {
		t.Errorf("ID = %q, want the generated identifier", user.ID)
	}
	if user.Username != "Somchai" {
		t.Errorf("Username = %q, want it stored as typed", user.Username)
	}
	if user.UsernameCanonical != "somchai" {
		t.Errorf("UsernameCanonical = %q, want the folded form", user.UsernameCanonical)
	}
	if !user.CreatedAt.Equal(fixedNow) {
		t.Errorf("CreatedAt = %v, want the injected clock's time %v", user.CreatedAt, fixedNow)
	}

	stored := users.byCanonical["somchai"]
	if strings.Contains(string(stored.PasswordHash), "sup3r-secret") {
		t.Fatalf("PasswordHash = %q, the plaintext password must never be stored", stored.PasswordHash)
	}
	if stored.PasswordHash != fakeHash("sup3r-secret") {
		t.Errorf("PasswordHash = %q, want the hasher's output", stored.PasswordHash)
	}
}

func TestRegisterRejectsATakenUsernameRegardlessOfCase(t *testing.T) {
	users := newFakeRepository()
	users.put(domain.User{ID: "existing", Username: "somchai", UsernameCanonical: "somchai"})
	service := newService(users, newFakeHasher(), newFakeIssuer())

	registration, err := domain.NewRegistration("SOMCHAI", "sup3r-secret", "sup3r-secret")
	if err != nil {
		t.Fatalf("NewRegistration() error = %v", err)
	}

	if _, err := service.Register(context.Background(), registration); !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("Register() error = %v, want ErrUsernameTaken", err)
	}
}

// The pre-insert lookup can be won by a concurrent registration, so the
// repository's own conflict has to surface unchanged.
func TestRegisterSurfacesTheRepositoryConflict(t *testing.T) {
	users := newFakeRepository()
	users.createErr = domain.ErrUsernameTaken
	service := newService(users, newFakeHasher(), newFakeIssuer())

	registration, _ := domain.NewRegistration("somchai", "sup3r-secret", "sup3r-secret")

	if _, err := service.Register(context.Background(), registration); !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("Register() error = %v, want ErrUsernameTaken", err)
	}
}

func TestRegisterWrapsAnUnexpectedLookupFailure(t *testing.T) {
	users := newFakeRepository()
	users.findErr = errUpstream
	service := newService(users, newFakeHasher(), newFakeIssuer())

	registration, _ := domain.NewRegistration("somchai", "sup3r-secret", "sup3r-secret")

	_, err := service.Register(context.Background(), registration)
	if !errors.Is(err, errUpstream) {
		t.Fatalf("Register() error = %v, want it to wrap %v", err, errUpstream)
	}
	// An infrastructure failure must not be mistaken for a taken username.
	if errors.Is(err, domain.ErrUsernameTaken) {
		t.Error("Register() reported a conflict for what was a transport failure")
	}
}

func TestLoginIssuesATokenForCorrectCredentials(t *testing.T) {
	users := newFakeRepository()
	users.put(domain.User{
		ID:                "user-1",
		Username:          "Somchai",
		UsernameCanonical: "somchai",
		PasswordHash:      fakeHash("sup3r-secret"),
	})
	service := newService(users, newFakeHasher(), newFakeIssuer())

	attempt, err := domain.NewLoginAttempt("SOMCHAI", "sup3r-secret")
	if err != nil {
		t.Fatalf("NewLoginAttempt() error = %v", err)
	}

	session, err := service.Login(context.Background(), attempt)
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if session.Token.Value != "token-for:user-1" {
		t.Errorf("Token.Value = %q, want the issuer's output for that account", session.Token.Value)
	}
	if session.User.Username != "Somchai" {
		t.Errorf("User.Username = %q, want the stored display name", session.User.Username)
	}
	if !session.Token.ExpiresAt.After(fixedNow) {
		t.Errorf("ExpiresAt = %v, want a time after the injected now %v", session.Token.ExpiresAt, fixedNow)
	}
}

// Both failures answer identically, so a caller cannot use the error to
// discover which usernames exist.
func TestLoginReportsTheSameErrorForAWrongPasswordAndAnUnknownUser(t *testing.T) {
	users := newFakeRepository()
	users.put(domain.User{ID: "user-1", UsernameCanonical: "somchai", PasswordHash: fakeHash("sup3r-secret")})
	service := newService(users, newFakeHasher(), newFakeIssuer())

	wrongPassword, _ := domain.NewLoginAttempt("somchai", "not-the-password")
	unknownUser, _ := domain.NewLoginAttempt("nobody", "sup3r-secret")

	for name, attempt := range map[string]domain.LoginAttempt{
		"wrong password": wrongPassword,
		"unknown user":   unknownUser,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := service.Login(context.Background(), attempt); !errors.Is(err, domain.ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
			}
		})
	}
}

func TestLoginWrapsAFailingVerifier(t *testing.T) {
	users := newFakeRepository()
	users.put(domain.User{ID: "user-1", UsernameCanonical: "somchai", PasswordHash: "corrupt"})
	hasher := newFakeHasher()
	hasher.verifyErr = errUpstream
	service := newService(users, hasher, newFakeIssuer())

	attempt, _ := domain.NewLoginAttempt("somchai", "sup3r-secret")

	_, err := service.Login(context.Background(), attempt)
	if !errors.Is(err, errUpstream) {
		t.Fatalf("Login() error = %v, want it to wrap %v", err, errUpstream)
	}
	// An unreadable stored hash is a fault, not a rejected sign in.
	if errors.Is(err, domain.ErrInvalidCredentials) {
		t.Error("Login() reported bad credentials for what was a hasher failure")
	}
}

func TestCurrentUserResolvesAVerifiedSubject(t *testing.T) {
	users := newFakeRepository()
	users.put(domain.User{ID: "user-1", Username: "Somchai", UsernameCanonical: "somchai"})
	service := newService(users, newFakeHasher(), newFakeIssuer())

	user, err := service.CurrentUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("CurrentUser() error = %v, want nil", err)
	}
	if user.Username != "Somchai" {
		t.Errorf("Username = %q, want %q", user.Username, "Somchai")
	}
}

// A token outliving the account it names must stop working at once.
func TestCurrentUserReportsADeletedAccount(t *testing.T) {
	service := newService(newFakeRepository(), newFakeHasher(), newFakeIssuer())

	if _, err := service.CurrentUser(context.Background(), "user-1"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("CurrentUser() error = %v, want ErrUserNotFound", err)
	}
}

func newService(users *fakeRepository, hasher *fakeHasher, tokens *fakeIssuer) *service.AuthService {
	return service.NewAuthService(
		users,
		hasher,
		tokens,
		func() string { return "generated-id" },
		func() time.Time { return fixedNow },
	)
}

// fakeRepository is an in memory stand in for the persistence port, with hooks
// for the failures a real database can produce.
type fakeRepository struct {
	byCanonical map[string]domain.User
	byID        map[string]domain.User
	findErr     error
	createErr   error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		byCanonical: make(map[string]domain.User),
		byID:        make(map[string]domain.User),
	}
}

func (r *fakeRepository) put(user domain.User) {
	r.byCanonical[user.UsernameCanonical] = user
	r.byID[user.ID] = user
}

func (r *fakeRepository) Create(_ context.Context, user domain.User) (domain.User, error) {
	if r.createErr != nil {
		return domain.User{}, r.createErr
	}
	r.put(user)
	return user, nil
}

func (r *fakeRepository) FindByCanonicalUsername(_ context.Context, canonical string) (domain.User, error) {
	if r.findErr != nil {
		return domain.User{}, r.findErr
	}
	user, ok := r.byCanonical[canonical]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *fakeRepository) FindByID(_ context.Context, userID string) (domain.User, error) {
	if r.findErr != nil {
		return domain.User{}, r.findErr
	}
	user, ok := r.byID[userID]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

// fakeHasher stands in for bcrypt with a cheap one way function, so the tests
// run fast while still holding the service to the rule that what reaches the
// repository is never the plaintext password.
type fakeHasher struct {
	hashErr   error
	verifyErr error
}

func newFakeHasher() *fakeHasher { return &fakeHasher{} }

func (h *fakeHasher) Hash(password string) (domain.PasswordHash, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	return fakeHash(password), nil
}

func (h *fakeHasher) Verify(hash domain.PasswordHash, password string) (bool, error) {
	if h.verifyErr != nil {
		return false, h.verifyErr
	}
	return hash == fakeHash(password), nil
}

// fakeHash is the fake hasher's algorithm, shared with the tests so they can
// state the expected stored value without repeating how it is produced.
func fakeHash(password string) domain.PasswordHash {
	sum := sha256.Sum256([]byte(password))
	return domain.PasswordHash("fake$" + hex.EncodeToString(sum[:]))
}

type fakeIssuer struct {
	issueErr error
}

func newFakeIssuer() *fakeIssuer { return &fakeIssuer{} }

func (i *fakeIssuer) Issue(user domain.User, issuedAt time.Time) (service.IssuedToken, error) {
	if i.issueErr != nil {
		return service.IssuedToken{}, i.issueErr
	}
	return service.IssuedToken{
		Value:     "token-for:" + user.ID,
		ExpiresAt: issuedAt.Add(time.Hour),
	}, nil
}
