package manager

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
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

func TestBearerAuthAndPagination(t *testing.T) {
	var calls int
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			_, _ = w.Write([]byte(userPage(uuidA, "a@example.com", 1, 2)))
		} else {
			_, _ = w.Write([]byte(userPage(uuidB, "b@example.com", 2, 2)))
		}
	}))
	defer srv.Close()

	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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
	if gotAuth != "Bearer TEST" {
		t.Fatalf("expected Authorization header %q, got %q", "Bearer TEST", gotAuth)
	}
}

func TestTypedAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Users.ListAll(context.Background(), ListUsersQuery{})

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 403 || apiErr.Code != "FORBIDDEN" || apiErr.Message != "nope" {
		t.Fatalf("unexpected APIError: %+v", apiErr)
	}
}

func TestUsersRolesAndReset(t *testing.T) {
	type rec struct{ method, path, body string }
	var last rec
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		last = rec{r.Method, r.URL.Path, string(b)}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v2/users/roles" && r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"items":["manager","agent"]}`))
			return
		}
		_, _ = w.Write([]byte(`{"message":"ok","success":true}`))
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	roles, err := mgr.Users.ListRoles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 2 || roles[0] != "manager" || roles[1] != "agent" {
		t.Fatalf("roles = %v", roles)
	}

	if err := mgr.Users.AddRoles(ctx, []string{"a@example.com"}, []managerapi.AccountRole{"agent"}); err != nil {
		t.Fatal(err)
	}
	if last.method != http.MethodPost || last.path != "/api/v2/users/roles" {
		t.Fatalf("addRoles request = %+v", last)
	}
	if !strings.Contains(last.body, "a@example.com") || !strings.Contains(last.body, `"agent"`) {
		t.Fatalf("addRoles body = %s", last.body)
	}

	if err := mgr.Users.RemoveRoles(ctx, []string{"a@example.com"}, []managerapi.AccountRole{"agent"}); err != nil {
		t.Fatal(err)
	}
	if last.path != "/api/v2/users/roles/remove" {
		t.Fatalf("removeRoles path = %s", last.path)
	}

	if err := mgr.Users.ResetPasswords(ctx, []string{"a@example.com"}); err != nil {
		t.Fatal(err)
	}
	if last.method != http.MethodPost || last.path != "/api/v2/users/reset-password" {
		t.Fatalf("resetPasswords request = %+v", last)
	}
}

func TestUsersMeAndGetByEmail(t *testing.T) {
	var gotMe, gotByEmail string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v2/me" && r.Method == http.MethodGet:
			gotMe = r.Method + " " + r.URL.Path
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","email":"me@example.com"},"success":true}`))
		case r.URL.Path == "/api/v2/users/by-email/a@example.com" && r.Method == http.MethodGet:
			gotByEmail = r.Method + " " + r.URL.Path
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidB + `","email":"a@example.com","roles":[]},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	me, err := mgr.Users.Me(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if gotMe != "GET /api/v2/me" {
		t.Fatalf("Me hit %q, want GET /api/v2/me", gotMe)
	}
	if me.Item["email"] != "me@example.com" || me.Success == nil || !*me.Success {
		t.Fatalf("Me = %+v", me)
	}

	usr, err := mgr.Users.GetByEmail(ctx, "a@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if gotByEmail != "GET /api/v2/users/by-email/a@example.com" {
		t.Fatalf("GetByEmail hit %q, want GET /api/v2/users/by-email/a@example.com", gotByEmail)
	}
	if string(usr.Item.Email) != "a@example.com" {
		t.Fatalf("GetByEmail = %+v", usr.Item)
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
