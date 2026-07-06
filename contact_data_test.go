package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestContactData(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, m := r.URL.Path, r.Method
		// phonebook bulk download is raw CSV, not JSON.
		if p == "/api/v2/phonebook/bulk" && m == http.MethodGet {
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write([]byte("number,name\n+49,acme\n"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case p == "/api/v2/outbound/lists" && m == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"}],"success":true}`))
		case p == "/api/v2/outbound/lists" && m == http.MethodPost:
			_, _ = w.Write([]byte(item))
		case p == "/api/v2/outbound/lists/l1/leads" && m == http.MethodDelete:
			_, _ = w.Write([]byte(item))
		case p == "/api/v2/outbound/lists/l1/leads" && m == http.MethodPost:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidB + `"},"success":true}`))
		case p == "/api/v2/outbound/lists/l1/leads/lead1" && m == http.MethodPut:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidB + `"},"success":true}`))
		case p == "/api/v2/outbound/lists/l1/leads/lead1" && m == http.MethodDelete:
			_, _ = w.Write([]byte(`{"success":true}`))
		case p == "/api/v2/phonebook" && m == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"}],"pagination":{"pages":1,"current":1,"total":1,"max":50}}`))
		case p == "/api/v2/phonebook" && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case p == "/api/v2/phonebook/bulk" && m == http.MethodPost:
			_, _ = w.Write([]byte(`{"message":"ok","success":true}`))
		case p == "/api/v2/phonebook/bulk" && m == http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"deleted","success":true}`))
		case p == "/api/v2/outbound/campaigns" && m == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"}],"success":true}`))
		case p == "/api/v2/outbound/campaigns" && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case strings.HasPrefix(p, "/api/v2/outbound/campaigns/") && m == http.MethodPut:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case m == http.MethodDelete:
			_, _ = w.Write([]byte(item))
		default: // GET / PUT single items
			_, _ = w.Write([]byte(item))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	// outbound
	if ls, err := mgr.Outbound.Lists(ctx); err != nil || len(ls) != 1 {
		t.Fatalf("outbound lists: %v len=%d", err, len(ls))
	}
	if _, err := mgr.Outbound.CreateList(ctx, managerapi.CreateOutboundListRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Outbound.ClearList(ctx, "l1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Outbound.AddLead(ctx, "l1", managerapi.AddOutboundLeadRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Outbound.UpdateLead(ctx, "l1", "lead1", managerapi.AddOutboundLeadRequest{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Outbound.DeleteLead(ctx, "l1", "lead1"); err != nil {
		t.Fatal(err)
	}

	// phonebook
	if es, err := mgr.Phonebook.ListAll(ctx, managerapi.ListPhonebookEntrysParams{}); err != nil || len(es) != 1 {
		t.Fatalf("phonebook list: %v len=%d", err, len(es))
	}
	if _, err := mgr.Phonebook.Create(ctx, managerapi.RestCreatePhonebookEntry{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Phonebook.Get(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	if csv, err := mgr.Phonebook.Download(ctx); err != nil || !strings.Contains(string(csv), "acme") {
		t.Fatalf("phonebook download: %v body=%q", err, string(csv))
	}
	if err := mgr.Phonebook.Upload(ctx, "text/csv", strings.NewReader("number,name\n+49,acme\n")); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Phonebook.Delete(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
	if msg, err := mgr.Phonebook.BulkDelete(ctx, []string{uuidA, uuidB}); err != nil || msg.Message == nil || *msg.Message != "deleted" {
		t.Fatalf("phonebook bulk delete: %v msg=%+v", err, msg)
	}

	// campaigns
	if cs, err := mgr.Campaigns.List(ctx); err != nil || len(cs) != 1 {
		t.Fatalf("campaigns list: %v len=%d", err, len(cs))
	}
	if _, err := mgr.Campaigns.Create(ctx, managerapi.CreateCampaignRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Campaigns.Get(ctx, "c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Campaigns.Update(ctx, "c1", managerapi.UpdateCampaignRequest{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Campaigns.Delete(ctx, "c1"); err != nil {
		t.Fatal(err)
	}
}
