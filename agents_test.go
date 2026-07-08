package manager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

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
		case strings.HasPrefix(p, "/api/v2/agents/bulk/") && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"bulk":[{"id":"` + uuidA + `","success":true}]}`))
		case p == "/api/v2/agents/provision/validate" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"success":true,"total":3,"message":"valid"}`))
		case strings.HasPrefix(p, "/api/v2/agents/provision/jobs/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"id":"job-1","status":"completed","success":true},"success":true}`))
		case p == "/api/v2/agents/provision" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write([]byte("id,name\n" + uuidA + ",Alice\n"))
		case p == "/api/v2/agents/provision" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"success":true,"total":2,"message":"imported"}`))
		case p == "/api/v2/agents/logs" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"type":"login","presence":"available"}],` +
				`"pagination":{"pages":1,"current":1,"total":1,"max":1},"success":true}`))
		case strings.HasSuffix(p, "/logs") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"type":"call","presence":"busy"}],` +
				`"pagination":{"pages":1,"current":1,"total":1,"max":1},"success":true}`))
		case p == "/api/v2/agents/push" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"message":"pushed","success":true}`))
		case strings.HasSuffix(p, "/password") && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"message":"password updated","success":true}`))
		case p == "/api/v2/agents/groups/bulk" && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"groups deleted","success":true}`))
		case strings.HasPrefix(p, "/api/v2/agents/groups/") && strings.HasSuffix(p, "/agents") && r.Method == http.MethodGet:
			// list agents in a group
			_, _ = w.Write([]byte(agentPage(uuidB, "Bob", 1, 1)))
		case strings.Contains(p, "/agents/") && strings.HasPrefix(p, "/api/v2/agents/groups/") && r.Method == http.MethodDelete:
			// remove an agent from a group
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Support"},"success":true}`))
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
		case p == "/api/v2/agents/presence/available" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"item":{"name":"lunch","label":"Lunch","available":false},"success":true}`))
		case p == "/api/v2/agents/presence/available" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"name":"lunch","label":"Lunch","available":false}],"success":true}`))
		case strings.HasPrefix(p, "/api/v2/agents/presence/available/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"name":"lunch","label":"Lunch","available":false},"success":true}`))
		case strings.HasPrefix(p, "/api/v2/agents/presence/available/") && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"item":{"name":"lunch","label":"Out to lunch","available":false},"success":true}`))
		case strings.HasPrefix(p, "/api/v2/agents/presence/available/") && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"deleted","success":true}`))
		case p == "/api/v2/agents/status/available" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"name":"available","label":"Available","type":"presence"}],"success":true}`))
		case strings.HasSuffix(p, "/status") && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"enabled":true,"available":true,"display_status":"online",` +
				`"line_status":"available","outbound_status":"idle","presence":{}}`))
		case strings.HasSuffix(p, "/status") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":[{"enabled":true,"available":true,"display_status":"online",` +
				`"line_status":"available"}],"success":true}`))
		case strings.HasSuffix(p, "/enable") && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"message":"enabled","success":true}`))
		case strings.HasSuffix(p, "/disable") && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"message":"disabled","success":true}`))
		case strings.HasSuffix(p, "/hangup") && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"message":"hungup","success":true}`))
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
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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

func TestAgentPresencesCRUD(t *testing.T) {
	srv := agentServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	list, err := mgr.Agents.Presences(ctx)
	if err != nil {
		t.Fatalf("presences: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].Name == nil || *list.Items[0].Name != "lunch" {
		t.Fatalf("expected 1 presence lunch, got %+v", list.Items)
	}

	created, err := mgr.Agents.CreatePresence(ctx, managerapi.AgentPresenceWriteBody{Label: "Lunch"})
	if err != nil {
		t.Fatalf("create presence: %v", err)
	}
	if created.Item.Name != "lunch" {
		t.Fatalf("expected lunch, got %q", created.Item.Name)
	}

	got, err := mgr.Agents.GetPresence(ctx, "lunch")
	if err != nil {
		t.Fatalf("get presence: %v", err)
	}
	if got.Item.Label != "Lunch" {
		t.Fatalf("expected Lunch, got %q", got.Item.Label)
	}

	updated, err := mgr.Agents.UpdatePresence(ctx, "lunch", managerapi.AgentPresenceWriteBody{Label: "Out to lunch"})
	if err != nil {
		t.Fatalf("update presence: %v", err)
	}
	if updated.Item.Label != "Out to lunch" {
		t.Fatalf("expected updated label, got %q", updated.Item.Label)
	}

	deleted, err := mgr.Agents.DeletePresence(ctx, "lunch")
	if err != nil {
		t.Fatalf("delete presence: %v", err)
	}
	if !deleted.Success {
		t.Fatalf("expected delete success, got %+v", deleted)
	}
}

func TestAgentStatusAndLifecycle(t *testing.T) {
	srv := agentServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	status, err := mgr.Agents.GetStatus(ctx, uuidA)
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	if len(status.Item) != 1 || status.Item[0].Enabled == nil || !*status.Item[0].Enabled {
		t.Fatalf("expected enabled total status, got %+v", status.Item)
	}

	avail, err := mgr.Agents.AvailableStatuses(ctx)
	if err != nil {
		t.Fatalf("available statuses: %v", err)
	}
	if len(avail.Items) != 1 || avail.Items[0].Name == nil || *avail.Items[0].Name != "available" {
		t.Fatalf("expected 1 availability, got %+v", avail.Items)
	}

	enabled, err := mgr.Agents.Enable(ctx, uuidA)
	if err != nil {
		t.Fatalf("enable: %v", err)
	}
	if enabled.Message == nil || *enabled.Message != "enabled" {
		t.Fatalf("expected enabled message, got %+v", enabled)
	}

	disabled, err := mgr.Agents.Disable(ctx, uuidA)
	if err != nil {
		t.Fatalf("disable: %v", err)
	}
	if disabled.Message == nil || *disabled.Message != "disabled" {
		t.Fatalf("expected disabled message, got %+v", disabled)
	}

	hung, err := mgr.Agents.HangupCall(ctx, uuidA)
	if err != nil {
		t.Fatalf("hangup: %v", err)
	}
	if hung.Message == nil || *hung.Message != "hungup" {
		t.Fatalf("expected hungup message, got %+v", hung)
	}
}

func TestAgentBulkActionAndPush(t *testing.T) {
	srv := agentServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	id, _ := uuid.Parse(uuidA)
	bulk, err := mgr.Agents.BulkAction(ctx, "enable", managerapi.AgentBulkRequest{Ids: []managerapi.ObjectUuid{id}})
	if err != nil {
		t.Fatalf("bulk action: %v", err)
	}
	if len(bulk.Bulk) != 1 || bulk.Bulk[0].Success == nil || !*bulk.Bulk[0].Success {
		t.Fatalf("expected 1 successful bulk result, got %+v", bulk.Bulk)
	}

	pushed, err := mgr.Agents.Push(ctx, managerapi.AgentPushRequest{})
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	if pushed.Message == nil || *pushed.Message != "pushed" {
		t.Fatalf("expected pushed message, got %+v", pushed)
	}

	updated, err := mgr.Agents.UpdatePassword(ctx, uuidA, managerapi.AgentPasswordUpdateRequest{})
	if err != nil {
		t.Fatalf("update password: %v", err)
	}
	if updated.Message == nil || *updated.Message != "password updated" {
		t.Fatalf("expected password updated message, got %+v", updated)
	}
}

func TestAgentProvisioning(t *testing.T) {
	srv := agentServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	csv, err := mgr.Agents.Export(ctx, "csv")
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(string(csv), "Alice") {
		t.Fatalf("expected export to contain Alice, got %q", string(csv))
	}

	imported, err := mgr.Agents.Import(ctx, "text/csv", strings.NewReader("id,name\n"),
		&managerapi.ImportAgentsParams{Format: managerapi.ImportAgentsParamsFormat("csv")})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if !imported.Success || imported.Total == nil || *imported.Total != 2 {
		t.Fatalf("expected import total 2, got %+v", imported)
	}

	validated, err := mgr.Agents.ValidateImport(ctx, "text/csv", strings.NewReader("id,name\n"))
	if err != nil {
		t.Fatalf("validate import: %v", err)
	}
	if !validated.Success || validated.Total == nil || *validated.Total != 3 {
		t.Fatalf("expected validate total 3, got %+v", validated)
	}

	job, err := mgr.Agents.GetImportJob(ctx, "job-1")
	if err != nil {
		t.Fatalf("get import job: %v", err)
	}
	if job.Item.Id != "job-1" {
		t.Fatalf("expected job-1, got %q", job.Item.Id)
	}
}

func TestAgentLogs(t *testing.T) {
	srv := agentServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	var agentLogs []managerapi.AgentLogEntry
	for e, err := range mgr.Agents.Logs(ctx, uuidA, managerapi.ListAgentLogsParams{}) {
		if err != nil {
			t.Fatalf("logs: %v", err)
		}
		agentLogs = append(agentLogs, e)
	}
	if len(agentLogs) != 1 || agentLogs[0].Type == nil || *agentLogs[0].Type != "call" {
		t.Fatalf("expected 1 call log, got %+v", agentLogs)
	}

	var allLogs []managerapi.AgentLogEntry
	for e, err := range mgr.Agents.AllLogs(ctx, managerapi.ListAllAgentLogsParams{}) {
		if err != nil {
			t.Fatalf("all logs: %v", err)
		}
		allLogs = append(allLogs, e)
	}
	if len(allLogs) != 1 || allLogs[0].Type == nil || *allLogs[0].Type != "login" {
		t.Fatalf("expected 1 login log, got %+v", allLogs)
	}
}

func TestAgentGroupAgentsAndBulkDelete(t *testing.T) {
	srv := agentServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	var groupAgents []managerapi.Agent
	for a, err := range mgr.Agents.Groups.ListAgents(ctx, uuidA, managerapi.ListAgentsInGroupParams{}) {
		if err != nil {
			t.Fatalf("list agents in group: %v", err)
		}
		groupAgents = append(groupAgents, a)
	}
	if len(groupAgents) != 1 || groupAgents[0].Name != "Bob" {
		t.Fatalf("expected 1 group agent Bob, got %+v", groupAgents)
	}

	removed, err := mgr.Agents.Groups.RemoveAgent(ctx, uuidA, uuidB)
	if err != nil {
		t.Fatalf("remove agent: %v", err)
	}
	if removed.Item.Name != "Support" {
		t.Fatalf("expected Support group, got %q", removed.Item.Name)
	}

	deleted, err := mgr.Agents.Groups.BulkDelete(ctx, []string{uuidA, uuidB})
	if err != nil {
		t.Fatalf("bulk delete: %v", err)
	}
	if deleted.Message == nil || *deleted.Message != "groups deleted" {
		t.Fatalf("expected groups deleted message, got %+v", deleted)
	}
}

func TestAgentErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Agents.Get(context.Background(), uuidA)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
