package execution_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chiragguptadtu/trigger/db"
	"github.com/chiragguptadtu/trigger/internal/execution"
	"github.com/chiragguptadtu/trigger/internal/store"
)

const defaultTestDSN = "postgres://trigger:trigger@localhost:5432/trigger?sslmode=disable"

var (
	testPool  *pgxpool.Pool
	testStore *store.Queries
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	if err := db.RunMigrations(dsn); err != nil {
		panic("migrations: " + err.Error())
	}
	pool, err := db.Connect(context.Background(), dsn)
	if err != nil {
		panic("connect: " + err.Error())
	}
	testPool = pool
	testStore = store.New(pool)
	defer pool.Close()
	os.Exit(m.Run())
}

func skipIfNoDB(t *testing.T) {
	t.Helper()
	if testStore == nil {
		t.Skip("database not available")
	}
}

// workerCleanup registers a t.Cleanup that runs the given SQL. Cleanups run in
// LIFO order — register parent rows first so child rows run first at teardown.
func workerCleanup(t *testing.T, query string, args ...any) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := testPool.Exec(context.Background(), query, args...); err != nil {
			t.Logf("cleanup query failed: %v | query: %s", err, query)
		}
	})
}

func uniqueWorkerSlug(base string) string {
	return fmt.Sprintf("worker-%s-%d", base, time.Now().UnixNano())
}

func TestWorker_SuccessfulExecution(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	// loadConfig decrypts every config_entry in the DB. Pre-existing entries from
	// functional/curl testing (encrypted with a different key) cause decrypt to fail,
	// which makes the worker return an error and River retries indefinitely.
	// We clear all config entries as test environment setup — not as cleanup of our data.
	_, err := testPool.Exec(ctx, "DELETE FROM config_entries")
	require.NoError(t, err)

	user, err := testStore.CreateUser(ctx, store.CreateUserParams{
		Email: uniqueWorkerEmail(), Name: "Worker User", PasswordHash: "hashed",
	})
	require.NoError(t, err)
	workerCleanup(t, "DELETE FROM users WHERE id = $1", user.ID) // registered 1st — runs last

	scriptPath := writeScript(t, "worker_ok.py", `#!/usr/bin/env python3
import sys, json
inputs = json.loads(sys.argv[1])
# success
`)
	slug := uniqueWorkerSlug("ok")
	cmd, err := testStore.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: slug, Name: "Worker OK", ScriptPath: scriptPath,
	})
	require.NoError(t, err)
	// Delete executions first (registered 2nd — runs 2nd-to-last), then command (3rd — runs last-ish).
	// We use command_id to catch any executions from previous runs using the same command row.
	workerCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID)          // registered 2nd — runs 3rd-to-last
	workerCleanup(t, "DELETE FROM executions WHERE command_id = $1", cmd.ID) // registered 3rd — runs 2nd-to-last

	inputsJSON, _ := json.Marshal(map[string]any{"env": "staging"})
	exec_, err := testStore.CreateExecution(ctx, store.CreateExecutionParams{
		CommandID:   cmd.ID,
		TriggeredBy: user.ID,
		Inputs:      inputsJSON,
	})
	require.NoError(t, err)

	worker := execution.NewWorker(testStore, make([]byte, 32))
	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, worker)

	riverClient, err := river.NewClient(riverpgxv5.New(testPool), &river.Config{
		Workers: riverWorkers,
		Queues:  map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 2}},
	})
	require.NoError(t, err)

	require.NoError(t, riverClient.Start(ctx))
	t.Cleanup(func() { _ = riverClient.Stop(ctx) })

	_, err = riverClient.Insert(ctx, execution.ExecutionArgs{ExecutionID: exec_.ID.String()}, nil)
	require.NoError(t, err)

	// Poll until the execution reaches a terminal state (max 5s)
	var result store.Execution
	require.Eventually(t, func() bool {
		result, err = testStore.GetExecutionByID(ctx, exec_.ID)
		return err == nil && (result.Status == "success" || result.Status == "failure")
	}, 5*time.Second, 100*time.Millisecond)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.ErrorMessage)
	assert.True(t, result.StartedAt.Valid)
	assert.True(t, result.CompletedAt.Valid)
}

func TestWorker_FailedExecution(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	// See comment in TestWorker_SuccessfulExecution — clear config_entries as setup.
	_, err := testPool.Exec(ctx, "DELETE FROM config_entries")
	require.NoError(t, err)

	user, err := testStore.CreateUser(ctx, store.CreateUserParams{
		Email: uniqueWorkerEmail(), Name: "Worker User", PasswordHash: "hashed",
	})
	require.NoError(t, err)
	workerCleanup(t, "DELETE FROM users WHERE id = $1", user.ID)

	scriptPath := writeScript(t, "worker_fail.py", `#!/usr/bin/env python3
import sys
print("something went wrong", file=sys.stderr)
sys.exit(1)
`)
	slug := uniqueWorkerSlug("fail")
	cmd, err := testStore.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: slug, Name: "Worker Fail", ScriptPath: scriptPath,
	})
	require.NoError(t, err)
	workerCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID)
	workerCleanup(t, "DELETE FROM executions WHERE command_id = $1", cmd.ID)

	exec_, err := testStore.CreateExecution(ctx, store.CreateExecutionParams{
		CommandID: cmd.ID, TriggeredBy: user.ID, Inputs: []byte("{}"),
	})
	require.NoError(t, err)

	worker := execution.NewWorker(testStore, make([]byte, 32))
	riverWorkers := river.NewWorkers()
	river.AddWorker(riverWorkers, worker)

	riverClient, err := river.NewClient(riverpgxv5.New(testPool), &river.Config{
		Workers: riverWorkers,
		Queues:  map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 2}},
	})
	require.NoError(t, err)
	require.NoError(t, riverClient.Start(ctx))
	t.Cleanup(func() { _ = riverClient.Stop(ctx) })

	_, err = riverClient.Insert(ctx, execution.ExecutionArgs{ExecutionID: exec_.ID.String()}, nil)
	require.NoError(t, err)

	var result store.Execution
	require.Eventually(t, func() bool {
		result, err = testStore.GetExecutionByID(ctx, exec_.ID)
		return err == nil && (result.Status == "success" || result.Status == "failure")
	}, 5*time.Second, 100*time.Millisecond)

	assert.Equal(t, "failure", result.Status)
	assert.Equal(t, "something went wrong", result.ErrorMessage)
}

func uniqueWorkerEmail() string {
	return "worker-" + time.Now().Format("150405.000000000") + "@example.com"
}
