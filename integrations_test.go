package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestIntegrations(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		case p == "/api/v2/integrations/available":
			_, _ = w.Write([]byte(`{"items":[]}`))
		case strings.HasSuffix(p, "/actions/variables"):
			_, _ = w.Write([]byte(`{"items":[]}`))
		case strings.Contains(p, "/logo/"):
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(p, "/association/") && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"success":true}`))
		case strings.Contains(p, "/association/") && m == http.MethodDelete:
			_, _ = w.Write([]byte(`{"success":true}`))
		case p == "/api/v2/integrations" && m == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"}],"pagination":{"pages":1,"current":1,"total":1,"max":50}}`))
		case p == "/api/v2/integrations" && m == http.MethodPost:
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
	i := mgr.Integrations

	if xs, err := i.ListAll(ctx, managerapi.ListIntegrationsParams{}); err != nil || len(xs) != 1 {
		t.Fatalf("list: %v len=%d", err, len(xs))
	}
	if _, err := i.Create(ctx, managerapi.IntegrationCreateRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := i.Get(ctx, "i1"); err != nil {
		t.Fatal(err)
	}
	if _, err := i.Update(ctx, "i1", managerapi.IntegrationUpdateRequest{}); err != nil {
		t.Fatal(err)
	}
	if err := i.Delete(ctx, "i1"); err != nil {
		t.Fatal(err)
	}
	if _, err := i.Available(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := i.AddAssociation(ctx, "i1", "a1", "act"); err != nil {
		t.Fatal(err)
	}
	if _, err := i.RemoveAssociation(ctx, "i1", "a1", "act"); err != nil {
		t.Fatal(err)
	}
	if _, err := i.ProviderLogo(ctx, managerapi.IntegrationProvider("salesforce"), "64"); err != nil {
		t.Fatal(err)
	}
	if _, err := i.ProviderSessionVariables(ctx, "salesforce"); err != nil {
		t.Fatal(err)
	}
}
