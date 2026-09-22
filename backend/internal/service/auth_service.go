package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"example.com/it02-auth/backend/internal/domain"
)

// AuthService contains the three use cases used by the frontend.
type AuthService struct {
	users  UserRepository
	hasher PasswordHasher
	tokens TokenIssuer
	newID  func() string
	now    func() time.Time
}

func NewAuthService(
	users UserRepository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	newID func() string,
	now func() time.Time,
) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens, newID: newID, now: now}
}

type Session struct {
	Token IssuedToken
	User  domain.User
}

func (s *AuthService) Register(ctx context.Context, input domain.Registration) (domain.User, error) {
	canonical := input.Canonical()
	if _, err := s.users.FindByCanonicalUsername(ctx, canonical); err == nil {
		return domain.User{}, domain.ErrUsernameTaken
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, fmt.Errorf("find username: %w", err)
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		ID:                s.newID(),
		Username:          input.Username,
		UsernameCanonical: canonical,
		PasswordHash:      hash,
		CreatedAt:         s.now().UTC(),
	}
	created, err := s.users.Create(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrUsernameTaken) {
			return domain.User{}, err
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return created, nil
}

func (s *AuthService) Login(ctx context.Context, input domain.LoginAttempt) (Session, error) {
	user, err := s.users.FindByCanonicalUsername(ctx, input.Canonical())
	if errors.Is(err, domain.ErrUserNotFound) {
		return Session{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, fmt.Errorf("find username: %w", err)
	}

	matches, err := s.hasher.Verify(user.PasswordHash, input.Password)
	if err != nil {
		return Session{}, fmt.Errorf("verify password: %w", err)
	}
	if !matches {
		return Session{}, domain.ErrInvalidCredentials
	}

	issued, err := s.tokens.Issue(user, s.now().UTC())
	if err != nil {
		return Session{}, fmt.Errorf("issue token: %w", err)
	}
	return Session{Token: issued, User: user}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID string) (domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}
