package handler

import (
	"context"
	"net/http"

	"github.com/chiragguptadtu/trigger/internal/middleware"
	"github.com/chiragguptadtu/trigger/internal/store"
)

// JobEnqueuer abstracts River so handlers can be tested without it.
type JobEnqueuer interface {
	Enqueue(ctx context.Context, executionID string) error
}

type Config struct {
	JWTSecret     string
	TokenTTL      int    // seconds
	EncryptionKey []byte // 32-byte AES-256 key
}

type Server struct {
	store    *store.Queries
	config   Config
	enqueuer JobEnqueuer
	mux      *http.ServeMux
}

func NewServer(q *store.Queries, cfg Config, enqueuer JobEnqueuer) *Server {
	s := &Server{store: q, config: cfg, enqueuer: enqueuer, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /auth/login", s.handleLogin)

	// Commands + Executions — require authentication
	auth := middleware.Authenticate(s.config.JWTSecret)
	access := middleware.RequireCommandAccess(s.store)
	s.mux.HandleFunc("GET /commands", chain(auth, nil, s.handleListCommands))
	s.mux.HandleFunc("GET /commands/import-errors", chain(auth, nil, s.handleListImportErrors))
	s.mux.HandleFunc("GET /commands/{slug}", chain(auth, access, s.handleGetCommand))
	s.mux.HandleFunc("POST /commands/{slug}/executions", chain(auth, access, s.handleTriggerExecution))
	s.mux.HandleFunc("GET /executions/{id}", chain(auth, nil, s.handleGetExecution))
	s.mux.HandleFunc("GET /commands/{slug}/executions", chain(auth, access, s.handleListExecutions))

	// Admin: users
	s.mux.HandleFunc("GET /admin/users", s.requireAdmin(s.handleListUsers))
	s.mux.HandleFunc("POST /admin/users", s.requireAdmin(s.handleCreateUser))
	s.mux.HandleFunc("PATCH /admin/users/{id}", s.requireAdmin(s.handleUpdateUser))
	s.mux.HandleFunc("DELETE /admin/users/{id}", s.requireAdmin(s.handleDeactivateUser))

	// Admin: groups
	s.mux.HandleFunc("GET /admin/groups", s.requireAdmin(s.handleListGroups))
	s.mux.HandleFunc("POST /admin/groups", s.requireAdmin(s.handleCreateGroup))
	s.mux.HandleFunc("DELETE /admin/groups/{id}", s.requireAdmin(s.handleDeleteGroup))
	s.mux.HandleFunc("GET /admin/groups/{id}/members", s.requireAdmin(s.handleListGroupMembers))
	s.mux.HandleFunc("POST /admin/groups/{id}/members", s.requireAdmin(s.handleAddGroupMember))
	s.mux.HandleFunc("DELETE /admin/groups/{id}/members/{userID}", s.requireAdmin(s.handleRemoveGroupMember))

	// Admin: command permissions + import errors
	s.mux.HandleFunc("GET /admin/commands/import-errors", s.requireAdmin(s.handleListImportErrors))
	s.mux.HandleFunc("GET /admin/commands/{slug}/permissions", s.requireAdmin(s.handleListPermissions))
	s.mux.HandleFunc("POST /admin/commands/{slug}/permissions", s.requireAdmin(s.handleGrantPermission))
	s.mux.HandleFunc("DELETE /admin/commands/{slug}/permissions/{granteeType}/{granteeID}", s.requireAdmin(s.handleRevokePermission))

	// Admin: config
	s.mux.HandleFunc("POST /admin/config", s.requireAdmin(s.handleCreateConfig))
	s.mux.HandleFunc("GET /admin/config", s.requireAdmin(s.handleListConfig))
	s.mux.HandleFunc("PUT /admin/config/{key}", s.requireAdmin(s.handleUpdateConfig))
	s.mux.HandleFunc("DELETE /admin/config/{key}", s.requireAdmin(s.handleDeleteConfig))
}

// requireAdmin chains Authenticate → RequireAdmin around a handler.
func (s *Server) requireAdmin(h http.HandlerFunc) http.HandlerFunc {
	return middleware.Authenticate(s.config.JWTSecret)(middleware.RequireAdmin(h)).ServeHTTP
}

// chain applies Authenticate, then an optional second middleware, then the handler.
func chain(
	auth func(http.Handler) http.Handler,
	second func(http.Handler) http.Handler,
	h http.HandlerFunc,
) http.HandlerFunc {
	if second != nil {
		return auth(second(h)).ServeHTTP
	}
	return auth(h).ServeHTTP
}
