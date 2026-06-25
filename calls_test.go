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
		case p == "/api/v2/calls/reporting/simple/inbound" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(callPage(1, 1, 3)))
		case p == "/api/v2/calls/c1/cancel" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"item":{},"success":true}`))
		case p == "/api/v2/queues/q1/calls" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(callPage(1, 1, 2)))
		case p == "/api/v2/queues/q1/callback" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"item":{},"message":"queued","success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestReportingListAndSimple(t *testing.T) {
	srv := callServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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
}

func TestCallControlAndQueueing(t *testing.T) {
	srv := callServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	cancelled, err := mgr.Calls.Cancel(ctx, "c1")
	if err != nil {
		t.Fatal(err)
	}
	if !cancelled.Success {
		t.Fatalf("expected cancelled call success, got %+v", cancelled)
	}

	cb, err := mgr.Calls.QueueCallback(ctx, "q1", managerapi.QueueCallbackRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !cb.Success || cb.Message == nil || *cb.Message != "queued" {
		t.Fatalf("expected queued callback, got %+v", cb)
	}

	var queued []managerapi.QueuedCall
	for q, err := range mgr.Calls.ListQueued(ctx, "q1", managerapi.ListQueuedCallsParams{}) {
		if err != nil {
			t.Fatal(err)
		}
		queued = append(queued, q)
	}
	if len(queued) != 2 {
		t.Fatalf("expected 2 queued calls, got %d", len(queued))
	}

	inbound, err := mgr.Calls.Reporting.InboundSimpleAll(ctx, managerapi.ListInboundSimpleReportingCallsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(inbound) != 3 {
		t.Fatalf("expected 3 inbound simple calls, got %d", len(inbound))
	}
}

func TestReportingErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Calls.Reporting.ListAll(context.Background(), managerapi.ListReportingCallsParams{})

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
