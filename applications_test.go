package manager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func appServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/applications" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		case p == "/api/v2/applications" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(`{"items":[{}],"pagination":{"pages":2,"current":2,"total":3,"max":2}}`))
			} else {
				_, _ = w.Write([]byte(`{"items":[{},{}],"pagination":{"pages":2,"current":1,"total":3,"max":2}}`))
			}
		case p == "/api/v2/applications/modules" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":["mod1","mod2"]}`))
		case p == "/api/v2/applications/bulk" && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"deleted"}`))
		case strings.HasSuffix(p, "/actions") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		case strings.HasSuffix(p, "/actions") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{}],"pagination":{"pages":1,"current":1,"total":1,"max":1}}`))
		case strings.HasPrefix(p, "/api/v2/applications/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestApplicationsCRUDAndModules(t *testing.T) {
	srv := appServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	created, err := mgr.Applications.Create(ctx, managerapi.ApplicationCreateBody{})
	if err != nil || !created.Success {
		t.Fatalf("create: %v (success=%v)", err, created)
	}

	apps, err := mgr.Applications.ListAll(ctx, ListApplicationsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 3 {
		t.Fatalf("expected 3 auto-paged applications, got %d", len(apps))
	}

	got, err := mgr.Applications.Get(ctx, "app1")
	if err != nil || !got.Success {
		t.Fatalf("get: %v", err)
	}

	mods, err := mgr.Applications.ListModules(ctx)
	if err != nil || mods == nil {
		t.Fatalf("listModules: %v", err)
	}

	if _, err := mgr.Applications.DeleteMany(ctx, []string{"app1", "app2"}); err != nil {
		t.Fatalf("deleteMany: %v", err)
	}
}

func TestApplicationsListOptionalPagination(t *testing.T) {
	// A response with no `pagination` block must yield its items then stop (not loop forever).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{},{}]}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	apps, err := mgr.Applications.ListAll(context.Background(), ListApplicationsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 applications (single page, no pagination), got %d", len(apps))
	}
}

func TestAppActionsCreateList(t *testing.T) {
	srv := appServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	ctx := context.Background()

	created, err := mgr.Applications.Actions.Create(ctx, "app1", managerapi.RestCreateLocalAutomation{})
	if err != nil || !created.Success {
		t.Fatalf("action create: %v", err)
	}

	actions, err := mgr.Applications.Actions.ListAll(ctx, "app1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
}

func TestApplicationErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	_, err := mgr.Applications.Get(context.Background(), "app1")

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
