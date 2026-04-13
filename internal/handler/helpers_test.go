package handler_test

import (
	"context"
	"testing"

	"trigger/internal/auth"
	"trigger/internal/handler"
	"trigger/internal/middleware"
	"trigger/internal/store"
)

// testCleanup registers a t.Cleanup that runs the given SQL against testPool.
// Registered cleanups run in LIFO order — register parent rows first so child
// rows (which cascade or are registered later) are deleted first.
func testCleanup(t *testing.T, query string, args ...any) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := testPool.Exec(context.Background(), query, args...); err != nil {
			t.Logf("cleanup query failed: %v | query: %s", err, query)
		}
	})
}

// withClaims injects auth claims into a context — used by handler tests to
// bypass the Authenticate middleware.
func withClaims(ctx context.Context, c *auth.Claims) context.Context {
	return middleware.WithClaims(ctx, c)
}

var testEncryptionKey = make([]byte, 32) // all-zero key, fine for tests

func newTestServer(q *store.Queries) *handler.Server {
	return handler.NewServer(q, handler.Config{
		JWTSecret:     "test-secret",
		TokenTTL:      3600,
		EncryptionKey: testEncryptionKey,
	}, nil)
}

// createAdminParams returns a unique CreateUserParams suitable for test setup.
func createAdminParams() store.CreateUserParams {
	return store.CreateUserParams{
		Email:        uniqueEmail(),
		Name:         "Test Admin",
		PasswordHash: "hashed",
		IsAdmin:      false,
	}
}

// adminCtx creates a test admin user, registers cleanup, and returns a context
// with admin claims injected.
func adminCtx(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()
	u, err := testQueries.CreateUser(ctx, createAdminParams())
	if err != nil {
		t.Fatalf("adminCtx: create user: %v", err)
	}
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID)
	return withClaims(ctx, &auth.Claims{UserID: u.ID.String(), IsAdmin: true})
}
