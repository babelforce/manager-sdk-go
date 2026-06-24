package manager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func agentPage(id, name string, current, pages int) string {
	return `{"items":[{"id":"` + id + `","name":"` + name + `"}],` +
		`"pagination":{"pages":` + itoa(pages) + `,"current":` + itoa(current) + `,"total":` + itoa(pages) + `,"max":1}}`
}

func agentServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/agents" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Alice"},"success":true}`))
		case p == "/api/v2/agents" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(agentPage(uuidB, "Bob", 2, 2)))
			} else {
				_, _ = w.Write([]byte(agentPage(uuidA, "Alice", 1, 2)))
			}
		case p == "/api/v2/agents/groups" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Support"},"success":true}`))
		case p == "/api/v2/agents/groups" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `","name":"Support"}],` +
				`"pagination":{"pages":1,"current":1,"total":1,"max":1}}`))
		case strings.HasSuffix(p, "/agents") && r.Method == http.MethodPost:
			// add an agent to a group
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Support"},` +
				`"agent":{"id":"` + uuidB + `","name":"Bob"},"success":true}`))
		case strings.HasSuffix(p, "/status") && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"enabled":true,"available":true,"display_status":"online",` +
				`"line_status":"available","outbound_status":"idle","presence":{}}`))
		case strings.HasPrefix(p, "/api/v2/agents/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Alice"},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestAgentsCreateListGetStatus(t *testing.T) {
	srv := agentServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	created, err := mgr.Agents.Create(ctx, managerapi.RestCreateAgent{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Item.Name != "Alice" {
		t.Fatalf("expected Alice, got %q", created.Item.Name)
	}

	agents, err := mgr.Agents.ListAll(ctx, ListAgentsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 2 || agents[0].Name != "Alice" || agents[1].Name != "Bob" {
		t.Fatalf("expected 2 auto-paged agents, got %+v", agents)
	}

	got, err := mgr.Agents.Get(ctx, uuidA)
	if err != nil {
		t.Fatal(err)
	}
	if got.Item.Name != "Alice" {
		t.Fatalf("get: %q", got.Item.Name)
	}

	enabled := true
	status, err := mgr.Agents.UpdateStatus(ctx, uuidA, managerapi.UpdateAgentStatusRequest{Enabled: &enabled})
	if err != nil {
		t.Fatal(err)
	}
	if !status.Enabled {
		t.Fatal("expected enabled status")
	}
}

func TestAgentGroupsCreateListAddAgent(t *testing.T) {
	srv := agentServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	ctx := context.Background()

	group, err := mgr.Agents.Groups.Create(ctx, managerapi.RestCreateAgentGroup{})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if group.Item.Name != "Support" {
		t.Fatalf("expected Support, got %q", group.Item.Name)
	}

	groups, err := mgr.Agents.Groups.ListAll(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups[0].Name != "Support" {
		t.Fatalf("expected 1 group, got %+v", groups)
	}

	added, err := mgr.Agents.Groups.AddAgent(ctx, uuidA, uuidB)
	if err != nil {
		t.Fatalf("add agent: %v", err)
	}
	if added.Agent == nil || added.Agent.Name != "Bob" {
		t.Fatalf("expected added agent Bob, got %+v", added.Agent)
	}
}

func TestAgentErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	_, err := mgr.Agents.Get(context.Background(), uuidA)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
