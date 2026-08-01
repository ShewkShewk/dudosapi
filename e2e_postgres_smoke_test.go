//go:build e2e

package main

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestPostgresSnapshotIsolatesTests proves the TestMain container is up,
// migrations actually ran (the insert below fails otherwise), and
// resetDB's snapshot/restore genuinely wipes data between tests rather than
// just existing as an unused API call. Subtests run sequentially, so the
// first one's Cleanup-triggered Restore is guaranteed to happen before the
// second one starts.
func TestPostgresSnapshotIsolatesTests(t *testing.T) {
	ctx := context.Background()

	t.Run("insert a row", func(t *testing.T) {
		resetDB(t)

		pool, err := pgxpool.New(ctx, pgDSN(ctx))
		if err != nil {
			t.Fatalf("pgxpool.New: %v", err)
		}
		defer pool.Close()

		if _, err := pool.Exec(ctx, `INSERT INTO schools (id, name) VALUES (1, 'Leftover School')`); err != nil {
			t.Fatalf("insert: %v", err)
		}
	})

	t.Run("previous row must not exist", func(t *testing.T) {
		resetDB(t)

		pool, err := pgxpool.New(ctx, pgDSN(ctx))
		if err != nil {
			t.Fatalf("pgxpool.New: %v", err)
		}
		defer pool.Close()

		var count int
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schools`).Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 0 {
			t.Fatalf("schools count = %d, want 0 (restore should have wiped the previous subtest's insert)", count)
		}
	})
}
