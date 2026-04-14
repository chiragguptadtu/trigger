package handler

import (
	"encoding/json"

	"github.com/chiragguptadtu/trigger/internal/store"
)

const timeFormat = "2006-01-02T15:04:05Z"

// UserResponse is the JSON shape returned for a user.
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsAdmin   bool   `json:"is_admin"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func newUserResponse(u store.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		IsAdmin:   u.IsAdmin,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt.Time.Format(timeFormat),
	}
}

// GroupResponse is the JSON shape returned for a group.
type GroupResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func newGroupResponse(g store.Group) GroupResponse {
	return GroupResponse{
		ID:        g.ID.String(),
		Name:      g.Name,
		IsActive:  g.IsActive,
		CreatedAt: g.CreatedAt.Time.Format(timeFormat),
	}
}

// PermissionResponse is the JSON shape returned for a command permission.
type PermissionResponse struct {
	ID          string `json:"id"`
	CommandID   string `json:"command_id"`
	GranteeType string `json:"grantee_type"`
	GranteeID   string `json:"grantee_id"`
}

func newPermissionResponse(p store.CommandPermission) PermissionResponse {
	return PermissionResponse{
		ID:          p.ID.String(),
		CommandID:   p.CommandID.String(),
		GranteeType: p.GranteeType,
		GranteeID:   p.GranteeID.String(),
	}
}

// ExecutionResponse is the JSON shape returned for an execution.
// StartedAt and CompletedAt are omitted when not yet set.
// TriggeredByName/Email and Inputs are populated only in list responses.
type ExecutionResponse struct {
	ID               string         `json:"id"`
	CommandID        string         `json:"command_id"`
	Status           string         `json:"status"`
	ErrorMessage     string         `json:"error_message"`
	CreatedAt        string         `json:"created_at"`
	StartedAt        *string        `json:"started_at,omitempty"`
	CompletedAt      *string        `json:"completed_at,omitempty"`
	TriggeredByName  string         `json:"triggered_by_name,omitempty"`
	TriggeredByEmail string         `json:"triggered_by_email,omitempty"`
	Inputs           map[string]any `json:"inputs,omitempty"`
}

func newExecutionResponse(e store.Execution) ExecutionResponse {
	r := ExecutionResponse{
		ID:           e.ID.String(),
		CommandID:    e.CommandID.String(),
		Status:       e.Status,
		ErrorMessage: e.ErrorMessage,
		CreatedAt:    e.CreatedAt.Time.Format(timeFormat),
	}
	if e.StartedAt.Valid {
		s := e.StartedAt.Time.Format(timeFormat)
		r.StartedAt = &s
	}
	if e.CompletedAt.Valid {
		s := e.CompletedAt.Time.Format(timeFormat)
		r.CompletedAt = &s
	}
	return r
}

func newExecutionListResponse(e store.ListExecutionsForCommandRow) ExecutionResponse {
	r := ExecutionResponse{
		ID:               e.ID.String(),
		CommandID:        e.CommandID.String(),
		Status:           e.Status,
		ErrorMessage:     e.ErrorMessage,
		CreatedAt:        e.CreatedAt.Time.Format(timeFormat),
		TriggeredByName:  e.TriggeredByName,
		TriggeredByEmail: e.TriggeredByEmail,
	}
	if len(e.Inputs) > 0 {
		var inputs map[string]any
		if err := json.Unmarshal(e.Inputs, &inputs); err == nil {
			r.Inputs = inputs
		}
	}
	if e.StartedAt.Valid {
		s := e.StartedAt.Time.Format(timeFormat)
		r.StartedAt = &s
	}
	if e.CompletedAt.Valid {
		s := e.CompletedAt.Time.Format(timeFormat)
		r.CompletedAt = &s
	}
	return r
}

// ConfigResponse is the JSON shape returned for a config entry.
// The encrypted value is never included.
type ConfigResponse struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func newConfigResponse(c store.ConfigEntry) ConfigResponse {
	return ConfigResponse{
		ID:          c.ID.String(),
		Key:         c.Key,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Time.Format(timeFormat),
		UpdatedAt:   c.UpdatedAt.Time.Format(timeFormat),
	}
}

// CommandInputResponse is the JSON shape returned for a command input.
type CommandInputResponse struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Options  []string `json:"options,omitempty"`
	Multi    bool     `json:"multi"`
	Required bool     `json:"required"`
}

func newCommandInputResponse(i store.CommandInput) CommandInputResponse {
	var opts []string
	if len(i.Options) > 0 {
		_ = json.Unmarshal(i.Options, &opts)
	}
	return CommandInputResponse{
		Name:     i.Name,
		Label:    i.Label,
		Type:     i.Type,
		Options:  opts,
		Multi:    i.Multi,
		Required: i.Required,
	}
}

// CommandResponse is the JSON shape returned for a command (with inputs).
type CommandResponse struct {
	ID          string                 `json:"id"`
	Slug        string                 `json:"slug"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Inputs      []CommandInputResponse `json:"inputs"`
}

func newCommandResponse(c store.Command, inputs []store.CommandInput) CommandResponse {
	r := CommandResponse{
		ID:          c.ID.String(),
		Slug:        c.Slug,
		Name:        c.Name,
		Description: c.Description,
		Inputs:      make([]CommandInputResponse, len(inputs)),
	}
	for i, inp := range inputs {
		r.Inputs[i] = newCommandInputResponse(inp)
	}
	return r
}

// ImportErrorResponse is the JSON shape returned for a command import error.
type ImportErrorResponse struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
	FailedAt string `json:"failed_at"`
}

func newImportErrorResponse(e store.CommandImportError) ImportErrorResponse {
	return ImportErrorResponse{
		Filename: e.Filename,
		Error:    e.Error,
		FailedAt: e.FailedAt.Time.Format(timeFormat),
	}
}

// newConfigResponseFromRow converts a ListConfigEntriesRow (which omits
// value_encrypted) to a ConfigResponse. Used by handleListConfig.
func newConfigResponseFromRow(r store.ListConfigEntriesRow) ConfigResponse {
	return ConfigResponse{
		ID:          r.ID.String(),
		Key:         r.Key,
		Description: r.Description,
		CreatedAt:   r.CreatedAt.Time.Format(timeFormat),
		UpdatedAt:   r.UpdatedAt.Time.Format(timeFormat),
	}
}
