package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/chiragguptadtu/trigger/internal/auth"
	"github.com/chiragguptadtu/trigger/internal/middleware"
)

const testSecret = "test-secret"

func tokenFor(t *testing.T, userID string, isAdmin bool) string {
	t.Helper()
	tok, err := auth.GenerateToken(userID, isAdmin, testSecret, time.Hour)
	require.NoError(t, err)
	return tok
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// --- Authenticate middleware ---

func TestAuthenticate_ValidToken(t *testing.T) {
	tok := tokenFor(t, "user-1", false)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()

	middleware.Authenticate(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "user-1", claims.UserID)
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthenticate_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	middleware.Authenticate(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.token")
	rr := httptest.NewRecorder()
	middleware.Authenticate(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthenticate_ExpiredToken(t *testing.T) {
	tok, _ := auth.GenerateToken("user-1", false, testSecret, -time.Second)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	middleware.Authenticate(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthenticate_MalformedHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc123") // wrong scheme
	rr := httptest.NewRecorder()
	middleware.Authenticate(testSecret)(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- RequireAdmin middleware ---

func TestRequireAdmin_AdminUser(t *testing.T) {
	ctx := middleware.WithClaims(context.Background(), &auth.Claims{UserID: "admin-1", IsAdmin: true})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	middleware.RequireAdmin(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireAdmin_OperatorUser(t *testing.T) {
	ctx := middleware.WithClaims(context.Background(), &auth.Claims{UserID: "user-1", IsAdmin: false})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	middleware.RequireAdmin(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireAdmin_NoClaims(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	middleware.RequireAdmin(http.HandlerFunc(okHandler)).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
