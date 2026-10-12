package manager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	taskautomationapi "github.com/babelforce/manager-sdk-go/gen/taskautomation"
)

func taskServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v3/tasks/schedules" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"records":[{"name":"s1"}]}`))
		case p == "/api/v3/tasks" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v3/tasks" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(`{"records":[{}],"_metadata":{"page":2,"page_count":2}}`))
			} else {
				_, _ = w.Write([]byte(`{"records":[{},{}],"_metadata":{"page":1,"page_count":2}}`))
			}
		case strings.HasPrefix(p, "/api/v3/tasks/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestTasksCreateListSchedules(t *testing.T) {
	srv := taskServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := mgr.Tasks.Create(context.Background(), taskautomationapi.SubmitTask{}); err != nil {
		t.Fatalf("create: %v", err)
	}

	tasks, err := mgr.Tasks.ListAll(context.Background(), ListTasksQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 auto-paged tasks, got %d", len(tasks))
	}

	sched, err := mgr.Tasks.Schedules.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sched == nil {
		t.Fatal("expected schedules")
	}
}

func TestTaskErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Tasks.Get(context.Background(), "t1")

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
