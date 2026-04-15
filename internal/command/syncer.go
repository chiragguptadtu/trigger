package command

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/chiragguptadtu/trigger/internal/crypto"
	"github.com/chiragguptadtu/trigger/internal/execution"
	"github.com/chiragguptadtu/trigger/internal/store"
)

// Sync upserts all scanned commands into the DB and deactivates any active
// commands that are no longer present in the scan results. encKey is the
// AES-256 key used to decrypt config entries passed to dynamic inputs; it may
// be nil if no commands in this scan have dynamic inputs.
func Sync(ctx context.Context, q *store.Queries, scanned []ScannedCommand, encKey []byte) error {
	activeSlug := make(map[string]bool, len(scanned))

	// Load decrypted config once per sync cycle — only if any dynamic input exists.
	var configCache map[string]string
	configLoaded := false
	loadConfig := func() (map[string]string, error) {
		if configLoaded {
			return configCache, nil
		}
		m, err := loadDecryptedConfig(ctx, q, encKey)
		if err != nil {
			return nil, err
		}
		configCache = m
		configLoaded = true
		return configCache, nil
	}

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
			resolvedOptions := input.Options
			if input.Dynamic {
				cfg, err := loadConfig()
				if err != nil {
					log.Printf("get_options: load config for %s/%s: %v", sc.Slug, input.Name, err)
				} else {
					opts, err := execution.GetOptions(ctx, sc.ScriptPath, input.Name, cfg)
					if err != nil {
						log.Printf("get_options: %s/%s: %v", sc.Slug, input.Name, err)
					} else {
						resolvedOptions = opts
					}
				}
			}

			var options []byte
			if len(resolvedOptions) > 0 {
				options, err = json.Marshal(resolvedOptions)
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
func ScanLoop(ctx context.Context, dir string, q *store.Queries, encKey []byte, interval time.Duration) {
	for {
		result, err := ScanDir(dir)
		if err != nil {
			log.Printf("command scan error: %v", err)
		} else {
			if err := Sync(ctx, q, result.Commands, encKey); err != nil {
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

// loadDecryptedConfig fetches all config entries from the DB and decrypts their
// values. Entries deleted between list and fetch are silently skipped.
func loadDecryptedConfig(ctx context.Context, q *store.Queries, encKey []byte) (map[string]string, error) {
	entries, err := q.ListConfigEntries(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(entries))
	for _, e := range entries {
		full, err := q.GetConfigEntryByKey(ctx, e.Key)
		if err != nil {
			if errors.Is(store.Normalize(err), store.ErrNotFound) {
				continue
			}
			return nil, err
		}
		if len(encKey) == 0 {
			// No key provided (e.g. tests with no config entries) — skip decryption.
			continue
		}
		plaintext, err := crypto.Decrypt(encKey, full.ValueEncrypted)
		if err != nil {
			log.Printf("decrypt config %q: %v", e.Key, err)
			continue
		}
		result[e.Key] = plaintext
	}
	return result, nil
}

// pgBytes wraps a byte slice as a pgtype-compatible value for JSONB columns.
func pgBytes(b []byte) pgtype.Text {
	if b == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(b), Valid: true}
}
