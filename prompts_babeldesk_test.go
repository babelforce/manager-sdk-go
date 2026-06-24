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
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: APIKey("x", "y")})

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
