package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Hyphenated-UUID path segments must reach the wire as unhyphenated 32-char hex (parity with the
// Rust path helper); canonical and non-UUID ids pass through untouched.
func TestHyphenatedUUIDPathNormalization(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"item":{},"success":true}`))
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, err := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("t")})
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	for _, id := range []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222222222222222222222222222",
		"not-a-uuid",
	} {
		if _, err := mgr.Calls.Get(ctx, id); err != nil {
			t.Fatalf("Calls.Get(%q): %v", id, err)
		}
	}

	want := []string{
		"/api/v2/calls/11111111111111111111111111111111",
		"/api/v2/calls/22222222222222222222222222222222",
		"/api/v2/calls/not-a-uuid",
	}
	for i, w := range want {
		if paths[i] != w {
			t.Errorf("request %d path = %q, want %q", i, paths[i], w)
		}
	}
}
