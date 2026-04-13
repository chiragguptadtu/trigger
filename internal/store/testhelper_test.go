package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"trigger/db"
	"trigger/internal/store"
)

const defaultTestDSN = "postgres://trigger:trigger@localhost:5432/trigger?sslmode=disable"

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	if err := db.RunMigrations(dsn); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	pool, err := db.Connect(context.Background(), dsn)
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}
	testPool = pool
	defer testPool.Close()

	os.Exit(m.Run())
}

// withTx runs fn inside a transaction that is always rolled back,
// keeping each test fully isolated without truncating tables.
func withTx(t *testing.T, fn func(q *store.Queries)) {
	t.Helper()
	tx, err := testPool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	fn(store.New(tx))
}

// uniqueEmail returns a unique email address for each test invocation,
// preventing clashes with committed rows left by previous test runs or manual curl tests.
func uniqueEmail() string {
	return fmt.Sprintf("store-%d@example.com", time.Now().UnixNano())
}
