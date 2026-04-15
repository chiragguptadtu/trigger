package command_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chiragguptadtu/trigger/db"
	"github.com/chiragguptadtu/trigger/internal/command"
	"github.com/chiragguptadtu/trigger/internal/store"
)

const defaultTestDSN = "postgres://trigger:trigger@localhost:5432/trigger?sslmode=disable"

var testStore *store.Queries

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	// Migrations first, then connect — same order as handler/execution tests.
	// Non-fatal: if DB is unavailable, syncer tests skip and parser tests run.
	if err := db.RunMigrations(dsn); err == nil {
		if pool, err := db.Connect(context.Background(), dsn); err == nil {
			testStore = store.New(pool)
			defer pool.Close()
		}
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

	err := command.Sync(ctx, testStore, scanned, nil)
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
	require.NoError(t, command.Sync(ctx, testStore, initial, nil))
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
	require.NoError(t, command.Sync(ctx, testStore, updated, nil))

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
	require.NoError(t, command.Sync(ctx, testStore, []command.ScannedCommand{}, nil))

	_, err = testStore.GetCommandBySlug(ctx, "sync-removed-cmd")
	assert.Error(t, err, "deactivated command should not be returned by GetCommandBySlug")
}

func TestSync_DynamicInputResolvesOptions(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	// Write a real script with get_options for two inputs.
	scriptPath := filepath.Join(t.TempDir(), "dynamic_cmd.py")
	require.NoError(t, os.WriteFile(scriptPath, []byte(`#!/usr/bin/env python3
import sys, json

def get_options(input_name, config):
    if input_name == "environment":
        envs = config.get("AVAILABLE_ENVIRONMENTS", "staging,production")
        return [e.strip() for e in envs.split(",") if e.strip()]
    if input_name == "region":
        return ["us-east-1", "eu-west-1"]
    return []

if __name__ == "__main__":
    if sys.argv[1] == "--trigger-get-options":
        input_name = sys.argv[2] if len(sys.argv) > 2 else ""
        config = json.loads(sys.argv[3]) if len(sys.argv) > 3 else {}
        print("\n".join(get_options(input_name, config)))
        sys.exit(0)
`), 0755))

	scanned := []command.ScannedCommand{
		{
			Slug:       "sync-dynamic-cmd",
			ScriptPath: scriptPath,
			Command: &command.Command{
				Name: "Dynamic Cmd",
				Inputs: []command.Input{
					{Name: "environment", Label: "Environment", Type: command.InputTypeClosed, Dynamic: true, Required: true},
					{Name: "region", Label: "Region", Type: command.InputTypeClosed, Dynamic: true, Required: true},
				},
			},
		},
	}

	// nil key is fine — no config entries in the test DB to decrypt.
	err := command.Sync(ctx, testStore, scanned, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cleanupCommand(ctx, "sync-dynamic-cmd") })

	cmd, err := testStore.GetCommandBySlug(ctx, "sync-dynamic-cmd")
	require.NoError(t, err)

	inputs, err := testStore.ListCommandInputs(ctx, cmd.ID)
	require.NoError(t, err)
	require.Len(t, inputs, 2)

	envInput := inputs[0]
	assert.Equal(t, "environment", envInput.Name)

	var envOpts []string
	require.NoError(t, json.Unmarshal(envInput.Options, &envOpts))
	assert.Equal(t, []string{"staging", "production"}, envOpts)

	regionInput := inputs[1]
	assert.Equal(t, "region", regionInput.Name)

	var regionOpts []string
	require.NoError(t, json.Unmarshal(regionInput.Options, &regionOpts))
	assert.Equal(t, []string{"us-east-1", "eu-west-1"}, regionOpts)
}

func cleanupCommand(ctx context.Context, slug string) error {
	return testStore.DeactivateCommand(ctx, slug)
}
