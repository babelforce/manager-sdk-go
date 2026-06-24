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
		case "/api/v2/metrics/push", "/api/v2/metrics/reset",
			"/api/v2/metrics/m1", "/api/v2/metrics/m1/describe":
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestMetricsListGetPushReset(t *testing.T) {
	srv := metricServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
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
	if _, err := mgr.Metrics.Push(ctx); err != nil {
		t.Fatalf("push: %v", err)
	}
	if _, err := mgr.Metrics.Reset(ctx); err != nil {
		t.Fatalf("reset: %v", err)
	}
}

func TestMetricsErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	_, err := mgr.Metrics.Get(context.Background(), "m1")

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
