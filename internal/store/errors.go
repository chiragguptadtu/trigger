package store

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Sentinel errors returned by Normalize. Handlers import only these —
// no pgx packages needed outside the store package.
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// Normalize translates pgx-specific errors into store-level sentinels.
// Returns nil unchanged, wraps pgx.ErrNoRows → ErrNotFound,
// wraps unique-violation (23505) → ErrConflict, passes everything else through.
func Normalize(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}
