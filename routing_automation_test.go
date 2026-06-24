package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestRoutingTriggersAutomations(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	list := `{"items":[{"id":"` + uuidA + `"}],"pagination":{"pages":1,"current":1,"total":1,"max":50}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		// collection list/create
		case (p == "/api/v2/routings" || p == "/api/v2/triggers" || p == "/api/v2/events/triggers") && m == http.MethodGet:
			_, _ = w.Write([]byte(list))
		case (p == "/api/v2/routings" || p == "/api/v2/triggers" || p == "/api/v2/events/triggers") && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case p == "/api/v2/triggers/test":
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(p, "/clone"):
			_, _ = w.Write([]byte(item))
		case m == http.MethodDelete:
			_, _ = w.Write([]byte(`{"success":true}`))
		default: // GET / PUT single item
			_, _ = w.Write([]byte(item))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: APIKey("x", "y")})

	// routing
	if rs, err := mgr.Routing.ListAll(ctx, managerapi.ListRoutingsParams{}); err != nil || len(rs) != 1 {
		t.Fatalf("routing list: %v len=%d", err, len(rs))
	}
	if _, err := mgr.Routing.Create(ctx, managerapi.RestCreateRouting{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Routing.Get(ctx, "r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Routing.Update(ctx, "r1", managerapi.RestUpdateRouting{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Routing.Delete(ctx, "r1"); err != nil {
		t.Fatal(err)
	}

	// triggers
	if _, err := mgr.Triggers.Create(ctx, managerapi.RestCreateTrigger{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Triggers.Clone(ctx, "t1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Triggers.Test(ctx, managerapi.TestTriggersRequest{}, true); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Triggers.Delete(ctx, "t1"); err != nil {
		t.Fatal(err)
	}

	// global automations
	if as, err := mgr.Automations.ListAll(ctx, managerapi.ListGlobalAutomationsParams{}); err != nil || len(as) != 1 {
		t.Fatalf("automations list: %v len=%d", err, len(as))
	}
	if _, err := mgr.Automations.Create(ctx, managerapi.RestCreateGlobalAutomation{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Automations.Get(ctx, "a1"); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Automations.Delete(ctx, "a1"); err != nil {
		t.Fatal(err)
	}
}
