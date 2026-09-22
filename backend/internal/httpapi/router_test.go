package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/httpapi"
	"example.com/it02-auth/backend/internal/infrastructure/id"
	"example.com/it02-auth/backend/internal/infrastructure/security"
	"example.com/it02-auth/backend/internal/infrastructure/token"
	"example.com/it02-auth/backend/internal/service"
)

const allowedOrigin = "http://localhost:4200"

// The HTTP stack is exercised with real bcrypt and JWT adapters. Persistence
// has its own PostgreSQL repository tests, so these tests use a small in-memory
// implementation and need no external database service.

func TestRegisterThenSignInThenReadTheProfile(t *testing.T) {
	router := newRouter(t)

	// IT 02-2: create the account.
	registered := do(t, router, request(t, http.MethodPost, "/api/auth/register", map[string]string{
		"username":        "Somchai",
		"password":        "sup3r-secret",
		"confirmPassword": "sup3r-secret",
	}))
	if registered.status != http.StatusCreated {
		t.Fatalf("register status = %d, want %d (body %s)", registered.status, http.StatusCreated, registered.body)
	}
	// Registering does not start a session; the person is sent back to sign in.
	if strings.Contains(registered.body, "accessToken") {
		t.Errorf("register body = %s, want no token", registered.body)
	}
	if strings.Contains(registered.body, "sup3r-secret") || strings.Contains(registered.body, "$2") {
		t.Errorf("register body = %s, want no password material", registered.body)
	}

	// IT 02-1: sign in.
	signedIn := do(t, router, request(t, http.MethodPost, "/api/auth/login", map[string]string{
		"username": "somchai", // a different case than the one registered
		"password": "sup3r-secret",
	}))
	if signedIn.status != http.StatusOK {
		t.Fatalf("login status = %d, want %d (body %s)", signedIn.status, http.StatusOK, signedIn.body)
	}

	var session struct {
		AccessToken string `json:"accessToken"`
		TokenType   string `json:"tokenType"`
		ExpiresAt   string `json:"expiresAt"`
		User        struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"user"`
	}
	decode(t, signedIn.body, &session)

	if session.TokenType != "Bearer" {
		t.Errorf("tokenType = %q, want %q", session.TokenType, "Bearer")
	}
	if session.User.Username != "Somchai" {
		t.Errorf("user.username = %q, want the name as registered", session.User.Username)
	}
	if strings.Count(session.AccessToken, ".") != 2 {
		t.Errorf("accessToken = %q, want a three part JWT", session.AccessToken)
	}
	if _, err := time.Parse(time.RFC3339, session.ExpiresAt); err != nil {
		t.Errorf("expiresAt = %q, want RFC 3339: %v", session.ExpiresAt, err)
	}

	// IT 02-3: the welcome screen resolves the name behind the token.
	profileRequest := request(t, http.MethodGet, "/api/auth/me", nil)
	profileRequest.Header.Set("Authorization", "Bearer "+session.AccessToken)

	profile := do(t, router, profileRequest)
	if profile.status != http.StatusOK {
		t.Fatalf("me status = %d, want %d (body %s)", profile.status, http.StatusOK, profile.body)
	}

	var user struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	decode(t, profile.body, &user)

	if user.Username != "Somchai" {
		t.Errorf("username = %q, want %q", user.Username, "Somchai")
	}
	if user.ID != session.User.ID {
		t.Errorf("id = %q, want the signed in account %q", user.ID, session.User.ID)
	}
}

func TestRegisterRejectsAMismatchedConfirmation(t *testing.T) {
	router := newRouter(t)

	response := do(t, router, request(t, http.MethodPost, "/api/auth/register", map[string]string{
		"username":        "Somchai",
		"password":        "sup3r-secret",
		"confirmPassword": "sup3r-secrets",
	}))
	if response.status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body %s)", response.status, http.StatusBadRequest, response.body)
	}

	var problem struct {
		Errors map[string][]string `json:"errors"`
	}
	decode(t, response.body, &problem)

	// The message is attached to the field the form asks the person to fix.
	if len(problem.Errors["confirmPassword"]) == 0 {
		t.Errorf("errors = %+v, want a message on confirmPassword", problem.Errors)
	}
}

func TestRegisterRejectsATakenUsername(t *testing.T) {
	router := newRouter(t)
	body := map[string]string{
		"username":        "Somchai",
		"password":        "sup3r-secret",
		"confirmPassword": "sup3r-secret",
	}

	if first := do(t, router, request(t, http.MethodPost, "/api/auth/register", body)); first.status != http.StatusCreated {
		t.Fatalf("first register status = %d, want %d", first.status, http.StatusCreated)
	}

	body["username"] = "SOMCHAI"
	second := do(t, router, request(t, http.MethodPost, "/api/auth/register", body))
	if second.status != http.StatusConflict {
		t.Fatalf("second register status = %d, want %d (body %s)", second.status, http.StatusConflict, second.body)
	}
}

func TestLoginRejectsWrongCredentials(t *testing.T) {
	router := newRouter(t)
	register(t, router, "Somchai", "sup3r-secret")

	tests := map[string]map[string]string{
		"wrong password": {"username": "Somchai", "password": "not-the-password"},
		"unknown user":   {"username": "nobody", "password": "sup3r-secret"},
	}

	bodies := make(map[string]string, len(tests))
	for name, payload := range tests {
		response := do(t, router, request(t, http.MethodPost, "/api/auth/login", payload))
		if response.status != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want %d (body %s)", name, response.status, http.StatusUnauthorized, response.body)
		}
		bodies[name] = response.body
	}

	// Identical answers, so the endpoint cannot be used to discover which
	// usernames are registered.
	if bodies["wrong password"] != bodies["unknown user"] {
		t.Errorf("a wrong password and an unknown user answered differently:\n%s\n%s",
			bodies["wrong password"], bodies["unknown user"])
	}
}

func TestProfileRequiresAValidToken(t *testing.T) {
	router := newRouter(t)

	tests := map[string]string{
		"no header":       "",
		"wrong scheme":    "Basic c29tY2hhaTpzZWNyZXQ=",
		"empty bearer":    "Bearer ",
		"forged token":    "Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyLTEifQ.",
		"nonsense token":  "Bearer not-a-token",
		"token from else": "Bearer " + foreignToken(t),
	}

	for name, header := range tests {
		t.Run(name, func(t *testing.T) {
			profileRequest := request(t, http.MethodGet, "/api/auth/me", nil)
			if header != "" {
				profileRequest.Header.Set("Authorization", header)
			}

			response := do(t, router, profileRequest)
			if response.status != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d (body %s)", response.status, http.StatusUnauthorized, response.body)
			}
			// 401 without a challenge leaves a client nothing to act on.
			if response.header.Get("WWW-Authenticate") == "" {
				t.Error("missing WWW-Authenticate header")
			}
		})
	}
}

func TestMalformedBodiesAreRejected(t *testing.T) {
	router := newRouter(t)

	tests := map[string]struct {
		contentType string
		body        string
		want        int
	}{
		"not json":       {contentType: "application/json", body: "{", want: http.StatusBadRequest},
		"unknown field":  {contentType: "application/json", body: `{"username":"a","password":"b","role":"admin"}`, want: http.StatusBadRequest},
		"two objects":    {contentType: "application/json", body: `{"username":"a","password":"b"}{"username":"c"}`, want: http.StatusBadRequest},
		"wrong media":    {contentType: "text/plain", body: `{"username":"a","password":"b"}`, want: http.StatusUnsupportedMediaType},
		"empty username": {contentType: "application/json", body: `{"username":"","password":""}`, want: http.StatusBadRequest},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			raw := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(test.body))
			raw.Header.Set("Content-Type", test.contentType)

			if response := do(t, router, raw); response.status != test.want {
				t.Fatalf("status = %d, want %d (body %s)", response.status, test.want, response.body)
			}
		})
	}
}

func TestHealthReportsTheDatabase(t *testing.T) {
	response := do(t, newRouter(t), request(t, http.MethodGet, "/health", nil))

	if response.status != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %s)", response.status, http.StatusOK, response.body)
	}
	if !strings.Contains(response.body, `"status":"ok"`) {
		t.Errorf("body = %s, want an ok status", response.body)
	}
}

func TestCORSAnswersOnlyTheAllowedOrigin(t *testing.T) {
	router := newRouter(t)

	tests := map[string]struct {
		origin string
		want   string
	}{
		"allowed origin":    {origin: allowedOrigin, want: allowedOrigin},
		"disallowed origin": {origin: "https://evil.example", want: ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			preflight := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
			preflight.Header.Set("Origin", test.origin)
			preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)

			response := do(t, router, preflight)
			if got := response.header.Get("Access-Control-Allow-Origin"); got != test.want {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, test.want)
			}
		})
	}
}

// newRouter wires the production security adapters with bcrypt at its cheapest
// cost so the suite stays fast.
func newRouter(t *testing.T) http.Handler {
	t.Helper()

	signer := token.NewSigner(token.Options{
		Secret:   []byte("a-test-secret-of-at-least-32-bytes!!"),
		Issuer:   "it02-auth",
		Audience: "it02-web",
		Lifetime: time.Hour,
	})

	return httpapi.NewRouter(httpapi.Options{
		Auth: service.NewAuthService(
			newMemoryUserRepository(),
			security.NewBcryptHasher(security.MinCost),
			signer,
			id.NewUUID,
			time.Now,
		),
		Verifier:       signer,
		Health:         healthyDependency{},
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins: []string{allowedOrigin},
	})
}

type memoryUserRepository struct {
	mutex       sync.RWMutex
	byID        map[string]domain.User
	byCanonical map[string]domain.User
}

func newMemoryUserRepository() *memoryUserRepository {
	return &memoryUserRepository{
		byID:        make(map[string]domain.User),
		byCanonical: make(map[string]domain.User),
	}
}

func (r *memoryUserRepository) Create(_ context.Context, user domain.User) (domain.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, exists := r.byCanonical[user.UsernameCanonical]; exists {
		return domain.User{}, domain.ErrUsernameTaken
	}
	r.byID[user.ID] = user
	r.byCanonical[user.UsernameCanonical] = user
	return user, nil
}

func (r *memoryUserRepository) FindByCanonicalUsername(_ context.Context, canonical string) (domain.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	user, exists := r.byCanonical[canonical]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *memoryUserRepository) FindByID(_ context.Context, id string) (domain.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	user, exists := r.byID[id]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

type healthyDependency struct{}

func (healthyDependency) Check(context.Context) error { return nil }

// foreignToken is a well formed token signed by a different service, which
// must not be accepted here.
func foreignToken(t *testing.T) string {
	t.Helper()

	issued, err := token.NewSigner(token.Options{
		Secret:   []byte("a-different-secret-of-at-least-32-bytes"),
		Issuer:   "it02-auth",
		Audience: "it02-web",
		Lifetime: time.Hour,
	}).Issue(domain.User{ID: "user-1", Username: "Somchai"}, time.Now())
	if err != nil {
		t.Fatalf("issue foreign token: %v", err)
	}
	return issued.Value
}

type recorded struct {
	status int
	body   string
	header http.Header
}

func do(t *testing.T, router http.Handler, raw *http.Request) recorded {
	t.Helper()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, raw)
	result := recorder.Result()
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return recorded{status: result.StatusCode, body: string(body), header: result.Header}
}

func request(t *testing.T, method, target string, payload map[string]string) *http.Request {
	t.Helper()

	if payload == nil {
		return httptest.NewRequest(method, target, nil)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode request body: %v", err)
	}
	raw := httptest.NewRequest(method, target, bytes.NewReader(encoded))
	raw.Header.Set("Content-Type", "application/json")
	return raw
}

func register(t *testing.T, router http.Handler, username, password string) {
	t.Helper()

	response := do(t, router, request(t, http.MethodPost, "/api/auth/register", map[string]string{
		"username":        username,
		"password":        password,
		"confirmPassword": password,
	}))
	if response.status != http.StatusCreated {
		t.Fatalf("register status = %d, want %d (body %s)", response.status, http.StatusCreated, response.body)
	}
}

func decode(t *testing.T, body string, target any) {
	t.Helper()

	if err := json.Unmarshal([]byte(body), target); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
}
