package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"example.com/it02-auth/backend/internal/domain"
	"example.com/it02-auth/backend/internal/repository/postgres"
)

var createdAt = time.Date(2026, time.September, 22, 10, 0, 0, 0, time.UTC)

func TestCreateReturnsTheDatabaseRepresentation(t *testing.T) {
	database := &fakeDatabase{row: fakeRow{values: []any{
		"67efb2e7-5127-4fe1-8909-6172d60dbe17",
		"Somchai",
		"somchai",
		"$2a$11$hash",
		createdAt,
	}}}
	repository := postgres.NewUserRepository(database)

	created, err := repository.Create(context.Background(), domain.User{
		ID:                "67efb2e7-5127-4fe1-8909-6172d60dbe17",
		Username:          "Somchai",
		UsernameCanonical: "somchai",
		PasswordHash:      "$2a$11$hash",
		CreatedAt:         createdAt,
	})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if created.UsernameCanonical != "somchai" {
		t.Errorf("UsernameCanonical = %q, want persisted canonical value", created.UsernameCanonical)
	}
	if len(database.arguments) != 5 {
		t.Fatalf("query arguments = %d, want 5", len(database.arguments))
	}
}

func TestCreateMapsOnlyTheUsernameUniqueConstraint(t *testing.T) {
	tests := map[string]struct {
		constraint string
		wantTaken  bool
	}{
		"username conflict": {constraint: "app_users_username_canonical_key", wantTaken: true},
		"another conflict":  {constraint: "some_other_key", wantTaken: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			database := &fakeDatabase{row: fakeRow{err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: test.constraint,
			}}}
			_, err := postgres.NewUserRepository(database).Create(context.Background(), domain.User{})
			if errors.Is(err, domain.ErrUsernameTaken) != test.wantTaken {
				t.Fatalf("Create() error = %v, ErrUsernameTaken = %v, want %v", err, errors.Is(err, domain.ErrUsernameTaken), test.wantTaken)
			}
		})
	}
}

func TestFindMapsNoRows(t *testing.T) {
	repository := postgres.NewUserRepository(&fakeDatabase{row: fakeRow{err: pgx.ErrNoRows}})

	_, err := repository.FindByCanonicalUsername(context.Background(), "missing")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("FindByCanonicalUsername() error = %v, want ErrUserNotFound", err)
	}
}

type fakeDatabase struct {
	row       pgx.Row
	arguments []any
}

func (d *fakeDatabase) QueryRow(_ context.Context, _ string, arguments ...any) pgx.Row {
	d.arguments = arguments
	return d.row
}

type fakeRow struct {
	values []any
	err    error
}

func (r fakeRow) Scan(destinations ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(destinations) != len(r.values) {
		return errors.New("destination count does not match values")
	}

	for index, value := range r.values {
		switch destination := destinations[index].(type) {
		case *string:
			*destination = value.(string)
		case *time.Time:
			*destination = value.(time.Time)
		default:
			return errors.New("unsupported scan destination")
		}
	}
	return nil
}
