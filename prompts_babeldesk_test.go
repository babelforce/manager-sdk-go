package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestPromptsAndBabeldesk(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	list := `{"items":[{"id":"` + uuidA + `"}],"pagination":{"pages":1,"current":1,"total":1,"max":50}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		isCollection := p == "/api/v2/prompts" || p == "/api/v2/babeldesk/dashboards" || p == "/api/v2/babeldesk/widgets"
		switch {
		case isCollection && m == http.MethodGet:
			_, _ = w.Write([]byte(list))
		case isCollection && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case m == http.MethodDelete:
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			_, _ = w.Write([]byte(item))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	if ps, err := mgr.Prompts.ListAll(ctx, managerapi.ListPromptsParams{}); err != nil || len(ps) != 1 {
		t.Fatalf("prompts list: %v len=%d", err, len(ps))
	}
	if _, err := mgr.Prompts.Get(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Prompts.Upload(ctx, "audio/wav", strings.NewReader("RIFF")); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Prompts.Update(ctx, "p1", managerapi.RestUpdatePrompt{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Prompts.Delete(ctx, "p1"); err != nil {
		t.Fatal(err)
	}

	if ds, err := mgr.Babeldesk.ListAll(ctx, managerapi.ListBabeldesksParams{}); err != nil || len(ds) != 1 {
		t.Fatalf("babeldesk list: %v len=%d", err, len(ds))
	}
	if _, err := mgr.Babeldesk.Create(ctx, managerapi.RestCreateBabeldesk{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Babeldesk.Get(ctx, "d1"); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Babeldesk.Delete(ctx, "d1"); err != nil {
		t.Fatal(err)
	}

	if ws, err := mgr.Babeldesk.Widgets.ListAll(ctx, managerapi.ListBabeldeskWidgetsParams{}); err != nil || len(ws) != 1 {
		t.Fatalf("widgets list: %v len=%d", err, len(ws))
	}
	if _, err := mgr.Babeldesk.Widgets.Create(ctx, managerapi.RestCreateBabeldeskWidget{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Babeldesk.Widgets.Update(ctx, "w1", managerapi.RestUpdateBabeldeskWidget{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Babeldesk.Widgets.Delete(ctx, "w1"); err != nil {
		t.Fatal(err)
	}
}

func TestPromptUsesAndWidgetSettings(t *testing.T) {
	var gotUses, gotSettings string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v2/prompts/p1/uses" && r.Method == http.MethodGet:
			gotUses = r.Method + " " + r.URL.Path
			_, _ = w.Write([]byte(`{"items":[{"id":"app1"},{"id":"app2"}],"success":true}`))
		case r.URL.Path == "/api/v2/widget/babelconnect/settings" && r.Method == http.MethodGet:
			gotSettings = r.Method + " " + r.URL.Path
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	uses, err := mgr.Prompts.Uses(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if gotUses != "GET /api/v2/prompts/p1/uses" {
		t.Fatalf("Uses hit %q, want GET /api/v2/prompts/p1/uses", gotUses)
	}
	if len(uses.Items) != 2 {
		t.Fatalf("Uses = %+v", uses)
	}

	settings, err := mgr.Babeldesk.WidgetSettings(ctx, "babelconnect")
	if err != nil {
		t.Fatal(err)
	}
	if gotSettings != "GET /api/v2/widget/babelconnect/settings" {
		t.Fatalf("WidgetSettings hit %q, want GET /api/v2/widget/babelconnect/settings", gotSettings)
	}
	if !settings.Success {
		t.Fatalf("WidgetSettings = %+v", settings)
	}
}
