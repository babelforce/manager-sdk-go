package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestConversationsAndSessions(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		case p == "/api/v2/conversations" && m == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"}],"pagination":{"pages":1,"current":1,"total":1,"max":50}}`))
		case p == "/api/v2/conversations" && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case strings.HasSuffix(p, "/events/e1"):
			_, _ = w.Write([]byte(item))
		case strings.HasSuffix(p, "/events"):
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"},{"id":"` + uuidB + `"}],"pagination":{"pages":1,"current":1,"total":2,"max":50}}`))
		case strings.HasSuffix(p, "/session"):
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		case strings.HasPrefix(p, "/api/v2/sessions"):
			_, _ = w.Write([]byte(`{}`))
		case m == http.MethodDelete:
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			_, _ = w.Write([]byte(item))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: APIKey("x", "y")})

	if cs, err := mgr.Conversations.ListAll(ctx, managerapi.ListConversationsParams{}); err != nil || len(cs) != 1 {
		t.Fatalf("conversations list: %v len=%d", err, len(cs))
	}
	if _, err := mgr.Conversations.Create(ctx, managerapi.RestCreateConversation{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Conversations.Get(ctx, "cv1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Conversations.Update(ctx, "cv1", managerapi.RestUpdateConversation{}); err != nil {
		t.Fatal(err)
	}
	if evs, err := mgr.Conversations.Events(ctx, "cv1"); err != nil || len(evs) != 2 {
		t.Fatalf("events: %v len=%d", err, len(evs))
	}
	if _, err := mgr.Conversations.GetEvent(ctx, "cv1", "e1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Conversations.GetSession(ctx, "cv1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Conversations.UpdateSession(ctx, "cv1", managerapi.ConversationSessionVariables{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Conversations.Delete(ctx, "cv1"); err != nil {
		t.Fatal(err)
	}

	if _, err := mgr.Sessions.Create(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Sessions.Get(ctx, "s1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Sessions.UpdateVariables(ctx, "s1", managerapi.UpdateSessionVariablesRequest{}); err != nil {
		t.Fatal(err)
	}
}
