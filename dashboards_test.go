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

func dashboardPage(id, name string, current, pages int) string {
	return `{"items":[{"id":"` + id + `","name":"` + name + `","url":"https://example.com"}],` +
		`"pagination":{"pages":` + itoa(pages) + `,"current":` + itoa(current) + `,"total":` + itoa(pages) + `,"max":1}}`
}

func dashboardServer(t *testing.T, gotAddUserBody *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Every request must carry the bearer token.
		if r.Header.Get("Authorization") != "Bearer TEST" {
			t.Errorf("missing auth header on %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/dashboards" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Ops","url":"https://example.com"},"success":true}`))
		case p == "/api/v2/dashboards" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(dashboardPage(uuidB, "Sales", 2, 2)))
			} else {
				_, _ = w.Write([]byte(dashboardPage(uuidA, "Ops", 1, 2)))
			}
		case strings.HasSuffix(p, "/users") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidB + `","email":"u@example.com"}],"success":true}`))
		case strings.HasSuffix(p, "/users") && r.Method == http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			if gotAddUserBody != nil {
				*gotAddUserBody = string(body)
			}
			_, _ = w.Write([]byte(`{"message":"added","success":true}`))
		case strings.HasPrefix(p, "/api/v2/dashboards/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Ops","url":"https://example.com"},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestDashboardsCreateListGet(t *testing.T) {
	srv := dashboardServer(t, nil)
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	created, err := mgr.Dashboards.Create(ctx, managerapi.DashboardCreateBody{Name: "Ops", Url: "https://example.com"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Item.Name != "Ops" {
		t.Fatalf("expected Ops, got %q", created.Item.Name)
	}

	dashboards, err := mgr.Dashboards.ListAll(ctx, managerapi.ListDashboardsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(dashboards) != 2 || dashboards[0].Name != "Ops" || dashboards[1].Name != "Sales" {
		t.Fatalf("expected 2 auto-paged dashboards, got %+v", dashboards)
	}

	got, err := mgr.Dashboards.Get(ctx, uuidA)
	if err != nil {
		t.Fatal(err)
	}
	if got.Item.Name != "Ops" {
		t.Fatalf("get: %q", got.Item.Name)
	}
}

func TestDashboardsUsers(t *testing.T) {
	var addBody string
	srv := dashboardServer(t, &addBody)
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	users, err := mgr.Dashboards.ListUsers(ctx, uuidA)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users.Items) != 1 || string(users.Items[0].Email) != "u@example.com" {
		t.Fatalf("expected 1 user, got %+v", users.Items)
	}

	if _, err := mgr.Dashboards.AddUser(ctx, uuidA, "u@example.com"); err != nil {
		t.Fatalf("add user: %v", err)
	}
	if !strings.Contains(addBody, `"email":"u@example.com"`) {
		t.Fatalf("expected email body, got %q", addBody)
	}
}

func TestDashboardErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Dashboards.Get(context.Background(), uuidA)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
