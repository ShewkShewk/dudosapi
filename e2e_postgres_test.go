//go:build e2e

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the "pgx5" migrate driver, and as a side effect the "pgx" database/sql driver
	_ "github.com/golang-migrate/migrate/v4/source/file"     // registers the "file" migrate source
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// pgSnapshotName is the template database Restore() resets to before every
// test. It's created once in TestMain right after migrations run.
const pgSnapshotName = "e2e_migrated"

var pgContainer *postgres.PostgresContainer

// TestMain starts one Postgres container for the whole e2e test binary run,
// applies db/migrations/ to it, and snapshots the result as a clean baseline.
// Individual tests call resetDB(t) to restore that baseline after they run.
func TestMain(m *testing.M) {
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("dudos_duda"),
		postgres.WithUsername("dudos_test"),
		postgres.WithPassword("dudos_test"),
		postgres.WithSQLDriver("pgx"), // matches the driver registered by migrate's pgx/v5 package; avoids a lib/pq dependency
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("e2e TestMain: unable to start postgres container: %v", err)
	}
	pgContainer = ctr

	if err := runMigrations(ctx); err != nil {
		log.Fatalf("e2e TestMain: unable to run migrations: %v", err)
	}

	if err := pgContainer.Snapshot(ctx, postgres.WithSnapshotName(pgSnapshotName)); err != nil {
		log.Fatalf("e2e TestMain: unable to snapshot migrated database: %v", err)
	}

	if err := startFakeGCS(ctx); err != nil {
		log.Fatalf("e2e TestMain: unable to start fake-gcs-server: %v", err)
	}

	code := m.Run()

	if err := testcontainers.TerminateContainer(pgContainer); err != nil {
		log.Printf("e2e TestMain: unable to terminate postgres container: %v", err)
	}
	if err := testcontainers.TerminateContainer(gcsContainer); err != nil {
		log.Printf("e2e TestMain: unable to terminate fake-gcs-server container: %v", err)
	}

	os.Exit(code)
}

// pgDSN returns a standard postgres:// connection string for the shared
// test container, suitable for pgxpool.New (what the app itself uses).
func pgDSN(ctx context.Context) string {
	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("pgDSN: %v", err))
	}
	return dsn
}

func runMigrations(ctx context.Context) error {
	dsn := pgDSN(ctx)
	rest, ok := strings.CutPrefix(dsn, "postgres://")
	if !ok {
		return fmt.Errorf("unexpected dsn scheme, want postgres://: %s", dsn)
	}
	migrateDSN := "pgx5://" + rest

	m, err := migrate.New("file://db/migrations", migrateDSN)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate Up: %w", err)
	}
	return nil
}

// resetDB registers a cleanup that restores the database to its
// freshly-migrated snapshot after the current test finishes, so the *next*
// test starts clean regardless of whether this one passed, failed, or left
// rows behind.
func resetDB(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		if err := pgContainer.Restore(ctx, postgres.WithSnapshotName(pgSnapshotName)); err != nil {
			t.Fatalf("resetDB: unable to restore snapshot: %v", err)
		}
	})
}
