package manager

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func recordingPage(id string, current, pages int) string {
	return `{"items":[{"id":"` + id + `"}],` +
		`"pagination":{"pages":` + itoa(pages) + `,"current":` + itoa(current) + `,"total":` + itoa(pages) + `,"max":1}}`
}

func recordingItem(id string) string {
	return `{"item":{"id":"` + id + `"},"success":true}`
}

func recordingServer(t *testing.T, bulkBody *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Every request must carry the configured bearer token.
		if r.Header.Get("Authorization") != "Bearer TEST" {
			t.Errorf("missing auth header on %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/recordings" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(recordingPage(uuidB, 2, 2)))
			} else {
				_, _ = w.Write([]byte(recordingPage(uuidA, 1, 2)))
			}
		case strings.HasPrefix(p, "/api/v2/recordings/bulk/") && r.Method == http.MethodPost:
			if bulkBody != nil {
				b, _ := io.ReadAll(r.Body)
				*bulkBody = string(b)
			}
			_, _ = w.Write([]byte(`{"bulk":[{"id":"` + uuidA + `","success":true}]}`))
		case strings.HasSuffix(p, "/flag") && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(recordingItem(uuidA)))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestRecordingsListAllAndFlag(t *testing.T) {
	srv := recordingServer(t, nil)
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	recs, err := mgr.Recordings.ListAll(ctx, managerapi.ListRecordingsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 auto-paged recordings, got %d", len(recs))
	}

	flagged, err := mgr.Recordings.Flag(ctx, uuidA)
	if err != nil {
		t.Fatalf("flag: %v", err)
	}
	if !flagged.Success {
		t.Fatal("expected flag success")
	}
}

func TestRecordingsBulkActionSendsIds(t *testing.T) {
	var body string
	srv := recordingServer(t, &body)
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	resp, err := mgr.Recordings.BulkAction(context.Background(), "delete", []string{uuidA, uuidB})
	if err != nil {
		t.Fatalf("bulk action: %v", err)
	}
	if len(resp.Bulk) != 1 {
		t.Fatalf("expected 1 bulk result, got %d", len(resp.Bulk))
	}

	var got struct {
		Ids []string `json:"ids"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode bulk body %q: %v", body, err)
	}
	if len(got.Ids) != 2 || got.Ids[0] != uuidA || got.Ids[1] != uuidB {
		t.Fatalf("expected ids [%s %s], got %+v", uuidA, uuidB, got.Ids)
	}
}

func TestRecordingsErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Recordings.Get(context.Background(), uuidA)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
