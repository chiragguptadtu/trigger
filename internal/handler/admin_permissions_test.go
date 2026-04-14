package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/internal/store"
)

func TestListImportErrors(t *testing.T) {
	ctx := context.Background()
	filename := "bad_command_" + uniqueSlug("x") + ".py"
	err := testQueries.UpsertImportError(ctx, store.UpsertImportErrorParams{
		Filename: filename,
		Error:    "malformed YAML header",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM command_import_errors WHERE filename = $1", filename)

	req := httptest.NewRequest(http.MethodGet, "/admin/commands/import-errors", nil).
		WithContext(adminCtx(t))
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	// Find our entry — other import errors may exist from concurrent scans.
	var found bool
	for _, e := range resp {
		if e["filename"] == filename {
			assert.Equal(t, "malformed YAML header", e["error"])
			assert.NotEmpty(t, e["failed_at"])
			found = true
		}
	}
	assert.True(t, found, "expected import error for %s in response", filename)
}

func TestListCommandPermissions(t *testing.T) {
	ctx := context.Background()
	cmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: uniqueSlug("perm-list"), Name: "Perm List", ScriptPath: "/c.py",
	})
	require.NoError(t, err)
	// command_permissions and command_inputs cascade on command delete
	testCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID)

	req := httptest.NewRequest(http.MethodGet, "/admin/commands/"+cmd.Slug+"/permissions", nil).
		WithContext(adminCtx(t))
	req.SetPathValue("slug", cmd.Slug)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var perms []any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&perms))
	assert.NotNil(t, perms)
}

func TestGrantCommandPermission(t *testing.T) {
	ctx := context.Background()
	cmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: uniqueSlug("perm-grant"), Name: "Perm Grant", ScriptPath: "/c.py",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID) // cascades permissions

	u, err := testQueries.CreateUser(ctx, createAdminParams())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID)

	body := `{"grantee_type":"user","grantee_id":"` + u.ID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/commands/"+cmd.Slug+"/permissions",
		strings.NewReader(body)).WithContext(adminCtx(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("slug", cmd.Slug)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestGrantCommandPermission_Duplicate(t *testing.T) {
	ctx := context.Background()
	cmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: uniqueSlug("perm-dup"), Name: "Perm Dup", ScriptPath: "/c.py",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID)

	u, err := testQueries.CreateUser(ctx, createAdminParams())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID)

	body := `{"grantee_type":"user","grantee_id":"` + u.ID.String() + `"}`
	srv := newTestServer(testQueries)

	// First grant — should succeed.
	req1 := httptest.NewRequest(http.MethodPost, "/admin/commands/"+cmd.Slug+"/permissions",
		strings.NewReader(body)).WithContext(adminCtx(t))
	req1.Header.Set("Content-Type", "application/json")
	req1.SetPathValue("slug", cmd.Slug)
	srv.ServeHTTP(httptest.NewRecorder(), req1)

	// Second grant — should return 409.
	req2 := httptest.NewRequest(http.MethodPost, "/admin/commands/"+cmd.Slug+"/permissions",
		strings.NewReader(body)).WithContext(adminCtx(t))
	req2.Header.Set("Content-Type", "application/json")
	req2.SetPathValue("slug", cmd.Slug)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req2)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestRevokeCommandPermission(t *testing.T) {
	ctx := context.Background()
	cmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: uniqueSlug("perm-revoke"), Name: "Perm Revoke", ScriptPath: "/c.py",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID) // cascades permissions

	u, err := testQueries.CreateUser(ctx, createAdminParams())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID)

	_, err = testQueries.CreateCommandPermission(ctx, store.CreateCommandPermissionParams{
		CommandID: cmd.ID, GranteeType: "user", GranteeID: u.ID,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete,
		"/admin/commands/"+cmd.Slug+"/permissions/user/"+u.ID.String(), nil).
		WithContext(adminCtx(t))
	req.SetPathValue("slug", cmd.Slug)
	req.SetPathValue("granteeType", "user")
	req.SetPathValue("granteeID", u.ID.String())
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

