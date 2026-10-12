package manager

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	authapi "github.com/babelforce/manager-sdk-go/gen/auth"
)

func TestAuthNamespace(t *testing.T) {
	var tokenMethod, revokeMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth/token":
			tokenMethod = r.Method
			_, _ = w.Write([]byte(`{"access_token":"tok-123","token_type":"Bearer","expires_in":3600}`))
		case "/oauth/revoke":
			revokeMethod = r.Method
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	clientID := "manager"
	tok, err := mgr.Auth.Token(ctx, authapi.OAuthTokenRequest{
		GrantType: authapi.OAuthTokenRequestGrantType("client_credentials"),
		ClientId:  &clientID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "tok-123" || tok.TokenType != "Bearer" {
		t.Fatalf("Token = %+v", tok)
	}
	if tokenMethod != http.MethodPost {
		t.Fatalf("token method = %q, want POST", tokenMethod)
	}

	if err := mgr.Auth.Revoke(ctx, authapi.OAuthRevokeRequest{Token: "tok-123"}); err != nil {
		t.Fatal(err)
	}
	if revokeMethod != http.MethodPost {
		t.Fatalf("revoke method = %q, want POST", revokeMethod)
	}
}

func TestClientCredentialsAuth(t *testing.T) {
	var gotAuth, gotGrant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth/token" {
			_ = r.ParseForm()
			gotGrant = r.Form.Get("grant_type")
			_, _ = w.Write([]byte(`{"access_token":"CC","expires_in":3600}`))
			return
		}
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(userPage(uuidA, "a@example.com", 1, 1)))
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, err := Connect(ctx, Options{BaseURL: srv.URL, Auth: ClientCredentials("cid", "secret")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Users.ListAll(ctx, ListUsersQuery{}); err != nil {
		t.Fatal(err)
	}
	if gotGrant != "client_credentials" {
		t.Fatalf("grant_type = %q, want client_credentials", gotGrant)
	}
	if gotAuth != "Bearer CC" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer CC")
	}
}

func TestGeneratePKCE(t *testing.T) {
	p, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	if p.CodeChallengeMethod != "S256" {
		t.Fatalf("method = %q, want S256", p.CodeChallengeMethod)
	}
	if n := len(p.CodeVerifier); n < 43 || n > 128 {
		t.Fatalf("verifier length = %d, want 43..128", n)
	}
	if strings.ContainsAny(p.CodeVerifier+p.CodeChallenge, "=+/") {
		t.Fatalf("verifier/challenge must be base64url with no padding")
	}
	sum := sha256.Sum256([]byte(p.CodeVerifier))
	if want := base64.RawURLEncoding.EncodeToString(sum[:]); p.CodeChallenge != want {
		t.Fatalf("challenge = %q, want S256 of verifier %q", p.CodeChallenge, want)
	}
	if p2, _ := GeneratePKCE(); p2.CodeVerifier == p.CodeVerifier {
		t.Fatal("expected distinct verifiers across calls")
	}
}

func TestBuildAuthorizeURL(t *testing.T) {
	u, err := url.Parse(BuildAuthorizeURL(AuthorizeURLParams{
		BaseURL:       "https://acme.babelforce.com/",
		ClientID:      "spa",
		RedirectURI:   "https://app.example.com/cb",
		Scope:         "*",
		CodeChallenge: "CHAL",
		State:         "xyz",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/oauth/authorize" {
		t.Fatalf("path = %q", u.Path)
	}
	q := u.Query()
	for k, want := range map[string]string{
		"response_type":         "code",
		"client_id":             "spa",
		"redirect_uri":          "https://app.example.com/cb",
		"scope":                 "*",
		"code_challenge":        "CHAL",
		"code_challenge_method": "S256",
		"state":                 "xyz",
	} {
		if got := q.Get(k); got != want {
			t.Fatalf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestAuthorizationCodeGrant(t *testing.T) {
	var form url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		form = r.Form
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"A","refresh_token":"R","expires_in":3600}`))
	}))
	defer srv.Close()

	tok, err := AuthorizationCodeGrant(context.Background(), nil, srv.URL,
		"the-code", "https://app.example.com/cb", "spa", "the-verifier", "")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "A" || tok.RefreshToken != "R" {
		t.Fatalf("token = %+v", tok)
	}
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "the-code" ||
		form.Get("code_verifier") != "the-verifier" || form.Get("client_id") != "spa" {
		t.Fatalf("form = %v", form)
	}
	if _, ok := form["client_secret"]; ok {
		t.Fatalf("public client must not send client_secret")
	}
}

func TestRefreshTokenAuthRotation(t *testing.T) {
	var refreshSeen []string
	var gotAuth []string
	issued := []string{"R1", "R2"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth/token" {
			_ = r.ParseForm()
			refreshSeen = append(refreshSeen, r.Form.Get("refresh_token"))
			next := issued[0]
			if len(issued) > 1 {
				issued = issued[1:]
			}
			// expires_in=1 forces a refresh before each API call (the manager refreshes 30s early)
			_, _ = w.Write([]byte(`{"access_token":"A-` + next + `","refresh_token":"` + next + `","expires_in":1}`))
			return
		}
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(userPage(uuidA, "a@example.com", 1, 1)))
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, err := Connect(ctx, Options{BaseURL: srv.URL, Auth: RefreshToken("R0", "spa")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Users.ListAll(ctx, ListUsersQuery{}); err != nil { // refresh #1: seed R0
		t.Fatal(err)
	}
	if _, err := mgr.Users.ListAll(ctx, ListUsersQuery{}); err != nil { // refresh #2: rotated R1
		t.Fatal(err)
	}
	if len(refreshSeen) < 2 {
		t.Fatalf("expected >=2 token refreshes, got %v", refreshSeen)
	}
	if refreshSeen[0] != "R0" {
		t.Fatalf("first refresh_token = %q, want seed R0", refreshSeen[0])
	}
	if refreshSeen[1] != "R1" {
		t.Fatalf("second refresh_token = %q, want rotated R1 (not the stale R0)", refreshSeen[1])
	}
}
