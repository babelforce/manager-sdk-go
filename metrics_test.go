package manager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func metricServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v2/metrics/ids":
			_, _ = w.Write([]byte(`{"items":["m1","m2"]}`))
		case "/api/v2/metrics/m1", "/api/v2/metrics/m1/describe":
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestMetricsListGetDescribe(t *testing.T) {
	srv := metricServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	ids, err := mgr.Metrics.ListIds(ctx)
	if err != nil || ids == nil {
		t.Fatalf("listIds: %v", err)
	}
	if _, err := mgr.Metrics.Get(ctx, "m1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if _, err := mgr.Metrics.Describe(ctx, "m1"); err != nil {
		t.Fatalf("describe: %v", err)
	}
}

func TestMetricsDefinitions(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v2/metrics/describe" && r.Method == http.MethodGet {
			got = r.Method + " " + r.URL.Path
			_, _ = w.Write([]byte(`{"items":[{"id":"m1"},{"id":"m2"}],"success":true}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	defs, err := mgr.Metrics.Definitions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "GET /api/v2/metrics/describe" {
		t.Fatalf("Definitions hit %q, want GET /api/v2/metrics/describe", got)
	}
	if len(defs.Items) != 2 {
		t.Fatalf("Definitions = %+v", defs)
	}
}

func TestMetricsErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Metrics.Get(context.Background(), "m1")

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
