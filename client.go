package manager

import (
	"context"
	"errors"
	"net/http"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// Options configures a [ManagerClient].
type Options struct {
	// Environment selects a named host. Ignored when BaseURL is set. Defaults to Production.
	Environment Environment
	// BaseURL overrides the host explicitly (e.g. a per-customer URL).
	BaseURL string
	// Auth is how the client authenticates. Required.
	Auth Auth
	// HTTPClient is the underlying HTTP client. Defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// ManagerClient is the babelforce manager SDK client. Create one with [Connect].
type ManagerClient struct {
	// Users is the user-management namespace (/api/v2/users).
	Users *UsersResource
}

// Connect creates and configures a client.
func Connect(_ context.Context, opts Options) (*ManagerClient, error) {
	if opts.Auth == nil {
		return nil, errors.New("manager: Auth is required")
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = resolveBaseURL(opts.Environment)
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	edit := opts.Auth.editor(baseURL, hc)
	gc, err := managerapi.NewClientWithResponses(baseURL,
		managerapi.WithHTTPClient(hc),
		managerapi.WithRequestEditorFn(edit),
	)
	if err != nil {
		return nil, err
	}

	return &ManagerClient{
		Users: &UsersResource{gc: gc},
	}, nil
}
