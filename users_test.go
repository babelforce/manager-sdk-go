package manager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func userPage(uuid, email string, current, pages int) string {
	return `{"items":[{"id":"` + uuid + `","email":"` + email + `","roles":[]}],` +
		`"pagination":{"pages":` + itoa(pages) + `,"current":` + itoa(current) + `,"total":` + itoa(pages) + `,"max":1}}`
}

const (
	uuidA = "11111111-1111-1111-1111-111111111111"
	uuidB = "22222222-2222-2222-2222-222222222222"
)

func itoa(i int) string { return string(rune('0' + i)) }

func TestAPIKeyAuthAndPagination(t *testing.T) {
	var calls int
	var gotID, gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = r.Header.Get("X-Auth-Access-Id")
		gotToken = r.Header.Get("X-Auth-Access-Token")
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			_, _ = w.Write([]byte(userPage(uuidA, "a@example.com", 1, 2)))
		} else {
			_, _ = w.Write([]byte(userPage(uuidB, "b@example.com", 2, 2)))
		}
	}))
	defer srv.Close()

	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("AID", "ATOK")})
	if err != nil {
		t.Fatal(err)
	}
	users, err := mgr.Users.ListAll(context.Background(), ListUsersQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 || string(users[0].Email) != "a@example.com" || string(users[1].Email) != "b@example.com" {
		t.Fatalf("expected 2 auto-paged users, got %+v", users)
	}
	if calls != 2 {
		t.Fatalf("expected 2 page fetches, got %d", calls)
	}
	if gotID != "AID" || gotToken != "ATOK" {
		t.Fatalf("expected both X-Auth headers, got id=%q token=%q", gotID, gotToken)
	}
}

func TestTypedAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("x", "y")})
	_, err := mgr.Users.ListAll(context.Background(), ListUsersQuery{})

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 403 || apiErr.Code != "FORBIDDEN" || apiErr.Message != "nope" {
		t.Fatalf("unexpected APIError: %+v", apiErr)
	}
}

func TestPasswordGrantBearer(t *testing.T) {
	var tokenCalls int
	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth/token" {
			tokenCalls++
			_, _ = w.Write([]byte(`{"access_token":"tok123","expires_in":3600}`))
			return
		}
		authHeader = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(userPage(uuidA, "a@example.com", 1, 1)))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Password("me@example.com", "secret")})
	users, err := mgr.Users.ListAll(context.Background(), ListUsersQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if tokenCalls != 1 {
		t.Fatalf("expected 1 token fetch, got %d", tokenCalls)
	}
	if authHeader != "Bearer tok123" {
		t.Fatalf("expected bearer token, got %q", authHeader)
	}
}
