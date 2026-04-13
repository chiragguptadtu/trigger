package command_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/db"
	"trigger/internal/command"
	"trigger/internal/store"
)

const defaultTestDSN = "postgres://trigger:trigger@localhost:5432/trigger?sslmode=disable"

var testStore *store.Queries

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	// Non-fatal: parser tests run fine without DB; only syncer tests skip.
	if pool, err := db.Connect(context.Background(), dsn); err == nil {
		testStore = store.New(pool)
		defer pool.Close()
	}
	os.Exit(m.Run())
}

func skipIfNoDB(t *testing.T) {
	t.Helper()
	if testStore == nil {
		t.Skip("database not available")
	}
}

func TestSync_InsertsNewCommands(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	scanned := []command.ScannedCommand{
		{
			Slug:       "sync-test-cmd",
			ScriptPath: "/commands/sync_test_cmd.py",
			Command: &command.Command{
				Name:        "Sync Test Command",
				Description: "for testing sync",
				Inputs: []command.Input{
					{Name: "env", Label: "Env", Type: command.InputTypeClosed,
						Options: []string{"staging", "prod"}, Required: true},
				},
			},
		},
	}

	err := command.Sync(ctx, testStore, scanned)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cleanupCommand(ctx, "sync-test-cmd") })

	cmd, err := testStore.GetCommandBySlug(ctx, "sync-test-cmd")
	require.NoError(t, err)
	assert.Equal(t, "Sync Test Command", cmd.Name)

	inputs, err := testStore.ListCommandInputs(ctx, cmd.ID)
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	assert.Equal(t, "env", inputs[0].Name)
}

func TestSync_UpdatesExistingCommand(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	initial := []command.ScannedCommand{{
		Slug: "sync-update-cmd", ScriptPath: "/commands/update.py",
		Command: &command.Command{Name: "Old Name", Inputs: nil},
	}}
	require.NoError(t, command.Sync(ctx, testStore, initial))
	t.Cleanup(func() { _ = cleanupCommand(ctx, "sync-update-cmd") })

	updated := []command.ScannedCommand{{
		Slug: "sync-update-cmd", ScriptPath: "/commands/update.py",
		Command: &command.Command{
			Name: "New Name",
			Inputs: []command.Input{
				{Name: "reason", Label: "Reason", Type: command.InputTypeOpen, Required: true},
			},
		},
	}}
	require.NoError(t, command.Sync(ctx, testStore, updated))

	cmd, err := testStore.GetCommandBySlug(ctx, "sync-update-cmd")
	require.NoError(t, err)
	assert.Equal(t, "New Name", cmd.Name)

	inputs, err := testStore.ListCommandInputs(ctx, cmd.ID)
	require.NoError(t, err)
	require.Len(t, inputs, 1)
}

func TestSync_DeactivatesRemovedCommands(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	// Insert a command via direct upsert (simulates a previously-scanned command)
	_, err := testStore.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: "sync-removed-cmd", Name: "To Be Removed",
		ScriptPath: "/commands/removed.py",
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = cleanupCommand(ctx, "sync-removed-cmd") })

	// Sync with an empty list — removed-cmd should be deactivated
	require.NoError(t, command.Sync(ctx, testStore, []command.ScannedCommand{}))

	_, err = testStore.GetCommandBySlug(ctx, "sync-removed-cmd")
	assert.Error(t, err, "deactivated command should not be returned by GetCommandBySlug")
}

func cleanupCommand(ctx context.Context, slug string) error {
	return testStore.DeactivateCommand(ctx, slug)
}
