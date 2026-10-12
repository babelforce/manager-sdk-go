package manager

import (
	"context"

	authapi "github.com/babelforce/manager-sdk-go/gen/auth"
)

// OAuthApplicationsResource is the OAuth-applications namespace of the account
// (/oauth/applications). An application's client_id and client_secret are what a server uses for
// the [ClientCredentials] auth mode.
//
// Creating, changing and deleting needs OWNER or MANAGER on the account; resetting a secret and
// ending every token of an application need OWNER. The secret is answered once, by Create and
// ResetSecret: no read returns it. Application ids are used exactly as the API returns them
// (hyphenated).
type OAuthApplicationsResource struct {
	au *authapi.ClientWithResponses
}

// List returns the applications of your account and the partner applications available to it.
func (r *OAuthApplicationsResource) List(ctx context.Context) ([]authapi.OAuthApplication, error) {
	resp, err := r.au.ListOAuthApplicationsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	out, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

// Get returns one application of your account (404 for any other).
func (r *OAuthApplicationsResource) Get(ctx context.Context, id string) (*authapi.OAuthApplication, error) {
	resp, err := r.au.GetOAuthApplicationWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	out, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return &out.Application, nil
}

// Create creates an application; the answer carries ClientSecret once.
func (r *OAuthApplicationsResource) Create(ctx context.Context, body authapi.OAuthApplicationCreateRequest) (*authapi.OAuthApplicationCreated, error) {
	resp, err := r.au.CreateOAuthApplicationWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Update changes an application's name, label, redirect URI and homepage.
func (r *OAuthApplicationsResource) Update(ctx context.Context, id string, body authapi.OAuthApplicationUpdateRequest) error {
	resp, err := r.au.UpdateOAuthApplicationWithResponse(ctx, id, body)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Delete deletes an application; every token issued to it stops working.
func (r *OAuthApplicationsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.au.DeleteOAuthApplicationWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// ResetSecret issues a new client secret and returns it; the old one stops working at once.
func (r *OAuthApplicationsResource) ResetSecret(ctx context.Context, id string) (string, error) {
	resp, err := r.au.ResetOAuthApplicationSecretWithResponse(ctx, id)
	if err != nil {
		return "", err
	}
	out, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return "", err
	}
	return out.ClientSecret, nil
}

// RevokeTokens revokes every token of the application in your account, exchanged tokens included.
func (r *OAuthApplicationsResource) RevokeTokens(ctx context.Context, id string) error {
	resp, err := r.au.RevokeOAuthApplicationTokensWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
