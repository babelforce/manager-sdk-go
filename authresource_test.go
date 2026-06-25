package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
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
