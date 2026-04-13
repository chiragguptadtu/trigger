package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"trigger/internal/auth"
	"trigger/internal/store"
)

type createUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type updateUserRequest struct {
	Name     *string `json:"name"`
	Password *string `json:"password"`
	IsAdmin  *bool   `json:"is_admin"`
	IsActive *bool   `json:"is_active"`
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	result := make([]UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, newUserResponse(u))
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	user, err := s.store.CreateUser(r.Context(), store.CreateUserParams{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: hash,
		IsAdmin:      req.IsAdmin,
	})
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrConflict) {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, newUserResponse(user))
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	current, err := s.store.GetUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(store.Normalize(err), store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Reset password if requested.
	if req.Password != nil {
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if err := s.store.UpdateUserPassword(r.Context(), store.UpdateUserPasswordParams{
			ID:           id,
			PasswordHash: hash,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
	}

	// Apply only the fields that were explicitly provided.
	params := store.UpdateUserParams{
		ID:       id,
		Name:     current.Name,
		IsAdmin:  current.IsAdmin,
		IsActive: current.IsActive,
	}
	if req.Name != nil {
		params.Name = *req.Name
	}
	if req.IsAdmin != nil {
		params.IsAdmin = *req.IsAdmin
	}
	if req.IsActive != nil {
		params.IsActive = *req.IsActive
	}

	user, err := s.store.UpdateUser(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(user))
}

func (s *Server) handleDeactivateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := s.store.DeactivateUser(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
