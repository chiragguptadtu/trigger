package execution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/riverqueue/river"
	"trigger/internal/crypto"
	"trigger/internal/store"
)

// ExecutionArgs is the River job payload.
type ExecutionArgs struct {
	ExecutionID string `json:"execution_id"`
}

func (ExecutionArgs) Kind() string { return "execution" }

const (
	statusRunning = "running"
	statusSuccess = "success"
	statusFailure = "failure"
)

// Worker processes execution jobs: runs the script and updates the DB record.
type Worker struct {
	river.WorkerDefaults[ExecutionArgs]
	store         *store.Queries
	encryptionKey []byte
}

func NewWorker(q *store.Queries, encryptionKey []byte) *Worker {
	return &Worker{store: q, encryptionKey: encryptionKey}
}

func (w *Worker) Work(ctx context.Context, job *river.Job[ExecutionArgs]) error {
	execID, err := uuid.Parse(job.Args.ExecutionID)
	if err != nil {
		return fmt.Errorf("parse execution id: %w", err)
	}

	exec_, err := w.store.GetExecutionByID(ctx, execID)
	if err != nil {
		return fmt.Errorf("get execution: %w", err)
	}

	cmd, err := w.store.GetCommandByID(ctx, exec_.CommandID)
	if err != nil {
		return fmt.Errorf("get command: %w", err)
	}

	// Mark as running
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if _, err := w.store.UpdateExecutionStatus(ctx, store.UpdateExecutionStatusParams{
		ID: execID, Status: statusRunning, StartedAt: now,
	}); err != nil {
		return fmt.Errorf("update running: %w", err)
	}

	// Build config map from DB (decrypted).
	// loadConfig is best-effort: entries that disappear between list and fetch
	// (e.g. concurrent delete) are silently skipped.
	config, err := w.loadConfig(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Unmarshal inputs
	var inputs map[string]any
	if err := json.Unmarshal(exec_.Inputs, &inputs); err != nil {
		return fmt.Errorf("unmarshal inputs: %w", err)
	}

	// Run the script
	errMsg, runErr := RunScript(ctx, cmd.ScriptPath, inputs, config)

	completedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	status := statusSuccess
	if runErr != nil || errMsg != "" {
		status = statusFailure
		if runErr != nil {
			errMsg = runErr.Error()
		}
	}

	if _, err := w.store.UpdateExecutionStatus(ctx, store.UpdateExecutionStatusParams{
		ID:           execID,
		Status:       status,
		ErrorMessage: errMsg,
		StartedAt:    now,
		CompletedAt:  completedAt,
	}); err != nil {
		return fmt.Errorf("update completed: %w", err)
	}

	// Do not return an error for script failures — River retries on errors,
	// but a script failure is a terminal outcome not a transient one.
	return nil
}

func (w *Worker) loadConfig(ctx context.Context) (map[string]string, error) {
	entries, err := w.store.ListConfigEntries(ctx)
	if err != nil {
		return nil, err
	}

	// ListConfigEntries returns rows without value_encrypted (safe for API).
	// We need full entries for decryption — fetch them individually.
	result := make(map[string]string, len(entries))
	for _, e := range entries {
		full, err := w.store.GetConfigEntryByKey(ctx, e.Key)
		if err != nil {
			if errors.Is(store.Normalize(err), store.ErrNotFound) {
				// Entry was deleted between the list call and this fetch. Skip it.
				continue
			}
			return nil, fmt.Errorf("get config %q: %w", e.Key, err)
		}
		plaintext, err := crypto.Decrypt(w.encryptionKey, full.ValueEncrypted)
		if err != nil {
			return nil, fmt.Errorf("decrypt config %q: %w", e.Key, err)
		}
		result[e.Key] = plaintext
	}
	return result, nil
}

