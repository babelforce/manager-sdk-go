package manager

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestIntegrations(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	var bulkDeleteBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		case p == "/api/v2/integrations/providers":
			_, _ = w.Write([]byte(`{"items":[]}`))
		case p == "/api/v2/integrations/bulk" && m == http.MethodDelete:
			b, _ := io.ReadAll(r.Body)
			bulkDeleteBody = string(b)
			_, _ = w.Write([]byte(`{"success":true}`))
		case strings.HasSuffix(p, "/tokens"):
			_, _ = w.Write([]byte(`{"items":[]}`))
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
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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
	if _, err := i.Providers(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := i.ListTokens(ctx, "i1"); err != nil {
		t.Fatal(err)
	}
	if _, err := i.BulkDelete(ctx, []string{uuidA, uuidB}); err != nil {
		t.Fatal(err)
	}
	var got struct {
		Ids []string `json:"ids"`
	}
	if err := json.Unmarshal([]byte(bulkDeleteBody), &got); err != nil {
		t.Fatalf("bulk delete body %q: %v", bulkDeleteBody, err)
	}
	if len(got.Ids) != 2 || got.Ids[0] != uuidA || got.Ids[1] != uuidB {
		t.Fatalf("bulk delete sent unexpected ids: %q", bulkDeleteBody)
	}
}

func TestIntegrationsActions(t *testing.T) {
	var (
		gotMethod   string
		gotPath     string
		gotRawQuery string
		gotBody     string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		gotMethod, gotPath, gotRawQuery = r.Method, r.URL.Path, r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		switch {
		case r.URL.Path == "/api/v2/actions" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"name":"send_email"}],"success":true}`))
		case strings.HasSuffix(r.URL.Path, "/params"):
			_, _ = w.Write([]byte(`{"items":[{"name":"to"}],"success":true}`))
		case strings.HasPrefix(r.URL.Path, "/api/v2/actions/") && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"message":"done","success":true}`))
		case strings.Contains(r.URL.Path, "/dispatch/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"action":"lookup","context":{},"data":{},"params":{"id":"42"}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	i := mgr.Integrations

	// ListActions with type filter.
	typ := "salesforce"
	actions, err := i.ListActions(ctx, managerapi.ListActionsParams{Type: &typ})
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v2/actions" {
		t.Fatalf("ListActions: %s %s", gotMethod, gotPath)
	}
	if gotRawQuery != "type=salesforce" {
		t.Fatalf("ListActions query: %q", gotRawQuery)
	}
	if len(actions.Items) != 1 || actions.Items[0]["name"] != "send_email" {
		t.Fatalf("ListActions decoded: %+v", actions.Items)
	}

	// ListActionParams.
	params, err := i.ListActionParams(ctx, "salesforce", "send_email")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v2/actions/salesforce/send_email/params" {
		t.Fatalf("ListActionParams: %s %s", gotMethod, gotPath)
	}
	if len(params.Items) != 1 || params.Items[0]["name"] != "to" {
		t.Fatalf("ListActionParams decoded: %+v", params.Items)
	}

	// ExecuteAction.
	exec, err := i.ExecuteAction(ctx, "email", "send_email", map[string]any{"to": "x@y.z"})
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/v2/actions/email/send_email" {
		t.Fatalf("ExecuteAction: %s %s", gotMethod, gotPath)
	}
	if !strings.Contains(gotBody, `"to":"x@y.z"`) {
		t.Fatalf("ExecuteAction body: %q", gotBody)
	}
	if !exec.Success || exec.Message == nil || *exec.Message != "done" {
		t.Fatalf("ExecuteAction decoded: %+v", exec)
	}

	// DispatchActionGet with call and session scope.
	callID, sessionID := "c1", "s1"
	disp, err := i.DispatchActionGet(ctx, "i1", "lookup", managerapi.DispatchActionGetParams{
		CallId:    &callID,
		SessionId: &sessionID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v2/integrations/i1/dispatch/lookup" {
		t.Fatalf("DispatchActionGet: %s %s", gotMethod, gotPath)
	}
	if !strings.Contains(gotRawQuery, "callId=c1") || !strings.Contains(gotRawQuery, "sessionId=s1") {
		t.Fatalf("DispatchActionGet query: %q", gotRawQuery)
	}
	if disp.Action != "lookup" || disp.Params["id"] != "42" {
		t.Fatalf("DispatchActionGet decoded: %+v", disp)
	}
}
