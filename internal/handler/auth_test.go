package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chiragguptadtu/trigger/db"
	"github.com/chiragguptadtu/trigger/internal/auth"
	"github.com/chiragguptadtu/trigger/internal/store"
)

func uniqueEmail() string {
	return fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
}

const defaultTestDSN = "postgres://trigger:trigger@localhost:5432/trigger?sslmode=disable"

var (
	testPool    *pgxpool.Pool
	testQueries *store.Queries
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	if err := db.RunMigrations(dsn); err != nil {
		panic("migrations: " + err.Error())
	}
	pool, err := db.Connect(context.Background(), dsn)
	if err != nil {
		panic("db connect: " + err.Error())
	}
	defer pool.Close()
	testPool = pool
	testQueries = store.New(pool)
	os.Exit(m.Run())
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	hash, _ := auth.HashPassword("password123")
	email := uniqueEmail()
	user, err := testQueries.CreateUser(ctx, store.CreateUserParams{
		Email: email, Name: "Login User",
		PasswordHash: hash, IsAdmin: false,
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", user.ID)

	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	newTestServer(testQueries).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.NotEmpty(t, resp["token"])
}

func TestLogin_WrongPassword(t *testing.T) {
	ctx := context.Background()
	hash, _ := auth.HashPassword("correctpassword")
	email := uniqueEmail()
	user, err := testQueries.CreateUser(ctx, store.CreateUserParams{
		Email: email, Name: "Wrong PW",
		PasswordHash: hash, IsAdmin: false,
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", user.ID)

	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "wrongpassword",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogin_UnknownEmail(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"email": "nobody@example.com", "password": "whatever",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogin_InactiveUser(t *testing.T) {
	ctx := context.Background()
	hash, _ := auth.HashPassword("password123")
	email := uniqueEmail()
	user, err := testQueries.CreateUser(ctx, store.CreateUserParams{
		Email: email, Name: "Inactive",
		PasswordHash: hash, IsAdmin: false,
	})
	require.NoError(t, err)
	testCleanup(t, "DELETE FROM users WHERE id = $1", user.ID)
	require.NoError(t, testQueries.DeactivateUser(ctx, user.ID))

	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	newTestServer(testQueries).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
