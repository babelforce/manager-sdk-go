package manager

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// editorFn matches the generated clients' request-editor signature (an unnamed func type, so it is
// assignable to each gen package's RequestEditorFn).
type editorFn = func(ctx context.Context, req *http.Request) error

// Auth describes how the SDK authenticates. Construct one with [ClientCredentials], [Bearer], or [Password].
type Auth interface {
	editor(baseURL string, hc *http.Client) editorFn
}

// ClientCredentials authenticates via the OAuth2 client_credentials grant against /oauth/token (the
// primary server-to-server mode). The token is fetched lazily on first use and refreshed
// transparently before it expires.
func ClientCredentials(clientID, clientSecret string) Auth {
	return &clientCredentialsAuth{clientID: clientID, clientSecret: clientSecret}
}

type clientCredentialsAuth struct{ clientID, clientSecret string }

func (a *clientCredentialsAuth) editor(baseURL string, hc *http.Client) editorFn {
	base := strings.TrimRight(baseURL, "/")
	tm := &tokenManager{grant: func(ctx context.Context) (*TokenResponse, error) {
		return ClientCredentialsGrant(ctx, hc, base, a.clientID, a.clientSecret)
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

type refreshTokenAuth struct{ refreshToken, clientID string }

func (a *refreshTokenAuth) editor(baseURL string, hc *http.Client) editorFn {
	base := strings.TrimRight(baseURL, "/")
	tm := &tokenManager{}
	tm.grant = func(ctx context.Context) (*TokenResponse, error) {
		// Public-client refresh: no client secret. Use RefreshTokenGrant directly for confidential clients.
		return RefreshTokenGrant(ctx, hc, base, tm.currentRefresh(a.refreshToken), a.clientID, "")
	}
	return tm.editor()
}

// tokenManager lazily fetches a bearer token via its grant func and refreshes it transparently
// before expiry. It is shared by the OAuth2-based auth modes (password, client_credentials and
// refresh_token). For the refresh_token mode it also tracks the rotating refresh token: get() holds
// t.mu across the whole grant call, so the grant closure may read currentRefresh and get() may write
// the rotated token back without any extra locking.
type tokenManager struct {
	grant   func(ctx context.Context) (*TokenResponse, error)
	mu      sync.Mutex
	token   string
	expiry  time.Time
	refresh string
}

// currentRefresh returns the latest rotated refresh token, falling back to seed. It is called from
// the grant closure, which runs while get() holds t.mu — so it does not lock.
func (t *tokenManager) currentRefresh(seed string) string {
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

func (t *tokenManager) get(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.token != "" && time.Now().Before(t.expiry.Add(-30*time.Second)) {
		return t.token, nil
	}
	tok, err := t.grant(ctx)
	if err != nil {
		return "", err
	}
	t.token = tok.AccessToken
	if tok.RefreshToken != "" {
		t.refresh = tok.RefreshToken // capture rotation (refresh tokens are single-use)
	}
	secs := tok.ExpiresIn
	if secs == 0 {
		secs = 3600
	}
	t.expiry = time.Now().Add(time.Duration(secs) * time.Second)
	return t.token, nil
}

// TokenResponse is the OAuth2 token endpoint response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

// PasswordGrant exchanges a username/password for a token via {baseURL}/oauth/token. Exposed for
// callers who want to manage tokens themselves.
func PasswordGrant(ctx context.Context, hc *http.Client, baseURL, user, pass, clientID string) (*TokenResponse, error) {
	if hc == nil {
		hc = http.DefaultClient
	}
	if clientID == "" {
		clientID = "manager"
	}
	form := url.Values{
		"grant_type": {"password"},
		"username":   {user},
		"password":   {pass},
		"client_id":  {clientID},
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
	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil || tr.AccessToken == "" {
		return nil, fmt.Errorf("manager: password grant failed (status %d)", resp.StatusCode)
	}
	return &tr, nil
}

// ClientCredentialsGrant exchanges a client_id/client_secret for a token via {baseURL}/oauth/token
// using the OAuth2 client_credentials grant. Exposed for callers who want to manage tokens themselves.
func ClientCredentialsGrant(ctx context.Context, hc *http.Client, baseURL, clientID, clientSecret string) (*TokenResponse, error) {
	if hc == nil {
		hc = http.DefaultClient
	}
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
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
	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil || tr.AccessToken == "" {
		return nil, fmt.Errorf("manager: client credentials grant failed (status %d)", resp.StatusCode)
	}
	return &tr, nil
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
// manage tokens themselves.
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
// the RefreshToken auth mode does this transparently.
func RefreshTokenGrant(ctx context.Context, hc *http.Client, baseURL, refreshToken, clientID, clientSecret string) (*TokenResponse, error) {
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
	return postTokenForm(ctx, hc, baseURL, form, "refresh_token")
}

// postTokenForm POSTs a form-urlencoded grant to {baseURL}/oauth/token and decodes the response.
// grant names the grant for error messages.
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
	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil || tr.AccessToken == "" {
		return nil, fmt.Errorf("manager: %s grant failed (status %d)", grant, resp.StatusCode)
	}
	return &tr, nil
}
