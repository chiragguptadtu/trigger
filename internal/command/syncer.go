package command

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"trigger/internal/store"
)

// Sync upserts all scanned commands into the DB and deactivates any active
// commands that are no longer present in the scan results.
func Sync(ctx context.Context, q *store.Queries, scanned []ScannedCommand) error {
	activeSlug := make(map[string]bool, len(scanned))

	for _, sc := range scanned {
		cmd, err := q.UpsertCommand(ctx, store.UpsertCommandParams{
			Slug:        sc.Slug,
			Name:        sc.Command.Name,
			Description: sc.Command.Description,
			ScriptPath:  sc.ScriptPath,
		})
		if err != nil {
			return err
		}

		// Replace inputs atomically: delete then re-insert.
		if err := q.DeleteCommandInputs(ctx, cmd.ID); err != nil {
			return err
		}
		for i, input := range sc.Command.Inputs {
			var options []byte
			if len(input.Options) > 0 {
				options, err = json.Marshal(input.Options)
				if err != nil {
					return err
				}
			}
			_, err = q.CreateCommandInput(ctx, store.CreateCommandInputParams{
				CommandID: cmd.ID,
				Name:      input.Name,
				Label:     input.Label,
				Type:      string(input.Type),
				Options:   options,
				Multi:     input.Multi,
				Required:  input.Required,
				SortOrder: int32(i),
			})
			if err != nil {
				return err
			}
		}

		activeSlug[sc.Slug] = true
	}

	// Deactivate commands that were not in this scan.
	allCmds, err := q.ListAllCommands(ctx)
	if err != nil {
		return err
	}
	for _, cmd := range allCmds {
		if !activeSlug[cmd.Slug] {
			if err := q.DeactivateCommand(ctx, cmd.Slug); err != nil {
				return err
			}
		}
	}

	return nil
}

// ScanLoop periodically scans dir and syncs commands to the DB.
// An initial scan runs immediately on the first iteration; subsequent scans
// run every interval. Per-file parse errors are persisted to command_import_errors
// so the frontend can surface them. Resolved errors (file now parses cleanly) are
// deleted. The loop exits when ctx is cancelled.
func ScanLoop(ctx context.Context, dir string, q *store.Queries, interval time.Duration) {
	for {
		result, err := ScanDir(dir)
		if err != nil {
			log.Printf("command scan error: %v", err)
		} else {
			if err := Sync(ctx, q, result.Commands); err != nil {
				log.Printf("command sync error: %v", err)
			} else {
				log.Printf("commands: synced %d, import errors: %d", len(result.Commands), len(result.Errors))
			}

			// Replace the import error table with this scan's current errors.
			// Clearing first ensures stale entries (deleted files, fixed files) are removed.
			if err := q.ClearImportErrors(ctx); err != nil {
				log.Printf("clear import errors: %v", err)
			}
			for _, se := range result.Errors {
				log.Printf("command import error in %s: %v", se.Filename, se.Err)
				if err := q.UpsertImportError(ctx, store.UpsertImportErrorParams{
					Filename: se.Filename,
					Error:    se.Err.Error(),
				}); err != nil {
					log.Printf("store import error for %s: %v", se.Filename, err)
				}
			}
		}

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return
		}
	}
}

// pgBytes wraps a byte slice as a pgtype-compatible value for JSONB columns.
func pgBytes(b []byte) pgtype.Text {
	if b == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(b), Valid: true}
}
