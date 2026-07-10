package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestEventsLogsExpressions(t *testing.T) {
	page2 := `{"items":[{},{}],"pagination":{"pages":1,"current":1,"total":2,"max":50}}`
	page1 := `{"items":[{}],"pagination":{"pages":1,"current":1,"total":1,"max":50}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		case p == "/api/v2/events" && m == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{}],"success":true}`))
		case p == "/api/v2/events/custom" && m == http.MethodPost:
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		case strings.HasPrefix(p, "/api/v2/events/custom/") && m == http.MethodDelete:
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v2/audit/request":
			_, _ = w.Write([]byte(page2))
		case p == "/api/v2/logs":
			_, _ = w.Write([]byte(page1))
		case p == "/api/v2/expressions":
			_, _ = w.Write([]byte(page1))
		case p == "/api/v2/expressions/evaluate":
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(p, "/dispatch/"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(p, "/variables"):
			_, _ = w.Write([]byte(`{"items":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	if evs, err := mgr.Events.List(ctx); err != nil || len(evs) != 1 {
		t.Fatalf("events list: %v len=%d", err, len(evs))
	}
	if _, err := mgr.Events.CreateCustom(ctx, managerapi.CustomEventRequest{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Events.DeleteCustom(ctx, "ce1"); err != nil {
		t.Fatal(err)
	}

	if as, err := mgr.Logs.AuditAll(ctx, managerapi.ListAuditLogsParams{}); err != nil || len(as) != 2 {
		t.Fatalf("audit: %v len=%d", err, len(as))
	}
	if ls, err := mgr.Logs.Live(ctx); err != nil || len(ls) != 1 {
		t.Fatalf("live: %v len=%d", err, len(ls))
	}

	if xs, err := mgr.Expressions.List(ctx); err != nil || len(xs) != 1 {
		t.Fatalf("expressions list: %v len=%d", err, len(xs))
	}
	if _, err := mgr.Expressions.Evaluate(ctx, managerapi.EvaluateExpression{}, false); err != nil {
		t.Fatal(err)
	}

	if _, err := mgr.Integrations.DispatchAction(ctx, "i1", "act", managerapi.IntegrationDispatchActionRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Integrations.ActionVariables(ctx, managerapi.IntegrationProvider("salesforce"), "act"); err != nil {
		t.Fatal(err)
	}
}
