package manager

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	authapi "github.com/babelforce/manager-sdk-go/gen/auth"
	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// partnerCall is one request the partner test server saw.
type partnerCall struct {
	Method, Path, RawQuery string
	Header                 http.Header
	Body                   string
}

// partnerServer records every request and answers with respond.
func partnerServer(t *testing.T, respond func(c partnerCall) (int, any)) (*httptest.Server, *[]partnerCall) {
	t.Helper()
	var mu sync.Mutex
	calls := []partnerCall{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		c := partnerCall{Method: r.Method, Path: r.URL.Path, RawQuery: r.URL.RawQuery, Header: r.Header.Clone(), Body: string(b)}
		mu.Lock()
		calls = append(calls, c)
		mu.Unlock()
		status, body := respond(c)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func tokenBody(token string, extra map[string]any) map[string]any {
	b := map[string]any{"access_token": token, "token_type": "bearer", "expires_in": 3600}
	for k, v := range extra {
		b[k] = v
	}
	return b
}

var emptyUsers = map[string]any{"items": []any{}, "pagination": map[string]any{"pages": 1}}

func TestAccountsListGetAndCreate(t *testing.T) {
	status := http.StatusCreated
	items := []any{map[string]any{"uuid": "a-1", "name": "Org One", "orgKey": "org-1", "parent": "p-1"}}
	srv, calls := partnerServer(t, func(c partnerCall) (int, any) {
		if c.Method == http.MethodPost {
			return status, map[string]any{"success": true, "uuid": "a-9"}
		}
		return 200, map[string]any{"success": true, "items": items}
	})
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("T")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	all, err := mgr.Accounts.List(ctx, nil)
	if err != nil || len(all) != 1 || all[0].OrgKey != "org-1" {
		t.Fatalf("List = %v, %v", all, err)
	}
	got, err := mgr.Accounts.Get(ctx, "org-1")
	if err != nil || got == nil || got.Uuid != "a-1" {
		t.Fatalf("Get = %v, %v", got, err)
	}
	if q := (*calls)[1].RawQuery; q != "orgKey=org-1" {
		t.Fatalf("Get query = %q, want orgKey=org-1", q)
	}

	created, err := mgr.Accounts.Create(ctx, managerapi.CreateSubAccountRequest{Name: "Org Nine", OrgKey: "org-9"})
	if err != nil || created.UUID != "a-9" || !created.Created {
		t.Fatalf("Create (201) = %+v, %v", created, err)
	}
	var sent map[string]string
	_ = json.Unmarshal([]byte((*calls)[2].Body), &sent)
	if sent["name"] != "Org Nine" || sent["orgKey"] != "org-9" {
		t.Fatalf("Create body = %v", sent)
	}
	status = http.StatusOK
	created, err = mgr.Accounts.Create(ctx, managerapi.CreateSubAccountRequest{Name: "Org Nine", OrgKey: "org-9"})
	if err != nil || created.Created {
		t.Fatalf("Create (200) = %+v, %v; want Created false", created, err)
	}

	items = []any{}
	none, err := mgr.Accounts.Get(ctx, "org-2")
	if err != nil || none != nil {
		t.Fatalf("Get(unknown) = %v, %v; want nil, nil", none, err)
	}
}

func TestAccountsCreateConflictIsAPIError(t *testing.T) {
	srv, _ := partnerServer(t, func(partnerCall) (int, any) {
		return http.StatusConflict, map[string]any{"success": false, "error": "org_key_taken", "message": "taken", "orgKey": "org-9", "uuid": "a-9"}
	})
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("T")})
	_, err := mgr.Accounts.Create(context.Background(), managerapi.CreateSubAccountRequest{Name: "Other", OrgKey: "org-9"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusConflict || !strings.Contains(string(apiErr.Body), "org_key_taken") {
		t.Fatalf("err = %v; want a 409 *APIError carrying org_key_taken", err)
	}
}

func TestOrgIDOptionSendsHeaderOnAPICallsOnly(t *testing.T) {
	srv, calls := partnerServer(t, func(c partnerCall) (int, any) {
		if c.Path == "/oauth/token" {
			return 200, tokenBody("SVC", nil)
		}
		return 200, emptyUsers
	})
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: ClientCredentials("cid", "sec"), OrgID: "org-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Users.ListAll(context.Background(), ListUsersQuery{}); err != nil {
		t.Fatal(err)
	}
	for _, c := range *calls {
		got := c.Header.Get(OrgIDHeader)
		if c.Path == "/oauth/token" && got != "" {
			t.Fatalf("token grant carried %s=%q", OrgIDHeader, got)
		}
		if strings.HasPrefix(c.Path, "/api/v2/users") && (got != "org-1" || c.Header.Get("Authorization") != "Bearer SVC") {
			t.Fatalf("API call header = %q, auth %q", got, c.Header.Get("Authorization"))
		}
	}
}

func TestForSubAccountSharesTheTokenAndLeavesTheParentUnscoped(t *testing.T) {
	srv, calls := partnerServer(t, func(c partnerCall) (int, any) {
		if c.Path == "/oauth/token" {
			return 200, tokenBody("SVC", nil)
		}
		return 200, emptyUsers
	})
	ctx := context.Background()
	partner, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: ClientCredentials("cid", "sec")})
	sub, err := partner.ForSubAccount("org-2")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = sub.Users.ListAll(ctx, ListUsersQuery{})
	_, _ = partner.Users.ListAll(ctx, ListUsersQuery{})

	var api []partnerCall
	grants := 0
	for _, c := range *calls {
		if c.Path == "/oauth/token" {
			grants++
		} else {
			api = append(api, c)
		}
	}
	if grants != 1 {
		t.Fatalf("%d token grants, want 1 shared by both clients", grants)
	}
	if api[0].Header.Get(OrgIDHeader) != "org-2" || api[1].Header.Get(OrgIDHeader) != "" {
		t.Fatalf("headers = %q / %q", api[0].Header.Get(OrgIDHeader), api[1].Header.Get(OrgIDHeader))
	}
	if _, err := partner.ForSubAccount(""); err == nil {
		t.Fatal("ForSubAccount(\"\") must fail")
	}
}

func TestTokenExchangeAuthExchangesTheServiceToken(t *testing.T) {
	srv, calls := partnerServer(t, func(c partnerCall) (int, any) {
		if c.Path == "/oauth/token" {
			f, _ := url.ParseQuery(c.Body)
			if f.Get("grant_type") == "client_credentials" {
				return 200, tokenBody("SVC", nil)
			}
			return 200, tokenBody("SUB", map[string]any{"issued_token_type": AccessTokenType, "account_id": "a-1"})
		}
		return 200, emptyUsers
	})
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: TokenExchange("cid", "sec", "org-1", "api:users")})
	if _, err := mgr.Users.ListAll(context.Background(), ListUsersQuery{}); err != nil {
		t.Fatal(err)
	}
	var forms []url.Values
	for _, c := range *calls {
		if c.Path == "/oauth/token" {
			f, _ := url.ParseQuery(c.Body)
			forms = append(forms, f)
		} else if c.Header.Get("Authorization") != "Bearer SUB" {
			t.Fatalf("API call auth = %q, want the exchanged token", c.Header.Get("Authorization"))
		}
	}
	if len(forms) != 2 || forms[0].Get("grant_type") != "client_credentials" {
		t.Fatalf("grants = %v", forms)
	}
	want := url.Values{
		"grant_type":         {TokenExchangeGrantType},
		"client_id":          {"cid"},
		"client_secret":      {"sec"},
		"subject_token":      {"SVC"},
		"subject_token_type": {AccessTokenType},
		"audience":           {"org-1"},
		"scope":              {"api:users"},
	}
	if forms[1].Encode() != want.Encode() {
		t.Fatalf("exchange form = %v, want %v", forms[1], want)
	}
}

func TestTokenExchangeGrantRefusalIsAPIError(t *testing.T) {
	srv, calls := partnerServer(t, func(partnerCall) (int, any) {
		return http.StatusBadRequest, map[string]any{"error": "invalid_target", "error_description": "no"}
	})
	_, err := TokenExchangeGrant(context.Background(), srv.Client(), srv.URL, TokenExchangeRequest{
		ClientID: "cid", ClientSecret: "sec", SubjectToken: "SVC", Audience: "nope",
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusBadRequest {
		t.Fatalf("err = %v; want a 400 *APIError", err)
	}
	f, _ := url.ParseQuery((*calls)[0].Body)
	if f.Has("scope") {
		t.Fatal("scope sent although none was given")
	}
}

func TestClientCredentialsScopeAndRefreshAccountSwitch(t *testing.T) {
	srv, calls := partnerServer(t, func(partnerCall) (int, any) {
		return 200, tokenBody("A", map[string]any{"account_id": "p-1", "refresh_token": "R2"})
	})
	ctx := context.Background()
	tok, err := ClientCredentialsGrantWithScope(ctx, srv.Client(), srv.URL, "cid", "sec", "account:manage")
	if err != nil || tok.AccountID != "p-1" {
		t.Fatalf("grant = %+v, %v", tok, err)
	}
	if _, err := RefreshTokenGrantForAccount(ctx, srv.Client(), srv.URL, "R1", "", "", "acc-2"); err != nil {
		t.Fatal(err)
	}
	f0, _ := url.ParseQuery((*calls)[0].Body)
	f1, _ := url.ParseQuery((*calls)[1].Body)
	if f0.Get("scope") != "account:manage" {
		t.Fatalf("client_credentials scope = %q", f0.Get("scope"))
	}
	if f1.Get("grant_type") != "refresh_token" || f1.Get("account_id") != "acc-2" || f1.Get("refresh_token") != "R1" {
		t.Fatalf("refresh form = %v", f1)
	}
}

func TestOAuthApplicationsCoverEveryOperationWithHyphenatedIDs(t *testing.T) {
	const id = "0f8b9c1e-7a52-4d3b-9b8e-3c1f2a4d5e6f"
	app := map[string]any{"id": id, "clientId": "cid", "name": "backend"}
	srv, calls := partnerServer(t, func(c partnerCall) (int, any) {
		switch {
		case c.Method == http.MethodGet && c.Path == "/oauth/applications":
			return 200, map[string]any{"items": []any{app}}
		case c.Method == http.MethodGet && strings.HasSuffix(c.Path, "/revoke_tokens"):
			return 200, map[string]any{"item": nil}
		case c.Method == http.MethodGet:
			return 200, map[string]any{"application": app}
		case c.Method == http.MethodPost && c.Path == "/oauth/applications":
			return 201, map[string]any{"uuid": id, "client_id": "cid", "client_secret": "S1", "roles": []string{"MANAGER"}, "service_token_ttl_seconds": 900}
		case c.Method == http.MethodPost && strings.HasSuffix(c.Path, "/reset_secret"):
			return 200, map[string]any{"client_secret": "S2"}
		case c.Method == http.MethodPost:
			return 201, map[string]any{"item": nil}
		}
		return 200, map[string]any{"item": 200}
	})
	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("T")})
	apps := mgr.OAuthApplications

	list, err := apps.List(ctx)
	if err != nil || len(list) != 1 || list[0].ClientId != "cid" {
		t.Fatalf("List = %v, %v", list, err)
	}
	if got, err := apps.Get(ctx, id); err != nil || got.Name != "backend" {
		t.Fatalf("Get = %v, %v", got, err)
	}
	ttl := 900
	created, err := apps.Create(ctx, authapi.OAuthApplicationCreateRequest{Name: "backend", Label: "Backend", RedirectUri: "https://partner.example/cb", ServiceTokenTtlSeconds: &ttl})
	if err != nil || created.ClientSecret != "S1" {
		t.Fatalf("Create = %v, %v", created, err)
	}
	if err := apps.Update(ctx, id, authapi.OAuthApplicationUpdateRequest{Name: "backend", Label: "B", RedirectUri: "https://partner.example/cb"}); err != nil {
		t.Fatal(err)
	}
	if secret, err := apps.ResetSecret(ctx, id); err != nil || secret != "S2" {
		t.Fatalf("ResetSecret = %q, %v", secret, err)
	}
	if err := apps.RevokeTokens(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := apps.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	var seen []string
	for _, c := range *calls {
		seen = append(seen, c.Method+" "+c.Path)
	}
	want := []string{
		"GET /oauth/applications",
		"GET /oauth/applications/" + id,
		"POST /oauth/applications",
		"POST /oauth/applications/" + id,
		"POST /oauth/applications/" + id + "/reset_secret",
		"GET /oauth/applications/" + id + "/revoke_tokens",
		"DELETE /oauth/applications/" + id,
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("requests:\n%s\nwant:\n%s", strings.Join(seen, "\n"), strings.Join(want, "\n"))
	}
}
