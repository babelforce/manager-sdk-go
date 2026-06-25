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
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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

func TestApplicationsAllLocalAutomations(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v2/automations/local" && r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"items":[{},{}],"pagination":{"pages":1,"current":1,"total":2,"max":50}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	var count int
	for _, err := range mgr.Applications.AllLocalAutomations(ctx, managerapi.ListAllLocalAutomationsParams{}) {
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 2 {
		t.Fatalf("expected 2 local automations, got %d", count)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v2/automations/local" {
		t.Fatalf("AllLocalAutomations hit %s %q, want GET /api/v2/automations/local", gotMethod, gotPath)
	}
}

func TestApplicationsCloneBulkActionsErrors(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v2/applications/app1/clone" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		case r.URL.Path == "/api/v2/applications/bulk" && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"message":"updated"}`))
		case r.URL.Path == "/api/v2/applications/actions" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"a1"},{"id":"a2"}],"success":true}`))
		case r.URL.Path == "/api/v2/applications/errors" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"code":"BROKEN"}],"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
	defer srv.Close()

	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	cloned, err := mgr.Applications.Clone(ctx, "app1")
	if err != nil || !cloned.Success {
		t.Fatalf("Clone: %v (%+v)", err, cloned)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/v2/applications/app1/clone" {
		t.Fatalf("Clone hit %s %q, want POST /api/v2/applications/app1/clone", gotMethod, gotPath)
	}

	msg, err := mgr.Applications.BulkUpdate(ctx, managerapi.BulkUpdateApplicationsRequest{})
	if err != nil || msg.Message == nil || *msg.Message != "updated" {
		t.Fatalf("BulkUpdate: %v (%+v)", err, msg)
	}
	if gotMethod != http.MethodPut || gotPath != "/api/v2/applications/bulk" {
		t.Fatalf("BulkUpdate hit %s %q, want PUT /api/v2/applications/bulk", gotMethod, gotPath)
	}

	actions, err := mgr.Applications.ListActions(ctx)
	if err != nil || len(actions.Items) != 2 {
		t.Fatalf("ListActions: %v (%+v)", err, actions)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v2/applications/actions" {
		t.Fatalf("ListActions hit %s %q, want GET /api/v2/applications/actions", gotMethod, gotPath)
	}

	errs, err := mgr.Applications.ListErrors(ctx)
	if err != nil || len(errs.Items) != 1 {
		t.Fatalf("ListErrors: %v (%+v)", err, errs)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v2/applications/errors" {
		t.Fatalf("ListErrors hit %s %q, want GET /api/v2/applications/errors", gotMethod, gotPath)
	}
}

func TestApplicationViewOf(t *testing.T) {
	// Application is a oneOf union with no direct fields; ApplicationViewOf reads the shared ones.
	raw := `{"id":"11111111-1111-1111-1111-111111111111","name":"Main IVR","module":"simpleMenu",` +
		`"enabled":true,"dateCreated":"2024-01-02T03:04:05Z","lastUpdated":"2024-01-02T03:04:05Z",` +
		`"routings":[],"settings":{},"tags":["x","y"]}`
	var app managerapi.Application
	if err := app.UnmarshalJSON([]byte(raw)); err != nil {
		t.Fatal(err)
	}
	v, err := ApplicationViewOf(app)
	if err != nil {
		t.Fatal(err)
	}
	if v.Id != "11111111-1111-1111-1111-111111111111" || v.Name != "Main IVR" || v.Module != "simpleMenu" || !v.Enabled {
		t.Fatalf("unexpected view: %+v", v)
	}
	if len(v.Tags) != 2 || v.Tags[0] != "x" {
		t.Fatalf("unexpected tags: %+v", v.Tags)
	}
}

func TestApplicationsDispatch(t *testing.T) {
	var gotBody []byte
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	// nil body → no payload and no Content-Type sent (parity with the TS SDK).
	if _, err := mgr.Applications.Dispatch(ctx, "app1", "onEnter", true, nil); err != nil {
		t.Fatalf("dispatch nil: %v", err)
	}
	if len(gotBody) != 0 {
		t.Fatalf("expected empty body for nil dispatch, got %q", gotBody)
	}
	if gotCT != "" {
		t.Fatalf("expected no Content-Type for nil dispatch, got %q", gotCT)
	}

	// non-nil body → the payload is sent.
	cid := "call-123"
	if _, err := mgr.Applications.Dispatch(ctx, "app1", "onEnter", false, &managerapi.LocalAutomationDispatch{CallId: &cid}); err != nil {
		t.Fatalf("dispatch body: %v", err)
	}
	if !strings.Contains(string(gotBody), "call-123") {
		t.Fatalf("expected body to contain callId, got %q", gotBody)
	}
}

func TestApplicationErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Applications.Get(context.Background(), "app1")

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
