package manager

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// editorFn matches the generated clients' request-editor signature (an unnamed func type, so it is
// assignable to each gen package's RequestEditorFn).
type editorFn = func(ctx context.Context, req *http.Request) error

// Auth describes how the SDK authenticates. Construct one with [ClientCredentials], [Bearer],
// [Password], [RefreshToken] or [TokenExchange] (or their [ClientCredentialsWithScope] /
// [PasswordWithClientID] / [RefreshTokenWithSecret] variants).
type Auth interface {
	editor(baseURL string, hc *http.Client) editorFn
}

// ClientCredentials authenticates via the OAuth2 client_credentials grant against /oauth/token (the
// primary server-to-server mode). The token is fetched lazily on first use and refreshed
// transparently before it expires.
func ClientCredentials(clientID, clientSecret string) Auth {
	return &clientCredentialsAuth{clientID: clientID, clientSecret: clientSecret}
}

// ClientCredentialsWithScope is [ClientCredentials] narrowed to scope, a space-separated subset of
// the application's scopes.
func ClientCredentialsWithScope(clientID, clientSecret, scope string) Auth {
	return &clientCredentialsAuth{clientID: clientID, clientSecret: clientSecret, scope: scope}
}

type clientCredentialsAuth struct{ clientID, clientSecret, scope string }

func (a *clientCredentialsAuth) editor(baseURL string, hc *http.Client) editorFn {
	base := strings.TrimRight(baseURL, "/")
	tm := &tokenManager{grant: func(ctx context.Context) (*TokenResponse, error) {
		return ClientCredentialsGrantWithScope(ctx, hc, base, a.clientID, a.clientSecret, a.scope)
	}}
	return tm.editor()
}

// TokenExchange authenticates a partner account's application in one direct sub-account: the SDK
// gets a service token with the client_credentials grant and exchanges it (RFC 8693) for a token of
// the sub-account named by audience (its orgKey). scope may narrow the exchanged token; "" keeps
// the service token's scope. Both tokens are renewed transparently before they expire.
func TokenExchange(clientID, clientSecret, audience, scope string) Auth {
	return &tokenExchangeAuth{clientID: clientID, clientSecret: clientSecret, audience: audience, scope: scope}
}

type tokenExchangeAuth struct{ clientID, clientSecret, audience, scope string }

func (a *tokenExchangeAuth) editor(baseURL string, hc *http.Client) editorFn {
	base := strings.TrimRight(baseURL, "/")
	tm := &tokenManager{grant: func(ctx context.Context) (*TokenResponse, error) {
		// A fresh service token per exchange: the exchanged token never outlives its subject token.
		service, err := ClientCredentialsGrant(ctx, hc, base, a.clientID, a.clientSecret)
		if err != nil {
			return nil, err
		}
		return TokenExchangeGrant(ctx, hc, base, TokenExchangeRequest{
			ClientID:     a.clientID,
			ClientSecret: a.clientSecret,
			SubjectToken: service.AccessToken,
			Audience:     a.audience,
			Scope:        a.scope,
		})
	}}
	return tm.editor()
}

// Bearer authenticates with a bearer token you already hold.
func Bearer(token string) Auth { return bearerAuth{token: token} }

type bearerAuth struct{ token string }

func (a bearerAuth) editor(string, *http.Client) editorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+a.token)
		return nil
	}
}

// Password authenticates via the OAuth2 password grant against /oauth/token. The token is fetched
// lazily on first use and refreshed transparently before it expires. Convenience for interactive/dev use.
func Password(user, pass string) Auth {
	return &passwordAuth{user: user, pass: pass, clientID: "manager"}
}

// PasswordWithClientID is [Password] with an explicit OAuth2 client id instead of the default
// "manager" — for password grants issued against a custom OAuth2 application.
func PasswordWithClientID(user, pass, clientID string) Auth {
	return &passwordAuth{user: user, pass: pass, clientID: clientID}
}

type passwordAuth struct{ user, pass, clientID string }

func (a *passwordAuth) editor(baseURL string, hc *http.Client) editorFn {
	base := strings.TrimRight(baseURL, "/")
	tm := &tokenManager{grant: func(ctx context.Context) (*TokenResponse, error) {
		return PasswordGrant(ctx, hc, base, a.user, a.pass, a.clientID)
	}}
	return tm.editor()
}

// RefreshToken authenticates using a refresh token obtained from the Authorization Code + PKCE flow.
// The token is exchanged for an access token lazily on first use and refreshed transparently before
// it expires; the rotated refresh token returned by each exchange is captured and reused. clientID is
// the registered public client id used at /oauth/authorize. This is the recommended way to run a
// long-lived client on behalf of a user.
func RefreshToken(refreshToken, clientID string) Auth {
	return &refreshTokenAuth{refreshToken: refreshToken, clientID: clientID}
}

// RefreshTokenWithSecret is [RefreshToken] for confidential clients: the client secret is sent with
// every refresh-token exchange. Public (PKCE) clients have no secret and use [RefreshToken].
func RefreshTokenWithSecret(refreshToken, clientID, clientSecret string) Auth {
	return &refreshTokenAuth{refreshToken: refreshToken, clientID: clientID, clientSecret: clientSecret}
}

type refreshTokenAuth struct{ refreshToken, clientID, clientSecret string }

func (a *refreshTokenAuth) editor(baseURL string, hc *http.Client) editorFn {
	base := strings.TrimRight(baseURL, "/")
	tm := &tokenManager{}
	tm.grant = func(ctx context.Context) (*TokenResponse, error) {
		// clientSecret is empty for public-client (PKCE) refreshes; RefreshTokenGrant omits it then.
		return RefreshTokenGrant(ctx, hc, base, tm.currentRefresh(a.refreshToken), a.clientID, a.clientSecret)
	}
	return tm.editor()
}

// tokenManager lazily fetches a bearer token via its grant func and refreshes it transparently
// before expiry. It is shared by the OAuth2-based auth modes (password, client_credentials and
// refresh_token). Grants are single-flight: one caller (the leader) runs the network exchange
// while t.mu is released, and concurrent callers wait on the inflight channel — so a waiter whose
// context ends returns ctx.Err() instead of blocking behind a slow token endpoint. For the
// refresh_token mode the rotating refresh token is read via currentRefresh during the grant and
// written back under t.mu afterwards; single-flight keeps that rotation race-free.
type tokenManager struct {
	grant func(ctx context.Context) (*TokenResponse, error)

	mu       sync.Mutex
	token    string
	expiry   time.Time
	refresh  string
	inflight chan struct{} // non-nil while a grant runs; closed and cleared when it finishes
}

// currentRefresh returns the latest rotated refresh token, falling back to seed. It is called from
// the grant closure, which runs without holding t.mu — so it locks.
func (t *tokenManager) currentRefresh(seed string) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.refresh != "" {
		return t.refresh
	}
	return seed
}

func (t *tokenManager) editor() editorFn {
	return func(ctx context.Context, req *http.Request) error {
		tok, err := t.get(ctx)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		return nil
	}
}

// get returns a valid bearer token, running the grant when the cached one is missing or about to
// expire. At most one grant runs at a time; every other caller waits for it while honoring its own
// context.
func (t *tokenManager) get(ctx context.Context) (string, error) {
	t.mu.Lock()
	for {
		if t.token != "" && time.Now().Before(t.expiry.Add(-30*time.Second)) {
			tok := t.token
			t.mu.Unlock()
			return tok, nil
		}
		if t.inflight == nil {
			break // no grant in flight: become the leader
		}
		done := t.inflight
		t.mu.Unlock()
		select {
		case <-done: // the leader finished (successfully or not): re-check the state
			t.mu.Lock()
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	done := make(chan struct{})
	t.inflight = done
	t.mu.Unlock()

	tok, err := t.grant(ctx)

	t.mu.Lock()
	t.inflight = nil
	if err == nil {
		t.token = tok.AccessToken
		if tok.RefreshToken != "" {
			t.refresh = tok.RefreshToken // capture rotation (refresh tokens are single-use)
		}
		secs := tok.ExpiresIn
		if secs == 0 {
			secs = 3600
		}
		t.expiry = time.Now().Add(time.Duration(secs) * time.Second)
	}
	t.mu.Unlock()
	close(done) // wake the waiters; after a failure one of them becomes the next leader

	if err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}

// TokenResponse is the OAuth2 token endpoint response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	// Scope is the granted scope.
	Scope string `json:"scope,omitempty"`
	// AccountID is the account a service token or an exchanged token acts in.
	AccountID string `json:"account_id,omitempty"`
	// IssuedTokenType is set by the token exchange only ([AccessTokenType]).
	IssuedTokenType string `json:"issued_token_type,omitempty"`
	// ExpireTime is the moment the token expires, in milliseconds since the epoch.
	ExpireTime int64 `json:"expire_time,omitempty"`
}

// TokenExchangeGrantType is the RFC 8693 grant type of the token exchange.
const TokenExchangeGrantType = "urn:ietf:params:oauth:grant-type:token-exchange"

// AccessTokenType is the RFC 8693 token type of an access token, the only subject token type the
// API exchanges.
const AccessTokenType = "urn:ietf:params:oauth:token-type:access_token"

// TokenExchangeRequest is the input of [TokenExchangeGrant].
type TokenExchangeRequest struct {
	ClientID     string
	ClientSecret string
	// SubjectToken is a client_credentials token of the same application, in its own account.
	SubjectToken string
	// Audience names the direct sub-account by its orgKey.
	Audience string
	// Scope narrows the issued token; "" keeps the subject token's scope.
	Scope string
}

// TokenExchangeGrant exchanges a service token of a partner account's application for a token that
// acts in one direct sub-account (RFC 8693, {baseURL}/oauth/token). The issued token holds at most
// the subject token's scope, lives no longer than it and carries no refresh token. A non-2xx
// response surfaces as a typed [*APIError] (e.g. 400 invalid_target for an audience that is no
// direct sub-account).
func TokenExchangeGrant(ctx context.Context, hc *http.Client, baseURL string, req TokenExchangeRequest) (*TokenResponse, error) {
	form := url.Values{
		"grant_type":         {TokenExchangeGrantType},
		"client_id":          {req.ClientID},
		"client_secret":      {req.ClientSecret},
		"subject_token":      {req.SubjectToken},
		"subject_token_type": {AccessTokenType},
		"audience":           {req.Audience},
	}
	if req.Scope != "" {
		form.Set("scope", req.Scope)
	}
	return postTokenForm(ctx, hc, baseURL, form, "token-exchange")
}

// PasswordGrant exchanges a username/password for a token via {baseURL}/oauth/token. Exposed for
// callers who want to manage tokens themselves. A non-2xx response surfaces as a typed [*APIError].
func PasswordGrant(ctx context.Context, hc *http.Client, baseURL, user, pass, clientID string) (*TokenResponse, error) {
	if clientID == "" {
		clientID = "manager"
	}
	form := url.Values{
		"grant_type": {"password"},
		"username":   {user},
		"password":   {pass},
		"client_id":  {clientID},
	}
	return postTokenForm(ctx, hc, baseURL, form, "password")
}

// ClientCredentialsGrant exchanges a client_id/client_secret for a token via {baseURL}/oauth/token
// using the OAuth2 client_credentials grant. Exposed for callers who want to manage tokens
// themselves. A non-2xx response surfaces as a typed [*APIError].
func ClientCredentialsGrant(ctx context.Context, hc *http.Client, baseURL, clientID, clientSecret string) (*TokenResponse, error) {
	return ClientCredentialsGrantWithScope(ctx, hc, baseURL, clientID, clientSecret, "")
}

// ClientCredentialsGrantWithScope is [ClientCredentialsGrant] asking for scope, a space-separated
// subset of the application's scopes ("" asks for all of them).
func ClientCredentialsGrantWithScope(ctx context.Context, hc *http.Client, baseURL, clientID, clientSecret, scope string) (*TokenResponse, error) {
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
	if scope != "" {
		form.Set("scope", scope)
	}
	return postTokenForm(ctx, hc, baseURL, form, "client_credentials")
}

// PkceChallenge is a PKCE code verifier + S256 challenge (RFC 7636).
type PkceChallenge struct {
	CodeVerifier        string
	CodeChallenge       string
	CodeChallengeMethod string // always "S256"
}

// GeneratePKCE returns a fresh PKCE verifier + S256 challenge. Pass CodeChallenge to
// BuildAuthorizeURL and keep CodeVerifier to exchange the returned code via AuthorizationCodeGrant.
func GeneratePKCE() (PkceChallenge, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return PkceChallenge{}, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	return PkceChallenge{
		CodeVerifier:        verifier,
		CodeChallenge:       base64.RawURLEncoding.EncodeToString(sum[:]),
		CodeChallengeMethod: "S256",
	}, nil
}

// AuthorizeURLParams are the inputs to BuildAuthorizeURL.
type AuthorizeURLParams struct {
	BaseURL             string
	ClientID            string
	RedirectURI         string
	Scope               string
	CodeChallenge       string
	State               string // optional
	CodeChallengeMethod string // optional; defaults to "S256"
}

// BuildAuthorizeURL builds the GET {BaseURL}/oauth/authorize URL that starts the
// Authorization Code + PKCE flow. Redirect the user to it; babelforce redirects back to RedirectURI
// with a short-lived code.
func BuildAuthorizeURL(p AuthorizeURLParams) string {
	method := p.CodeChallengeMethod
	if method == "" {
		method = "S256"
	}
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {p.ClientID},
		"redirect_uri":          {p.RedirectURI},
		"scope":                 {p.Scope},
		"code_challenge":        {p.CodeChallenge},
		"code_challenge_method": {method},
	}
	if p.State != "" {
		q.Set("state", p.State)
	}
	return strings.TrimRight(p.BaseURL, "/") + "/oauth/authorize?" + q.Encode()
}

// AuthorizationCodeGrant exchanges an authorization code (+ PKCE verifier) for tokens via
// {baseURL}/oauth/token. Public clients pass an empty clientSecret. Exposed for callers who want to
// manage tokens themselves. A non-2xx response surfaces as a typed [*APIError].
func AuthorizationCodeGrant(ctx context.Context, hc *http.Client, baseURL, code, redirectURI, clientID, codeVerifier, clientSecret string) (*TokenResponse, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"code_verifier": {codeVerifier},
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	return postTokenForm(ctx, hc, baseURL, form, "authorization_code")
}

// RefreshTokenGrant exchanges the most-recently-issued refresh token for a fresh token set (rotated
// on every use) via {baseURL}/oauth/token. Exposed for callers who want to manage tokens themselves;
// the RefreshToken auth mode does this transparently. A non-2xx response surfaces as a typed
// [*APIError].
func RefreshTokenGrant(ctx context.Context, hc *http.Client, baseURL, refreshToken, clientID, clientSecret string) (*TokenResponse, error) {
	return RefreshTokenGrantForAccount(ctx, hc, baseURL, refreshToken, clientID, clientSecret, "")
}

// RefreshTokenGrantForAccount is [RefreshTokenGrant] that also switches the refreshed token to
// accountID, another account the user may act in (one listed by [MeResource.Accounts]); "" keeps the
// token's account. Any other account is refused with a non-2xx [*APIError].
func RefreshTokenGrantForAccount(ctx context.Context, hc *http.Client, baseURL, refreshToken, clientID, clientSecret, accountID string) (*TokenResponse, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	if clientID != "" {
		form.Set("client_id", clientID)
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	if accountID != "" {
		form.Set("account_id", accountID)
	}
	return postTokenForm(ctx, hc, baseURL, form, "refresh_token")
}

// postTokenForm POSTs a form-urlencoded grant to {baseURL}/oauth/token and decodes the response.
// A non-2xx response surfaces as a typed [*APIError] (branchable Status/Code, errors.As
// compatible) — parity with the TS/Rust SDKs. grant names the grant for the (2xx but token-less)
// decode-failure message.
func postTokenForm(ctx context.Context, hc *http.Client, baseURL string, form url.Values, grant string) (*TokenResponse, error) {
	if hc == nil {
		hc = http.DefaultClient
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/oauth/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if !isOK(resp) {
		return nil, newAPIError(resp, body)
	}
	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil || tr.AccessToken == "" {
		return nil, fmt.Errorf("manager: %s grant failed (status %d): response carries no access_token", grant, resp.StatusCode)
	}
	return &tr, nil
}
