package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"trigger/internal/auth"
	"trigger/internal/store"
)

func uniqueKey(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// adminRequest builds a request with admin claims injected and registers cleanup
// for the admin user it creates.
func adminRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	adminUser, err := testQueries.CreateUser(context.Background(), store.CreateUserParams{
		Email: uniqueEmail(), Name: "Admin", PasswordHash: "hashed", IsAdmin: true,
	})
	require.NoError(t, err)
	// config_entries.created_by references this user — config entry cleanups must be
	// registered AFTER this so they run FIRST (LIFO order).
	testCleanup(t, "DELETE FROM users WHERE id = $1", adminUser.ID)

	claims := &auth.Claims{UserID: adminUser.ID.String(), IsAdmin: true}
	return req.WithContext(withClaims(req.Context(), claims))
}

// operatorRequest builds a request with operator claims injected and registers
// cleanup for the operator user it creates.
func operatorRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	opUser, err := testQueries.CreateUser(context.Background(), store.CreateUserParams{
		Email: uniqueEmail(), Name: "Op", PasswordHash: "hashed", IsAdmin: false,
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", opUser.ID)

	claims := &auth.Claims{UserID: opUser.ID.String(), IsAdmin: false}
	return req.WithContext(withClaims(req.Context(), claims))
}

func TestCreateConfigEntry(t *testing.T) {
	key := uniqueKey("DB_PASSWORD")
	req := adminRequest(t, http.MethodPost, "/admin/config", map[string]string{
		"key": key, "value": "s3cr3t", "description": "main db password",
	})
	// Registered after adminRequest's user cleanup — runs first (LIFO), before user delete.
	testCleanup(t, "DELETE FROM config_entries WHERE key = $1", key)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, key, resp["key"])
	assert.Nil(t, resp["value"], "value must never be returned")
}

func TestCreateConfigEntry_DuplicateKey(t *testing.T) {
	key := uniqueKey("DUP_KEY")
	req1 := adminRequest(t, http.MethodPost, "/admin/config", map[string]string{
		"key": key, "value": "v1",
	})
	testCleanup(t, "DELETE FROM config_entries WHERE key = $1", key)
	newTestServer(testQueries).ServeHTTP(httptest.NewRecorder(), req1)

	req2 := adminRequest(t, http.MethodPost, "/admin/config", map[string]string{
		"key": key, "value": "v2",
	})
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req2)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestListConfigEntries_NoValues(t *testing.T) {
	key := uniqueKey("LIST_TEST")
	req := adminRequest(t, http.MethodPost, "/admin/config", map[string]string{
		"key": key, "value": "hidden",
	})
	testCleanup(t, "DELETE FROM config_entries WHERE key = $1", key)
	newTestServer(testQueries).ServeHTTP(httptest.NewRecorder(), req)

	listReq := adminRequest(t, http.MethodGet, "/admin/config", nil)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, listReq)
	assert.Equal(t, http.StatusOK, rr.Code)

	var entries []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&entries))
	for _, e := range entries {
		assert.Nil(t, e["value"], "value must never appear in list response")
	}
}

func TestUpdateConfigEntry(t *testing.T) {
	key := uniqueKey("UPDATE_KEY")
	createReq := adminRequest(t, http.MethodPost, "/admin/config", map[string]string{
		"key": key, "value": "old",
	})
	testCleanup(t, "DELETE FROM config_entries WHERE key = $1", key)
	newTestServer(testQueries).ServeHTTP(httptest.NewRecorder(), createReq)

	updateReq := adminRequest(t, http.MethodPut, "/admin/config/"+key, map[string]string{
		"value": "new", "description": "updated",
	})
	updateReq.SetPathValue("key", key)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, updateReq)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDeleteConfigEntry(t *testing.T) {
	key := uniqueKey("DELETE_KEY")
	createReq := adminRequest(t, http.MethodPost, "/admin/config", map[string]string{
		"key": key, "value": "val",
	})
	// Safety net: if the delete handler fails mid-test, this cleanup removes the entry
	// so the user cleanup (which runs after) doesn't hit a FK violation.
	testCleanup(t, "DELETE FROM config_entries WHERE key = $1", key)
	newTestServer(testQueries).ServeHTTP(httptest.NewRecorder(), createReq)

	delReq := adminRequest(t, http.MethodDelete, "/admin/config/"+key, nil)
	delReq.SetPathValue("key", key)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, delReq)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestConfigEntry_OperatorForbidden(t *testing.T) {
	req := operatorRequest(t, http.MethodGet, "/admin/config", nil)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
