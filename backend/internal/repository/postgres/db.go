// Package postgres implements the persistence layer with PostgreSQL. The same
// code works with a direct Supabase connection and its session or transaction
// pooler.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const connectTimeout = 10 * time.Second

// Open creates and verifies one process-wide PostgreSQL connection pool.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	configuration, err := poolConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, configuration)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	return pool, nil
}

func poolConfig(databaseURL string) (*pgxpool.Config, error) {
	configuration, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		// Parse errors can echo the input. Do not let a connection string that
		// contains a password reach startup logs.
		return nil, fmt.Errorf("parse DATABASE_URL: invalid PostgreSQL connection settings")
	}

	configuration.MaxConns = 10
	configuration.MinConns = 1
	// This mode works with direct, session and transaction pooler URLs.
	configuration.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	return configuration, nil
}

type pinger interface {
	Ping(context.Context) error
}

// HealthChecker reports whether PostgreSQL is reachable.
type HealthChecker struct {
	database pinger
}

func NewHealthChecker(database pinger) *HealthChecker {
	return &HealthChecker{database: database}
}

func (c *HealthChecker) Check(ctx context.Context) error {
	if err := c.database.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	return nil
}
