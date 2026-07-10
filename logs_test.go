package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestLiveLoggingControl(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v2/logs/enable" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"message":"enabled","success":true}`))
		case r.URL.Path == "/api/v2/logs/disable" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"message":"disabled","success":true}`))
		case r.URL.Path == "/api/v2/logs" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"item":{"id":"l1"},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, err := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	enabled, err := mgr.Logs.EnableLive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !enabled.Success || enabled.Message == nil || *enabled.Message != "enabled" {
		t.Fatalf("expected enabled message, got %+v", enabled)
	}

	disabled, err := mgr.Logs.DisableLive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !disabled.Success || disabled.Message == nil || *disabled.Message != "disabled" {
		t.Fatalf("expected disabled message, got %+v", disabled)
	}

	written, err := mgr.Logs.Write(ctx, managerapi.WriteLogRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if (*written.Item)["id"] != "l1" {
		t.Fatalf("expected written log item id l1, got %+v", written)
	}
}
