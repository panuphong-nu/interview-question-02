package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestPoolConfig(t *testing.T) {
	configuration, err := poolConfig(
		"postgresql://user:secret@localhost:5432/postgres?sslmode=disable",
	)
	if err != nil {
		t.Fatalf("poolConfig() error = %v, want nil", err)
	}
	if configuration.ConnConfig.DefaultQueryExecMode != pgx.QueryExecModeExec {
		t.Errorf("query mode = %v, want QueryExecModeExec", configuration.ConnConfig.DefaultQueryExecMode)
	}
	if configuration.MinConns != 1 || configuration.MaxConns != 10 {
		t.Errorf("pool = %d..%d, want 1..10", configuration.MinConns, configuration.MaxConns)
	}
}

func TestHealthCheckerWrapsPingFailure(t *testing.T) {
	want := errors.New("database unavailable")
	checker := NewHealthChecker(fakePinger{err: want})

	if err := checker.Check(context.Background()); !errors.Is(err, want) {
		t.Fatalf("Check() error = %v, want it to wrap %v", err, want)
	}
}

type fakePinger struct{ err error }

func (p fakePinger) Ping(context.Context) error { return p.err }
