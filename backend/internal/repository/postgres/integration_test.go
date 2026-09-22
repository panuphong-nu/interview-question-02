package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/infrastructure/id"
	"example.com/it02-auth/backend/internal/repository/postgres"
)

// Run with TEST_DATABASE_URL pointing at a migrated disposable database. The
// transaction is always rolled back, so the test also works with the runtime
// role, which intentionally has no DELETE permission.
func TestUserRepositoryAgainstPostgres(t *testing.T) {
	connectionString := os.Getenv("TEST_DATABASE_URL")
	if connectionString == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	transaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	t.Cleanup(func() { _ = transaction.Rollback(ctx) })

	repository := postgres.NewUserRepository(transaction)
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())
	want := domain.User{
		ID:                id.NewUUID(),
		Username:          username,
		UsernameCanonical: domain.CanonicalUsername(username),
		PasswordHash:      "$2a$11$abcdefghijklmnopqrstuvwxyz012345678901234567890123456",
		CreatedAt:         time.Now().UTC().Truncate(time.Microsecond),
	}

	created, err := repository.Create(ctx, want)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	found, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Username != want.Username || found.UsernameCanonical != want.UsernameCanonical {
		t.Errorf("found user = %+v, want username %q and canonical %q", found, want.Username, want.UsernameCanonical)
	}

	duplicate := want
	duplicate.ID = id.NewUUID()
	if _, err := repository.Create(ctx, duplicate); !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("duplicate Create() error = %v, want ErrUsernameTaken", err)
	}
}
