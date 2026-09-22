// Package service contains the application's use cases.
package service

import (
	"context"
	"time"

	"example.com/it02-auth/backend/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	FindByCanonicalUsername(ctx context.Context, canonical string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
}

type PasswordHasher interface {
	Hash(password string) (domain.PasswordHash, error)
	Verify(hash domain.PasswordHash, password string) (bool, error)
}

type TokenIssuer interface {
	Issue(user domain.User, issuedAt time.Time) (IssuedToken, error)
}

type IssuedToken struct {
	Value     string
	ExpiresAt time.Time
}
