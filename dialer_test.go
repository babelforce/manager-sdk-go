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

func behaviourPage(id, name string, current, pages int) string {
	return `{"items":[{"id":"` + id + `","name":"` + name + `"}],` +
		`"pagination":{"pages":` + itoa(pages) + `,"current":` + itoa(current) + `,"total":` + itoa(pages) + `,"max":1}}`
}

func dialerServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Every dialer request must carry the bearer token.
		if r.Header.Get("Authorization") != "Bearer TEST" {
			t.Errorf("missing auth headers on %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/dialer" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"queued":3},"success":true}`))
		case p == "/api/v2/outbound/dialer-behaviours" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Aggressive"},"success":true}`))
		case p == "/api/v2/outbound/dialer-behaviours" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(behaviourPage(uuidB, "Conservative", 2, 2)))
			} else {
				_, _ = w.Write([]byte(behaviourPage(uuidA, "Aggressive", 1, 2)))
			}
		case strings.HasPrefix(p, "/api/v2/outbound/dialer-behaviours/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Aggressive"},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestDialerBehavioursCreateListGet(t *testing.T) {
	srv := dialerServer(t)
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	created, err := mgr.Dialer.Behaviours.Create(ctx, managerapi.DialerBehaviourWriteBody{Name: "Aggressive"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Item.Name != "Aggressive" {
		t.Fatalf("expected Aggressive, got %q", created.Item.Name)
	}

	behaviours, err := mgr.Dialer.Behaviours.ListAll(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(behaviours) != 2 || behaviours[0].Name != "Aggressive" || behaviours[1].Name != "Conservative" {
		t.Fatalf("expected 2 auto-paged behaviours, got %+v", behaviours)
	}

	got, err := mgr.Dialer.Behaviours.Get(ctx, uuidA)
	if err != nil {
		t.Fatal(err)
	}
	if got.Item.Name != "Aggressive" {
		t.Fatalf("get: %q", got.Item.Name)
	}
}

func TestDialerInfo(t *testing.T) {
	srv := dialerServer(t)
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	info, err := mgr.Dialer.Info(context.Background())
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if info.Success == nil || !*info.Success {
		t.Fatal("expected success info response")
	}
}

func TestDialerErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Dialer.Behaviours.Get(context.Background(), uuidA)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
