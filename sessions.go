package manager

import (
	"context"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// SessionsResource is the call/automation sessions namespace (/api/v2/sessions).
type SessionsResource struct {
	gc *managerapi.ClientWithResponses
}

// Create creates a new session.
func (r *SessionsResource) Create(ctx context.Context) (*managerapi.SessionResponse, error) {
	resp, err := r.gc.CreateSessionWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Get returns a session (its variables) by id.
func (r *SessionsResource) Get(ctx context.Context, id string) (*managerapi.SessionResponse, error) {
	resp, err := r.gc.GetSessionVariablesWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdateVariables updates a session's variables.
func (r *SessionsResource) UpdateVariables(ctx context.Context, id string, variables managerapi.UpdateSessionVariablesRequest) (*managerapi.SessionResponse, error) {
	resp, err := r.gc.UpdateSessionVariablesWithResponse(ctx, id, variables)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
