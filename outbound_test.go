package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func leadPage(current, pages, n int) string {
	items := strings.Repeat("{},", n)
	if n > 0 {
		items = items[:len(items)-1]
	}
	return `{"items":[` + items + `],` +
		`"pagination":{"pages":` + itoa(pages) + `,"current":` + itoa(current) + `,"total":` + itoa(pages) + `,"max":1}}`
}

func outboundServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/outbound/lists/list-1" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"id":"11111111-1111-1111-1111-111111111111","name":"My List"},"success":true}`))
		case p == "/api/v2/outbound/leads" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(leadPage(2, 2, 1)))
			} else {
				_, _ = w.Write([]byte(leadPage(1, 2, 2)))
			}
		case p == "/api/v2/outbound/lists/list-1/leads/lead-1" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		case p == "/api/v2/outbound/lists/list-1/leads" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(leadPage(1, 1, 3)))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestOutboundGetList(t *testing.T) {
	srv := outboundServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	list, err := mgr.Outbound.GetList(context.Background(), "list-1")
	if err != nil {
		t.Fatal(err)
	}
	if !list.Success {
		t.Fatalf("expected success response")
	}
}

func TestOutboundLeadsAll(t *testing.T) {
	srv := outboundServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	leads, err := mgr.Outbound.LeadsAll(context.Background(), managerapi.ListOutboundLeadsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(leads) != 3 {
		t.Fatalf("expected 3 auto-paged leads, got %d", len(leads))
	}
}

func TestOutboundGetLead(t *testing.T) {
	srv := outboundServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	lead, err := mgr.Outbound.GetLead(context.Background(), "list-1", "lead-1")
	if err != nil {
		t.Fatal(err)
	}
	if !lead.Success {
		t.Fatalf("expected success response")
	}
}

func TestOutboundListLeads(t *testing.T) {
	srv := outboundServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	leads, err := mgr.Outbound.ListLeads(context.Background(), "list-1", managerapi.ListLeadsInListParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(leads.Items) != 3 {
		t.Fatalf("expected 3 leads, got %d", len(leads.Items))
	}
}
