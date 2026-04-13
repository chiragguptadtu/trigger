package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/db"
	"trigger/internal/auth"
	"trigger/internal/middleware"
	"trigger/internal/store"
)

const defaultTestDSN = "postgres://trigger:trigger@localhost:5432/trigger?sslmode=disable"

var testStore *store.Queries

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
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

func uniqueEmail() string {
	return fmt.Sprintf("mw-%d@example.com", time.Now().UnixNano())
}

func TestRequireCommandAccess_AdminBypasses(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	admin, err := testStore.CreateUser(ctx, store.CreateUserParams{
		Email: uniqueEmail(), Name: "Admin", PasswordHash: "hashed", IsAdmin: true,
	})
	require.NoError(t, err)

	cmd, err := testStore.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: "access-admin-bypass", Name: "Admin Bypass", ScriptPath: "/c.py",
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = testStore.DeactivateCommand(ctx, cmd.Slug) })

	reqCtx := middleware.WithClaims(context.Background(), &auth.Claims{UserID: admin.ID.String(), IsAdmin: true})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(reqCtx)
	req.SetPathValue("slug", "access-admin-bypass")
	rr := httptest.NewRecorder()

	middleware.RequireCommandAccess(testStore)(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireCommandAccess_OperatorWithGrant(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	user, err := testStore.CreateUser(ctx, store.CreateUserParams{
		Email: uniqueEmail(), Name: "Op", PasswordHash: "hashed", IsAdmin: false,
	})
	require.NoError(t, err)

	cmd, err := testStore.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: "access-op-granted", Name: "Granted", ScriptPath: "/c.py",
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = testStore.DeactivateCommand(ctx, cmd.Slug) })

	_, err = testStore.CreateCommandPermission(ctx, store.CreateCommandPermissionParams{
		CommandID: cmd.ID, GranteeType: "user", GranteeID: user.ID,
	})
	require.NoError(t, err)

	reqCtx := middleware.WithClaims(context.Background(), &auth.Claims{UserID: user.ID.String(), IsAdmin: false})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(reqCtx)
	req.SetPathValue("slug", "access-op-granted")
	rr := httptest.NewRecorder()

	middleware.RequireCommandAccess(testStore)(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireCommandAccess_OperatorWithoutGrant(t *testing.T) {
	skipIfNoDB(t)
	ctx := context.Background()

	user, err := testStore.CreateUser(ctx, store.CreateUserParams{
		Email: uniqueEmail(), Name: "Op", PasswordHash: "hashed", IsAdmin: false,
	})
	require.NoError(t, err)

	cmd, err := testStore.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: "access-op-denied", Name: "Denied", ScriptPath: "/c.py",
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = testStore.DeactivateCommand(ctx, cmd.Slug) })

	reqCtx := middleware.WithClaims(context.Background(), &auth.Claims{UserID: user.ID.String(), IsAdmin: false})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(reqCtx)
	req.SetPathValue("slug", "access-op-denied")
	rr := httptest.NewRecorder()

	middleware.RequireCommandAccess(testStore)(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
