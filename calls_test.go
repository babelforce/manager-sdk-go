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

func callPage(current, pages, n int) string {
	items := strings.Repeat("{},", n)
	if n > 0 {
		items = items[:len(items)-1]
	}
	return `{"items":[` + items + `],` +
		`"pagination":{"pages":` + itoa(pages) + `,"current":` + itoa(current) + `,"total":` + itoa(pages) + `,"max":1}}`
}

func callServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/calls/reporting" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(callPage(2, 2, 1)))
			} else {
				_, _ = w.Write([]byte(callPage(1, 2, 2)))
			}
		case p == "/api/v2/calls/reporting/simple" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(callPage(1, 1, 2)))
		case strings.HasPrefix(p, "/api/v2/calls/reporting/simple/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(callPage(1, 1, 1)))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestReportingListAndSimple(t *testing.T) {
	srv := callServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	calls, err := mgr.Calls.Reporting.ListAll(ctx, managerapi.ListReportingCallsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 {
		t.Fatalf("expected 3 auto-paged calls, got %d", len(calls))
	}

	simple, err := mgr.Calls.Reporting.SimpleAll(ctx, managerapi.ListAllSimpleReportingCallsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(simple) != 2 {
		t.Fatalf("expected 2 simple calls, got %d", len(simple))
	}

	byType, err := mgr.Calls.Reporting.SimpleAllByType(ctx, managerapi.SimpleReportingReportType("inbound"), managerapi.ListSimpleReportingCallsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(byType) != 1 {
		t.Fatalf("expected 1 inbound call, got %d", len(byType))
	}
}

func TestReportingErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	_, err := mgr.Calls.Reporting.ListAll(context.Background(), managerapi.ListReportingCallsParams{})

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
