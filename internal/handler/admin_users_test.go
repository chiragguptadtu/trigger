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
	"trigger/internal/auth"
)

func TestListUsers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil).WithContext(adminCtx(t))
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var users []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&users))
	assert.NotEmpty(t, users)
}

func TestCreateUser(t *testing.T) {
	email := uniqueEmail()
	testCleanup(t, "DELETE FROM users WHERE email = $1", email)
	body := `{"email":"` + email + `","name":"New User","password":"pass123","is_admin":false}`
	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(body)).WithContext(adminCtx(t))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.NotEmpty(t, resp["id"])
	assert.Nil(t, resp["password_hash"], "password_hash must never be returned")
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	email := uniqueEmail()
	testCleanup(t, "DELETE FROM users WHERE email = $1", email)
	body := `{"email":"` + email + `","name":"A","password":"p"}`
	srv := newTestServer(testQueries)
	srv.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(body)).WithContext(adminCtx(t)))

	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(body)).WithContext(adminCtx(t))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	u, err := testQueries.CreateUser(ctx, createAdminParams())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID)

	body := `{"name":"Updated Name","is_admin":true,"is_active":true}`
	req := httptest.NewRequest(http.MethodPatch, "/admin/users/"+u.ID.String(), strings.NewReader(body)).
		WithContext(adminCtx(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", u.ID.String())
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "Updated Name", resp["name"])
	assert.Equal(t, true, resp["is_admin"])
}

func TestDeactivateUser(t *testing.T) {
	ctx := context.Background()
	u, err := testQueries.CreateUser(ctx, createAdminParams())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+u.ID.String(), nil).
		WithContext(adminCtx(t))
	req.SetPathValue("id", u.ID.String())
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	fetched, err := testQueries.GetUserByID(ctx, u.ID)
	require.NoError(t, err)
	assert.False(t, fetched.IsActive)
}

func TestDeactivateUser_NotFound(t *testing.T) {
	fakeID := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+fakeID, nil).
		WithContext(adminCtx(t))
	req.SetPathValue("id", fakeID)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminUsers_OperatorForbidden(t *testing.T) {
	ctx := context.Background()
	u, err := testQueries.CreateUser(ctx, createAdminParams())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID)
	opCtx := withClaims(ctx, &auth.Claims{UserID: u.ID.String(), IsAdmin: false})

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil).WithContext(opCtx)
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
