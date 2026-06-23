package manager

import (
	"context"
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

// Auth describes how the SDK authenticates. Construct one with [APIKey], [Bearer], or [Password].
type Auth interface {
	editor(baseURL string, hc *http.Client) editorFn
}

// APIKey authenticates with the X-Auth-Access-Id / X-Auth-Access-Token header pair (the primary
// server-to-server mode).
func APIKey(accessID, accessToken string) Auth { return apiKeyAuth{id: accessID, token: accessToken} }

type apiKeyAuth struct{ id, token string }

func (a apiKeyAuth) editor(string, *http.Client) editorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("X-Auth-Access-Id", a.id)
		req.Header.Set("X-Auth-Access-Token", a.token)
		return nil
	}
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
	tm := &tokenManager{baseURL: strings.TrimRight(baseURL, "/"), hc: hc, creds: a}
	return func(ctx context.Context, req *http.Request) error {
		tok, err := tm.get(ctx)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		return nil
	}
}

type tokenManager struct {
	baseURL string
	hc      *http.Client
	creds   *passwordAuth
	mu      sync.Mutex
	token   string
	expiry  time.Time
}

func (t *tokenManager) get(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.token != "" && time.Now().Before(t.expiry.Add(-30*time.Second)) {
		return t.token, nil
	}
	tok, err := PasswordGrant(ctx, t.hc, t.baseURL, t.creds.user, t.creds.pass, t.creds.clientID)
	if err != nil {
		return "", err
	}
	t.token = tok.AccessToken
	secs := tok.ExpiresIn
	if secs == 0 {
		secs = 3600
	}
	t.expiry = time.Now().Add(time.Duration(secs) * time.Second)
	return t.token, nil
}

// TokenResponse is the OAuth2 token endpoint response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
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
