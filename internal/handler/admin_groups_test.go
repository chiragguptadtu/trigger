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
	"github.com/chiragguptadtu/trigger/internal/store"
)

func uniqueGroupName() string {
	return fmt.Sprintf("group-%d", time.Now().UnixNano())
}

func TestCreateGroup(t *testing.T) {
	name := uniqueGroupName()
	testCleanup(t, "DELETE FROM groups WHERE name = $1", name)
	body := `{"name":"` + name + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/groups", strings.NewReader(body)).WithContext(adminCtx(t))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.NotEmpty(t, resp["id"])
}

func TestListGroups(t *testing.T) {
	ctx := context.Background()
	g, err := testQueries.CreateGroup(ctx, uniqueGroupName())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM groups WHERE id = $1", g.ID)

	req := httptest.NewRequest(http.MethodGet, "/admin/groups", nil).WithContext(adminCtx(t))
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var groups []map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&groups))
	assert.NotEmpty(t, groups)
}

func TestDeleteGroup(t *testing.T) {
	ctx := context.Background()
	g, err := testQueries.CreateGroup(ctx, uniqueGroupName())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM groups WHERE id = $1", g.ID)

	req := httptest.NewRequest(http.MethodDelete, "/admin/groups/"+g.ID.String(), nil).
		WithContext(adminCtx(t))
	req.SetPathValue("id", g.ID.String())
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestAddAndListGroupMembers(t *testing.T) {
	ctx := context.Background()
	g, err := testQueries.CreateGroup(ctx, uniqueGroupName())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM groups WHERE id = $1", g.ID) // cascades user_group_members

	u, err := testQueries.CreateUser(ctx, createAdminParams())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID) // cascades user_group_members

	body := `{"user_id":"` + u.ID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/groups/"+g.ID.String()+"/members",
		strings.NewReader(body)).WithContext(adminCtx(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", g.ID.String())
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	listReq := httptest.NewRequest(http.MethodGet, "/admin/groups/"+g.ID.String()+"/members", nil).
		WithContext(adminCtx(t))
	listReq.SetPathValue("id", g.ID.String())
	listRR := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(listRR, listReq)
	assert.Equal(t, http.StatusOK, listRR.Code)

	var members []map[string]any
	require.NoError(t, json.NewDecoder(listRR.Body).Decode(&members))
	require.Len(t, members, 1)
	assert.Equal(t, u.ID.String(), members[0]["id"])
}

func TestRemoveGroupMember(t *testing.T) {
	ctx := context.Background()
	g, err := testQueries.CreateGroup(ctx, uniqueGroupName())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM groups WHERE id = $1", g.ID) // cascades user_group_members

	u, err := testQueries.CreateUser(ctx, createAdminParams())
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", u.ID) // cascades user_group_members

	require.NoError(t, testQueries.AddGroupMember(ctx, store.AddGroupMemberParams{
		UserID: u.ID, GroupID: g.ID,
	}))

	req := httptest.NewRequest(http.MethodDelete,
		"/admin/groups/"+g.ID.String()+"/members/"+u.ID.String(), nil).
		WithContext(adminCtx(t))
	req.SetPathValue("id", g.ID.String())
	req.SetPathValue("userID", u.ID.String())
	rr := httptest.NewRecorder()
	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	members, err := testQueries.ListGroupMembers(ctx, g.ID)
	require.NoError(t, err)
	assert.Empty(t, members)
}
