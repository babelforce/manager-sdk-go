package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestQueuesAndSelections(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	list := `{"items":[{"id":"` + uuidA + `"},{"id":"` + uuidB + `"}],"pagination":{"pages":1,"current":1,"total":2,"max":50}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		case p == "/api/v2/queues" && m == http.MethodGet:
			_, _ = w.Write([]byte(list))
		case p == "/api/v2/queues" && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case p == "/api/v2/queues/q1/select":
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v2/queues/q1/selections" && m == http.MethodGet:
			_, _ = w.Write([]byte(list))
		case p == "/api/v2/queues/q1/selections" && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case m == http.MethodDelete && (strings.HasSuffix(p, "/agents/a1") || strings.HasSuffix(p, "/groups/g1") || strings.HasSuffix(p, "/tags/t1")):
			_, _ = w.Write([]byte(`{"success":true}`))
		case m == http.MethodPost && (strings.HasSuffix(p, "/agents") || strings.HasSuffix(p, "/groups") || strings.HasSuffix(p, "/tags")):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"success":true}`))
		case p == "/api/v2/queues/q1/selections/sel1":
			_, _ = w.Write([]byte(item))
		case p == "/api/v2/queues/q1" && m == http.MethodDelete:
			_, _ = w.Write([]byte(`{"success":true}`))
		case p == "/api/v2/queues/q1":
			_, _ = w.Write([]byte(item))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: APIKey("x", "y")})

	queues, err := mgr.Queues.ListAll(ctx, managerapi.ListQueuesParams{})
	if err != nil || len(queues) != 2 {
		t.Fatalf("queues list: err=%v len=%d", err, len(queues))
	}
	if _, err := mgr.Queues.Create(ctx, managerapi.RestCreateQueue{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Queues.Get(ctx, "q1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Queues.Update(ctx, "q1", managerapi.RestUpdateQueue{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Queues.Delete(ctx, "q1"); err != nil {
		t.Fatal(err)
	}

	sel := mgr.Queues.Selections
	sels, err := sel.ListAll(ctx, "q1", managerapi.ListQueueSelectionsParams{})
	if err != nil || len(sels) != 2 {
		t.Fatalf("selections list: err=%v len=%d", err, len(sels))
	}
	if _, err := sel.Create(ctx, "q1", managerapi.RestCreateQueueSelection{}); err != nil {
		t.Fatal(err)
	}
	if _, err := sel.Get(ctx, "q1", "sel1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sel.SelectAgents(ctx, "q1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sel.AddAgent(ctx, "q1", "sel1", "a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sel.RemoveAgent(ctx, "q1", "sel1", "a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sel.AddGroup(ctx, "q1", "sel1", "g1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sel.AddTag(ctx, "q1", "sel1", "t1"); err != nil {
		t.Fatal(err)
	}
}
