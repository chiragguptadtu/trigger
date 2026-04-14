package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/internal/auth"
	"trigger/internal/store"
)

func TestListCommands_Admin(t *testing.T) {
	ctx := context.Background()
	slug := uniqueSlug("list-admin")
	cmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: slug, Name: "Admin Visible Cmd", ScriptPath: "/x.py",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID)

	req := httptest.NewRequest(http.MethodGet, "/commands", nil).
		WithContext(withClaims(ctx, &auth.Claims{IsAdmin: true}))
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var cmds []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cmds))
	var found bool
	for _, c := range cmds {
		if c["slug"] == slug {
			found = true
		}
	}
	assert.True(t, found, "admin should see all commands including %s", slug)
}

func TestListCommands_Operator_OnlyGranted(t *testing.T) {
	ctx := context.Background()

	// Create operator user.
	opUser, err := testQueries.CreateUser(ctx, store.CreateUserParams{
		Email: uniqueEmail(), Name: "Op", PasswordHash: "hashed",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", opUser.ID)

	// Command with access.
	grantedSlug := uniqueSlug("granted")
	grantedCmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: grantedSlug, Name: "Granted Cmd", ScriptPath: "/a.py",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM commands WHERE id = $1", grantedCmd.ID)
	_, err = testQueries.CreateCommandPermission(ctx, store.CreateCommandPermissionParams{
		CommandID: grantedCmd.ID, GranteeType: "user", GranteeID: opUser.ID,
	})
	require.NoError(t, err)

	// Command without access.
	deniedSlug := uniqueSlug("denied")
	deniedCmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: deniedSlug, Name: "Denied Cmd", ScriptPath: "/b.py",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM commands WHERE id = $1", deniedCmd.ID)

	req := httptest.NewRequest(http.MethodGet, "/commands", nil).
		WithContext(withClaims(ctx, &auth.Claims{UserID: opUser.ID.String(), IsAdmin: false}))
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var cmds []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cmds))
	slugs := make([]string, 0, len(cmds))
	for _, c := range cmds {
		slugs = append(slugs, c["slug"].(string))
	}
	assert.Contains(t, slugs, grantedSlug, "operator should see granted command")
	assert.NotContains(t, slugs, deniedSlug, "operator should not see unganted command")
}

func TestGetCommand_Success(t *testing.T) {
	ctx := context.Background()
	slug := uniqueSlug("get-cmd")
	cmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: slug, Name: "Get Cmd", ScriptPath: "/x.py",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID)

	_, err = testQueries.CreateCommandInput(ctx, store.CreateCommandInputParams{
		CommandID: cmd.ID, Name: "env", Label: "Env",
		Type: "closed", Options: []byte(`["staging","prod"]`), Required: true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/commands/"+slug, nil).
		WithContext(withClaims(ctx, &auth.Claims{IsAdmin: true}))
	req.SetPathValue("slug", slug)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, slug, resp["slug"])
	inputs := resp["inputs"].([]any)
	assert.Len(t, inputs, 1)
}

func TestGetCommand_NotFound(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest(http.MethodGet, "/commands/does-not-exist", nil).
		WithContext(withClaims(ctx, &auth.Claims{IsAdmin: true}))
	req.SetPathValue("slug", "does-not-exist")
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetCommand_OperatorDenied(t *testing.T) {
	ctx := context.Background()

	opUser, err := testQueries.CreateUser(ctx, store.CreateUserParams{
		Email: uniqueEmail(), Name: "Op", PasswordHash: "hashed",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", opUser.ID)

	slug := uniqueSlug("op-denied")
	cmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: slug, Name: "Op Denied Cmd", ScriptPath: "/x.py",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID)

	req := httptest.NewRequest(http.MethodGet, "/commands/"+slug, nil).
		WithContext(withClaims(ctx, &auth.Claims{UserID: opUser.ID.String(), IsAdmin: false}))
	req.SetPathValue("slug", slug)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
