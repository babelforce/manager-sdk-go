package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	taskautomationapi "github.com/babelforce/manager-sdk-go/gen/taskautomation"
)

func TestTaskRuntime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v3/tasks/usage/types" {
			_, _ = w.Write([]byte(`[]`)) // TaskTypes is a JSON array
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	if _, err := mgr.Tasks.Metrics.TaskJournal(ctx, "t1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.Metrics.AgentJournal(ctx, "a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.Metrics.AgentInteractionDurations(ctx, "a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.Usage(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.UsageTypes(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.Logs(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.AgentAction(ctx, "t1", taskautomationapi.AgentActions("Accept"), taskautomationapi.ManualActionRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.SetAgentLock(ctx, taskautomationapi.AgentLocking("unlock")); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tasks.ChangeState(ctx, "t1", taskautomationapi.TaskState("Completed")); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.TestAction(ctx, taskautomationapi.TestAction{}); err != nil {
		t.Fatal(err)
	}
}
