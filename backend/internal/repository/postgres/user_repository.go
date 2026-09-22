package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"example.com/it02-auth/backend/internal/domain"
)

const usernameConstraint = "app_users_username_canonical_key"

type database interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// UserRepository stores accounts in PostgreSQL. It satisfies
// service.UserRepository.
type UserRepository struct {
	database database
}

func NewUserRepository(database database) *UserRepository {
	return &UserRepository{database: database}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	const statement = `
		INSERT INTO public.app_users
			(id, username, username_canonical, password_hash, created_at)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING id::text, username, username_canonical, password_hash, created_at`

	created, err := scanUser(r.database.QueryRow(ctx, statement,
		user.ID,
		user.Username,
		user.UsernameCanonical,
		string(user.PasswordHash),
		user.CreatedAt,
	))
	if err != nil {
		if isUsernameConflict(err) {
			return domain.User{}, domain.ErrUsernameTaken
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return created, nil
}

func (r *UserRepository) FindByCanonicalUsername(ctx context.Context, canonical string) (domain.User, error) {
	const query = `
		SELECT id::text, username, username_canonical, password_hash, created_at
		FROM public.app_users
		WHERE username_canonical = $1`

	return r.queryOne(ctx, query, canonical)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	const query = `
		SELECT id::text, username, username_canonical, password_hash, created_at
		FROM public.app_users
		WHERE id = $1::uuid`

	return r.queryOne(ctx, query, id)
}

func (r *UserRepository) queryOne(ctx context.Context, query string, argument any) (domain.User, error) {
	user, err := scanUser(r.database.QueryRow(ctx, query, argument))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("query user: %w", err)
	}
	return user, nil
}

func scanUser(row pgx.Row) (domain.User, error) {
	var (
		user domain.User
		hash string
	)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.UsernameCanonical,
		&hash,
		&user.CreatedAt,
	)
	user.PasswordHash = domain.PasswordHash(hash)
	return user, err
}

func isUsernameConflict(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == usernameConstraint
}
