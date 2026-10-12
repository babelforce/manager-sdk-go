package manager

import (
	"context"

	authapi "github.com/babelforce/manager-sdk-go/gen/auth"
)

// AuthResource is the OAuth 2.0 namespace (/oauth): authorize, token, and revoke.
//
// These endpoints authenticate via OAuth client credentials carried in the request itself, so they
// operate independently of the client's configured authentication (via [Options.Auth]).
type AuthResource struct {
	au *authapi.ClientWithResponses
}

// Token exchanges OAuth 2.0 credentials (a grant) for an access token at /oauth/token.
func (r *AuthResource) Token(ctx context.Context, body authapi.OAuthTokenRequest) (*authapi.OAuthTokenResponse, error) {
	resp, err := r.au.TokenWithFormdataBodyWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Revoke invalidates an access or refresh token at /oauth/revoke.
func (r *AuthResource) Revoke(ctx context.Context, body authapi.OAuthRevokeRequest) error {
	resp, err := r.au.RevokeWithFormdataBodyWithResponse(ctx, body)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Authorize begins the OAuth 2.0 Authorization Code flow at /oauth/authorize. On success the server
// responds with a redirect (302) carrying the authorization code; the returned response exposes the
// raw HTTP response (including Location header) for the caller to follow.
func (r *AuthResource) Authorize(ctx context.Context, params authapi.AuthorizeParams) (*authapi.AuthorizeHTTPResp, error) {
	resp, err := r.au.AuthorizeWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if !isOK(resp.HTTPResponse) && resp.HTTPResponse != nil && resp.HTTPResponse.StatusCode != 302 {
		return nil, newAPIError(resp.HTTPResponse, resp.Body)
	}
	return resp, nil
}
