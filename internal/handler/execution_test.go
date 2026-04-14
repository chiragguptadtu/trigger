package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/internal/auth"
	"trigger/internal/handler"
	"trigger/internal/store"
)

// stubEnqueuer records calls but does nothing — keeps handler tests free of River.
type stubEnqueuer struct{ called bool }

func (s *stubEnqueuer) Enqueue(_ context.Context, _ string) error {
	s.called = true
	return nil
}

// Ensure stubEnqueuer satisfies the interface.
var _ handler.JobEnqueuer = (*stubEnqueuer)(nil)

func newTestServerWithEnqueuer(q *store.Queries, enq handler.JobEnqueuer) *handler.Server {
	return handler.NewServer(q, handler.Config{
		JWTSecret:     "test-secret",
		TokenTTL:      3600,
		EncryptionKey: testEncryptionKey,
	}, enq)
}

// setupCommandWithAccess creates a user, command, and permission for the user.
// It registers cleanups in dependency order (LIFO ensures correct deletion order):
//   - user cleanup registered first (runs last)
//   - command cleanup registered second (runs second-to-last; cascades inputs+permissions)
//
// Callers that create executions must register execution cleanup AFTER calling
// this function so it runs before command/user cleanup.
func setupCommandWithAccess(t *testing.T, ctx context.Context, slug string) (store.Command, store.User) {
	t.Helper()
	user, err := testQueries.CreateUser(ctx, store.CreateUserParams{
		Email: uniqueEmail(), Name: "Exec User", PasswordHash: "hashed",
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", user.ID) // registered 1st — runs last

	cmd, err := testQueries.UpsertCommand(ctx, store.UpsertCommandParams{
		Slug: slug, Name: "Exec Cmd", ScriptPath: "/cmd.py",
	})
	require.NoError(t, err)
	// cascades command_inputs and command_permissions; no executions yet at this point
	testCleanup(t, "DELETE FROM commands WHERE id = $1", cmd.ID) // registered 2nd — runs 2nd-to-last

	_, err = testQueries.CreateCommandPermission(ctx, store.CreateCommandPermissionParams{
		CommandID: cmd.ID, GranteeType: "user", GranteeID: user.ID,
	})
	require.NoError(t, err)

	return cmd, user
}

func TestTriggerExecution_Success(t *testing.T) {
	ctx := context.Background()
	cmd, user := setupCommandWithAccess(t, ctx, uniqueSlug("trigger-ok"))
	// executions must be deleted before the command — register after setupCommandWithAccess (LIFO runs first)
	testCleanup(t, "DELETE FROM executions WHERE command_id = $1", cmd.ID)

	_, err := testQueries.CreateCommandInput(ctx, store.CreateCommandInputParams{
		CommandID: cmd.ID, Name: "env", Label: "Env",
		Type: "closed", Options: []byte(`["staging","prod"]`), Required: true,
	})
	require.NoError(t, err)

	enq := &stubEnqueuer{}
	body := `{"inputs":{"env":"staging"}}`
	req := httptest.NewRequest(http.MethodPost, "/commands/"+cmd.Slug+"/executions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("slug", cmd.Slug)
	reqCtx := withClaims(req.Context(), &auth.Claims{UserID: user.ID.String(), IsAdmin: false})
	req = req.WithContext(reqCtx)

	rr := httptest.NewRecorder()
	newTestServerWithEnqueuer(testQueries, enq).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
	assert.True(t, enq.called)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.NotEmpty(t, resp["id"])
	assert.Equal(t, "pending", resp["status"])
}

func TestTriggerExecution_MissingRequiredInput(t *testing.T) {
	ctx := context.Background()
	cmd, user := setupCommandWithAccess(t, ctx, uniqueSlug("trigger-missing"))

	_, err := testQueries.CreateCommandInput(ctx, store.CreateCommandInputParams{
		CommandID: cmd.ID, Name: "env", Label: "Env",
		Type: "open", Required: true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/commands/"+cmd.Slug+"/executions", strings.NewReader(`{"inputs":{}}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("slug", cmd.Slug)
	req = req.WithContext(withClaims(req.Context(), &auth.Claims{UserID: user.ID.String()}))

	rr := httptest.NewRecorder()
	newTestServerWithEnqueuer(testQueries, &stubEnqueuer{}).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestTriggerExecution_InvalidClosedOption(t *testing.T) {
	ctx := context.Background()
	cmd, user := setupCommandWithAccess(t, ctx, uniqueSlug("trigger-invalid-opt"))

	_, err := testQueries.CreateCommandInput(ctx, store.CreateCommandInputParams{
		CommandID: cmd.ID, Name: "env", Label: "Env",
		Type: "closed", Options: []byte(`["staging","prod"]`), Required: true,
	})
	require.NoError(t, err)

	body := `{"inputs":{"env":"nope"}}`
	req := httptest.NewRequest(http.MethodPost, "/commands/"+cmd.Slug+"/executions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("slug", cmd.Slug)
	req = req.WithContext(withClaims(req.Context(), &auth.Claims{UserID: user.ID.String()}))

	rr := httptest.NewRecorder()
	newTestServerWithEnqueuer(testQueries, &stubEnqueuer{}).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetExecution(t *testing.T) {
	ctx := context.Background()
	cmd, user := setupCommandWithAccess(t, ctx, uniqueSlug("get-exec"))
	testCleanup(t, "DELETE FROM executions WHERE command_id = $1", cmd.ID)

	exec_, err := testQueries.CreateExecution(ctx, store.CreateExecutionParams{
		CommandID: cmd.ID, TriggeredBy: user.ID, Inputs: []byte("{}"),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/executions/"+exec_.ID.String(), nil)
	req.SetPathValue("id", exec_.ID.String())
	req = req.WithContext(withClaims(req.Context(), &auth.Claims{UserID: user.ID.String()}))

	rr := httptest.NewRecorder()
	newTestServerWithEnqueuer(testQueries, &stubEnqueuer{}).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "pending", resp["status"])
}

func TestListExecutions(t *testing.T) {
	ctx := context.Background()
	cmd, user := setupCommandWithAccess(t, ctx, uniqueSlug("list-exec"))
	testCleanup(t, "DELETE FROM executions WHERE command_id = $1", cmd.ID)

	for i := 0; i < 3; i++ {
		_, err := testQueries.CreateExecution(ctx, store.CreateExecutionParams{
			CommandID: cmd.ID, TriggeredBy: user.ID, Inputs: []byte("{}"),
		})
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/commands/"+cmd.Slug+"/executions", nil)
	req.SetPathValue("slug", cmd.Slug)
	req = req.WithContext(withClaims(req.Context(), &auth.Claims{UserID: user.ID.String()}))

	rr := httptest.NewRecorder()
	newTestServerWithEnqueuer(testQueries, &stubEnqueuer{}).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var executions []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&executions))
	assert.GreaterOrEqual(t, len(executions), 3)
}

func TestTriggerExecution_InvalidMultiOption(t *testing.T) {
	ctx := context.Background()
	cmd, user := setupCommandWithAccess(t, ctx, uniqueSlug("trigger-multi-invalid"))

	_, err := testQueries.CreateCommandInput(ctx, store.CreateCommandInputParams{
		CommandID: cmd.ID, Name: "envs", Label: "Envs",
		Type:  "closed",
		Multi: true,
		Options: []byte(`["staging","prod"]`), Required: true,
	})
	require.NoError(t, err)

	body := `{"inputs":{"envs":["staging","invalid-env"]}}`
	req := httptest.NewRequest(http.MethodPost, "/commands/"+cmd.Slug+"/executions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("slug", cmd.Slug)
	req = req.WithContext(withClaims(req.Context(), &auth.Claims{UserID: user.ID.String()}))

	rr := httptest.NewRecorder()
	newTestServerWithEnqueuer(testQueries, &stubEnqueuer{}).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func uniqueSlug(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
