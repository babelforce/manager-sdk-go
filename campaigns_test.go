package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func campaignsServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Every campaigns request must carry the bearer token.
		if r.Header.Get("Authorization") != "Bearer TEST" {
			t.Errorf("missing auth headers on %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case strings.HasSuffix(p, "/status") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"dial_method":"PROGRESSIVE","order":"OLDEST_FIRST"},"success":true}`))
		case strings.HasSuffix(p, "/leads") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"}],` +
				`"pagination":{"pages":1,"current":1,"total":1,"max":1}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestCampaignStatus(t *testing.T) {
	srv := campaignsServer(t)
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	status, err := mgr.Campaigns.Status(context.Background(), uuidA)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !status.Success {
		t.Fatal("expected success status response")
	}
	if status.Item.DialMethod == nil || *status.Item.DialMethod != "PROGRESSIVE" {
		t.Fatalf("expected PROGRESSIVE dial method, got %+v", status.Item.DialMethod)
	}
}

func TestCampaignLeads(t *testing.T) {
	srv := campaignsServer(t)
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	leads, err := mgr.Campaigns.Leads(context.Background(), uuidA)
	if err != nil {
		t.Fatalf("leads: %v", err)
	}
	if len(leads.Items) != 1 {
		t.Fatalf("expected 1 lead, got %d", len(leads.Items))
	}
	if leads.Items[0].Id.String() != uuidA {
		t.Fatalf("expected lead id %s, got %s", uuidA, leads.Items[0].Id.String())
	}
}
